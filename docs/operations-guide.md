# CliForge Operations Guide

**Version**: 0.9.0
**Last Updated**: 2025-11-25
**Target Audience**: DevOps Engineers, SREs, System Administrators

Comprehensive guide for deploying, managing, and maintaining CliForge-generated CLIs in production environments.

---

## Table of Contents

1. [Production Deployment](#production-deployment)
2. [Monitoring and Observability](#monitoring-and-observability)
3. [Update Management](#update-management)
4. [Enterprise Configuration](#enterprise-configuration)
5. [Troubleshooting in Production](#troubleshooting-in-production)
6. [Security Hardening](#security-hardening)
7. [Disaster Recovery](#disaster-recovery)
8. [Performance Optimization](#performance-optimization)

---

## Production Deployment

### Overview

CliForge CLIs are distributed as static, self-contained binaries with embedded configuration. This section covers production deployment strategies for various environments.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    PRODUCTION ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌────────────────────────────────────────────────────┐         │
│  │              Release Infrastructure                 │         │
│  ├────────────────────────────────────────────────────┤         │
│  │                                                     │         │
│  │  CDN/Object Storage (Binaries)                     │         │
│  │  ├── /latest/version.json                          │         │
│  │  ├── /v1.2.3/mycli-darwin-amd64                   │         │
│  │  ├── /v1.2.3/mycli-linux-amd64                    │         │
│  │  ├── /v1.2.3/checksums.txt                        │         │
│  │  └── /v1.2.3/signatures/                          │         │
│  │                                                     │         │
│  └─────────────────────────────────────────────────────┘         │
│                           │                                      │
│  ┌────────────────────────▼───────────────────────────┐         │
│  │           Distribution Channels                     │         │
│  ├────────────────────────────────────────────────────┤         │
│  │                                                     │         │
│  │  ┌────────────┐  ┌────────────┐  ┌─────────────┐  │         │
│  │  │  Homebrew  │  │  apt/yum   │  │   Docker    │  │         │
│  │  │   Tap      │  │Repository  │  │   Registry  │  │         │
│  │  └────────────┘  └────────────┘  └─────────────┘  │         │
│  │                                                     │         │
│  └─────────────────────────────────────────────────────┘         │
│                           │                                      │
│  ┌────────────────────────▼───────────────────────────┐         │
│  │              End User Systems                       │         │
│  ├────────────────────────────────────────────────────┤         │
│  │                                                     │         │
│  │  Developer Workstations → macOS, Linux, Windows    │         │
│  │  CI/CD Systems → Jenkins, GitLab, GitHub Actions   │         │
│  │  Production Servers → Linux servers, containers    │         │
│  │                                                     │         │
│  └─────────────────────────────────────────────────────┘         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

### 1.1 Binary Distribution Strategies

#### Direct Binary Distribution

**Recommended for**: Simple deployments, quick rollouts

**Setup**:

```bash
# Create release structure
mkdir -p releases/{latest,v1.2.3}

# Build multi-platform binaries
GOOS=darwin GOARCH=amd64 go build -o mycli-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o mycli-darwin-arm64
GOOS=linux GOARCH=amd64 go build -o mycli-linux-amd64
GOOS=windows GOARCH=amd64 go build -o mycli-windows-amd64.exe

# Generate checksums
sha256sum mycli-* > checksums.txt

# Sign binaries (optional but recommended)
gpg --armor --detach-sign mycli-linux-amd64
```

**Release Metadata** (`version.json`):

```json
{
  "version": "1.2.3",
  "release_date": "2025-11-25T10:00:00Z",
  "platforms": {
    "darwin-amd64": {
      "url": "https://releases.example.com/v1.2.3/mycli-darwin-amd64",
      "checksum": "abc123...",
      "checksum_algo": "sha256",
      "size": 15728640
    },
    "darwin-arm64": {
      "url": "https://releases.example.com/v1.2.3/mycli-darwin-arm64",
      "checksum": "def456...",
      "checksum_algo": "sha256",
      "size": 14680064
    },
    "linux-amd64": {
      "url": "https://releases.example.com/v1.2.3/mycli-linux-amd64",
      "checksum": "789xyz...",
      "checksum_algo": "sha256",
      "size": 16777216
    }
  },
  "changelog": "## v1.2.3\n- Security fix: Updated dependencies\n- Feature: Added retry logic",
  "critical": false,
  "min_version": "1.0.0"
}
```

**Installation Script** (`install.sh`):

```bash
#!/bin/bash
# Production-grade installation script

set -euo pipefail

# Configuration
GITHUB_REPO="myorg/mycli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="mycli"

# Detect platform
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

PLATFORM="${OS}-${ARCH}"

# Fetch latest version
VERSION=$(curl -sSL "https://releases.example.com/latest/version.json" | jq -r '.version')
DOWNLOAD_URL=$(curl -sSL "https://releases.example.com/latest/version.json" | \
    jq -r ".platforms[\"${PLATFORM}\"].url")
CHECKSUM=$(curl -sSL "https://releases.example.com/latest/version.json" | \
    jq -r ".platforms[\"${PLATFORM}\"].checksum")

echo "Installing ${BINARY_NAME} v${VERSION} for ${PLATFORM}..."

# Download binary
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

curl -sSL "${DOWNLOAD_URL}" -o "${TMP_DIR}/${BINARY_NAME}"

# Verify checksum
echo "${CHECKSUM}  ${TMP_DIR}/${BINARY_NAME}" | sha256sum -c -

# Install
chmod +x "${TMP_DIR}/${BINARY_NAME}"
sudo mv "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

echo "✓ ${BINARY_NAME} v${VERSION} installed successfully!"
echo "Run '${BINARY_NAME} --version' to verify."
```

---

### 1.2 Package Manager Distribution

#### Homebrew (macOS/Linux)

**Create Homebrew Formula** (`mycli.rb`):

```ruby
class Mycli < Formula
  desc "My API Command Line Interface"
  homepage "https://github.com/myorg/mycli"
  version "1.2.3"

  on_macos do
    if Hardware::CPU.arm?
      url "https://releases.example.com/v1.2.3/mycli-darwin-arm64"
      sha256 "def456..."
    else
      url "https://releases.example.com/v1.2.3/mycli-darwin-amd64"
      sha256 "abc123..."
    end
  end

  on_linux do
    url "https://releases.example.com/v1.2.3/mycli-linux-amd64"
    sha256 "789xyz..."
  end

  def install
    bin.install Dir["mycli*"].first => "mycli"

    # Install shell completions
    generate_completions_from_executable(bin/"mycli", "completion")

    # Install man pages (if available)
    # man1.install "mycli.1"
  end

  test do
    assert_match "mycli version 1.2.3", shell_output("#{bin}/mycli --version")
  end
end
```

**Setup Homebrew Tap**:

```bash
# Create tap repository
mkdir -p homebrew-tap/Formula
cp mycli.rb homebrew-tap/Formula/

# Publish to GitHub
cd homebrew-tap
git init
git add Formula/mycli.rb
git commit -m "Add mycli formula v1.2.3"
git remote add origin https://github.com/myorg/homebrew-tap.git
git push -u origin main

# Users install with:
# brew tap myorg/tap
# brew install mycli
```

**Automated Formula Updates**:

```bash
#!/bin/bash
# update-homebrew.sh - Called from CI/CD

VERSION=$1
DARWIN_AMD64_SHA=$(sha256sum mycli-darwin-amd64 | cut -d' ' -f1)
DARWIN_ARM64_SHA=$(sha256sum mycli-darwin-arm64 | cut -d' ' -f1)
LINUX_AMD64_SHA=$(sha256sum mycli-linux-amd64 | cut -d' ' -f1)

# Update formula
sed -i "s/version \".*\"/version \"${VERSION}\"/" Formula/mycli.rb
sed -i "s/darwin-amd64.*sha256.*/darwin-amd64\"\n      sha256 \"${DARWIN_AMD64_SHA}\"/" Formula/mycli.rb
sed -i "s/darwin-arm64.*sha256.*/darwin-arm64\"\n      sha256 \"${DARWIN_ARM64_SHA}\"/" Formula/mycli.rb
sed -i "s/linux-amd64.*sha256.*/linux-amd64\"\n      sha256 \"${LINUX_AMD64_SHA}\"/" Formula/mycli.rb

# Commit and push
git add Formula/mycli.rb
git commit -m "Update mycli to v${VERSION}"
git push
```

#### APT Repository (Debian/Ubuntu)

**Create DEB Package**:

```bash
# Directory structure
mkdir -p mycli-1.2.3/DEBIAN
mkdir -p mycli-1.2.3/usr/local/bin
mkdir -p mycli-1.2.3/usr/share/man/man1
mkdir -p mycli-1.2.3/etc/bash_completion.d

# Copy binary
cp mycli-linux-amd64 mycli-1.2.3/usr/local/bin/mycli
chmod +x mycli-1.2.3/usr/local/bin/mycli

# Create control file
cat > mycli-1.2.3/DEBIAN/control <<EOF
Package: mycli
Version: 1.2.3
Section: utils
Priority: optional
Architecture: amd64
Maintainer: DevOps <devops@example.com>
Description: My API Command Line Interface
 CliForge-generated CLI for My API.
 Provides command-line access to all API endpoints.
Depends: ca-certificates
EOF

# Create postinstall script
cat > mycli-1.2.3/DEBIAN/postinst <<'EOF'
#!/bin/bash
set -e

# Generate shell completions
if [ -x /usr/local/bin/mycli ]; then
    /usr/local/bin/mycli completion bash > /etc/bash_completion.d/mycli 2>/dev/null || true
fi

exit 0
EOF
chmod +x mycli-1.2.3/DEBIAN/postinst

# Build package
dpkg-deb --build mycli-1.2.3

# Sign package
dpkg-sig --sign builder mycli-1.2.3.deb
```

**Setup APT Repository**:

```bash
# Create repository structure
mkdir -p apt-repo/{pool,dists/stable/main/binary-amd64}

# Copy packages
cp mycli-1.2.3.deb apt-repo/pool/

# Generate Packages file
cd apt-repo
dpkg-scanpackages pool /dev/null | gzip -9c > dists/stable/main/binary-amd64/Packages.gz

# Generate Release file
cat > dists/stable/Release <<EOF
Origin: MyOrg
Label: MyOrg APT Repository
Suite: stable
Codename: stable
Architectures: amd64 arm64
Components: main
Description: MyOrg package repository
EOF

# Sign Release file
gpg --clearsign -o dists/stable/InRelease dists/stable/Release

# Serve via nginx or upload to S3
# Users add with:
# echo "deb https://apt.example.com/ stable main" | sudo tee /etc/apt/sources.list.d/mycli.list
# curl -fsSL https://apt.example.com/gpg.key | sudo apt-key add -
# sudo apt update && sudo apt install mycli
```

#### YUM Repository (RHEL/CentOS)

**Create RPM Package**:

```bash
# Create RPM build structure
mkdir -p rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Create spec file
cat > rpmbuild/SPECS/mycli.spec <<EOF
Name:           mycli
Version:        1.2.3
Release:        1%{?dist}
Summary:        My API Command Line Interface
License:        MIT
URL:            https://github.com/myorg/mycli
Source0:        mycli-1.2.3-linux-amd64

BuildArch:      x86_64
Requires:       ca-certificates

%description
CliForge-generated CLI for My API.
Provides command-line access to all API endpoints.

%install
mkdir -p %{buildroot}/usr/local/bin
cp %{SOURCE0} %{buildroot}/usr/local/bin/mycli
chmod +x %{buildroot}/usr/local/bin/mycli

%files
/usr/local/bin/mycli

%post
# Generate shell completions
/usr/local/bin/mycli completion bash > /etc/bash_completion.d/mycli 2>/dev/null || true

%changelog
* Mon Nov 25 2025 DevOps <devops@example.com> - 1.2.3-1
- Security fix: Updated dependencies
- Feature: Added retry logic
EOF

# Build RPM
rpmbuild -bb rpmbuild/SPECS/mycli.spec

# Sign RPM
rpm --addsign rpmbuild/RPMS/x86_64/mycli-1.2.3-1.x86_64.rpm
```

**Setup YUM Repository**:

```bash
# Create repository
mkdir -p yum-repo/stable/x86_64

# Copy RPMs
cp rpmbuild/RPMS/x86_64/*.rpm yum-repo/stable/x86_64/

# Create repository metadata
createrepo yum-repo/stable/x86_64/

# Sign repository metadata
gpg --detach-sign --armor yum-repo/stable/x86_64/repodata/repomd.xml

# Users add with:
# cat > /etc/yum.repos.d/mycli.repo <<EOF
# [mycli]
# name=MyOrg Repository
# baseurl=https://yum.example.com/stable/x86_64
# enabled=1
# gpgcheck=1
# gpgkey=https://yum.example.com/gpg.key
# EOF
# sudo yum install mycli
```

---

### 1.3 Docker Deployment

#### Docker Image

**Dockerfile**:

```dockerfile
# Multi-stage build for minimal image
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.2.3 -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o mycli \
    ./cmd/mycli

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy binary
COPY --from=builder /build/mycli /usr/local/bin/mycli

# Create non-root user
USER nobody:nobody

# Runtime
ENTRYPOINT ["/usr/local/bin/mycli"]
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="mycli" \
      org.opencontainers.image.description="My API Command Line Interface" \
      org.opencontainers.image.version="1.2.3" \
      org.opencontainers.image.vendor="MyOrg"
```

**Build and Push**:

```bash
# Build multi-platform images
docker buildx create --use --name multiarch
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag myorg/mycli:1.2.3 \
  --tag myorg/mycli:latest \
  --push \
  .

# Verify
docker run --rm myorg/mycli:1.2.3 --version
```

**Docker Compose** (for development/testing):

```yaml
# docker-compose.yml
version: '3.8'

services:
  mycli:
    image: myorg/mycli:1.2.3
    environment:
      - MYCLI_API_KEY=${MYCLI_API_KEY}
      - MYCLI_OUTPUT_FORMAT=json
    volumes:
      # Persist configuration and cache
      - mycli-config:/home/nobody/.config/mycli
      - mycli-cache:/home/nobody/.cache/mycli
    command: ["users", "list"]
    networks:
      - mycli-net

volumes:
  mycli-config:
  mycli-cache:

networks:
  mycli-net:
```

**Kubernetes Deployment**:

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mycli-cronjob
  namespace: automation
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mycli
  template:
    metadata:
      labels:
        app: mycli
    spec:
      containers:
      - name: mycli
        image: myorg/mycli:1.2.3
        imagePullPolicy: IfNotPresent
        env:
        - name: MYCLI_API_KEY
          valueFrom:
            secretKeyRef:
              name: mycli-secrets
              key: api-key
        - name: MYCLI_OUTPUT_FORMAT
          value: "json"
        - name: MYCLI_NO_COLOR
          value: "1"
        volumeMounts:
        - name: config
          mountPath: /home/nobody/.config/mycli
          readOnly: true
        - name: cache
          mountPath: /home/nobody/.cache/mycli
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: mycli-config
      - name: cache
        emptyDir: {}

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: mycli-config
  namespace: automation
data:
  config.yaml: |
    defaults:
      output:
        format: json
        color: never
      http:
        timeout: 30s

---
apiVersion: v1
kind: Secret
metadata:
  name: mycli-secrets
  namespace: automation
type: Opaque
stringData:
  api-key: "sk-your-secret-key-here"

---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: mycli-sync
  namespace: automation
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: mycli
            image: myorg/mycli:1.2.3
            args: ["sync", "--all"]
            env:
            - name: MYCLI_API_KEY
              valueFrom:
                secretKeyRef:
                  name: mycli-secrets
                  key: api-key
          restartPolicy: OnFailure
```

---

### 1.4 Air-Gapped Installations

#### Overview

Air-gapped environments have no internet access. This requires bundling all dependencies.

**Preparation**:

```bash
# Create air-gap bundle
mkdir -p airgap-bundle/{bin,certs,config,docs}

# Copy binary
cp mycli-linux-amd64 airgap-bundle/bin/mycli

# Copy CA certificates (required for HTTPS)
cp /etc/ssl/certs/ca-certificates.crt airgap-bundle/certs/

# Copy OpenAPI spec (for offline use)
curl https://api.example.com/openapi.yaml -o airgap-bundle/config/openapi.yaml

# Copy default configuration
cat > airgap-bundle/config/config.yaml <<EOF
metadata:
  name: mycli
  version: 1.2.3

api:
  # Use local spec instead of remote URL
  openapi_url: file:///opt/mycli/config/openapi.yaml
  base_url: https://api.internal.example.com

behaviors:
  caching:
    enabled: true
    ttl: 24h  # Longer cache for air-gapped

updates:
  enabled: false  # Disable auto-updates
EOF

# Copy documentation
cp docs/*.md airgap-bundle/docs/

# Create installation script
cat > airgap-bundle/install.sh <<'EOF'
#!/bin/bash
set -euo pipefail

INSTALL_PREFIX="${INSTALL_PREFIX:-/opt/mycli}"

echo "Installing mycli in air-gapped mode..."

# Create directories
sudo mkdir -p "$INSTALL_PREFIX"/{bin,config,certs}

# Install binary
sudo cp bin/mycli "$INSTALL_PREFIX/bin/"
sudo chmod +x "$INSTALL_PREFIX/bin/mycli"

# Install configuration
sudo cp config/* "$INSTALL_PREFIX/config/"

# Install CA certificates
sudo cp certs/ca-certificates.crt "$INSTALL_PREFIX/certs/"

# Set environment variable for CA bundle
echo "export SSL_CERT_FILE=$INSTALL_PREFIX/certs/ca-certificates.crt" | \
  sudo tee -a /etc/profile.d/mycli.sh

# Add to PATH
echo "export PATH=\$PATH:$INSTALL_PREFIX/bin" | \
  sudo tee -a /etc/profile.d/mycli.sh

echo "✓ Installation complete!"
echo "Run: source /etc/profile.d/mycli.sh"
echo "Then: mycli --version"
EOF
chmod +x airgap-bundle/install.sh

# Create tarball
tar czf mycli-1.2.3-airgap.tar.gz airgap-bundle/

# Generate checksum
sha256sum mycli-1.2.3-airgap.tar.gz > mycli-1.2.3-airgap.tar.gz.sha256
```

**Installation in Air-Gapped Environment**:

```bash
# Transfer tarball to air-gapped system (USB, etc.)

# Verify checksum
sha256sum -c mycli-1.2.3-airgap.tar.gz.sha256

# Extract
tar xzf mycli-1.2.3-airgap.tar.gz

# Install
cd airgap-bundle
sudo ./install.sh

# Configure for internal API
cat > ~/.config/mycli/config.yaml <<EOF
api:
  base_url: https://api.internal.corp
  openapi_url: file:///opt/mycli/config/openapi.yaml

defaults:
  http:
    timeout: 60s
    # Use internal CA bundle
    tls_ca_file: /etc/pki/tls/certs/internal-ca.crt
EOF
```

**Update Management in Air-Gapped**:

```bash
# Create update bundle
mkdir mycli-update-1.2.4
cp mycli-linux-amd64 mycli-update-1.2.4/mycli
cp openapi.yaml mycli-update-1.2.4/
tar czf mycli-update-1.2.4.tar.gz mycli-update-1.2.4/

# On air-gapped system
tar xzf mycli-update-1.2.4.tar.gz
sudo cp mycli-update-1.2.4/mycli /opt/mycli/bin/
sudo cp mycli-update-1.2.4/openapi.yaml /opt/mycli/config/
```

---

## Monitoring and Observability

### Overview

Production CLIs require monitoring to track usage, detect failures, and optimize performance.

### 2.1 Logging Configuration

#### Log Levels

CliForge supports standard log levels:

- **ERROR**: Critical failures requiring immediate attention
- **WARN**: Non-critical issues, degraded functionality
- **INFO**: Normal operational messages
- **DEBUG**: Detailed diagnostic information

**Configuration**:

```yaml
# config.yaml
defaults:
  logging:
    level: info  # error, warn, info, debug
    format: json  # json, text
    output: stderr  # stdout, stderr, file
    file: /var/log/mycli/mycli.log
    max_size: 100  # MB
    max_backups: 5
    max_age: 30  # days
    compress: true
```

**Structured Logging Output** (JSON format):

```json
{
  "timestamp": "2025-11-25T14:30:15Z",
  "level": "info",
  "message": "Command executed successfully",
  "fields": {
    "command": "users list",
    "duration_ms": 245,
    "user": "alice",
    "client_version": "1.2.3",
    "api_version": "2.1.0",
    "request_id": "req-abc123",
    "status_code": 200
  }
}
```

#### Centralized Logging

**Syslog Integration**:

```yaml
# config.yaml
defaults:
  logging:
    syslog:
      enabled: true
      network: udp
      address: syslog.example.com:514
      facility: user
      tag: mycli
```

**Fluentd/Fluent Bit**:

```yaml
# fluent-bit.conf
[INPUT]
    Name              tail
    Path              /var/log/mycli/*.log
    Parser            json
    Tag               mycli.*
    Refresh_Interval  5

[FILTER]
    Name    modify
    Match   mycli.*
    Add     service mycli
    Add     environment production

[OUTPUT]
    Name  es
    Match mycli.*
    Host  elasticsearch.example.com
    Port  9200
    Index mycli-logs
```

**CloudWatch Integration** (AWS):

```bash
# Install CloudWatch agent
sudo yum install amazon-cloudwatch-agent

# Configure
cat > /opt/aws/amazon-cloudwatch-agent/etc/config.json <<EOF
{
  "logs": {
    "logs_collected": {
      "files": {
        "collect_list": [
          {
            "file_path": "/var/log/mycli/mycli.log",
            "log_group_name": "/aws/mycli/production",
            "log_stream_name": "{instance_id}",
            "timestamp_format": "%Y-%m-%dT%H:%M:%S"
          }
        ]
      }
    }
  }
}
EOF

# Start agent
sudo /opt/aws/amazon-cloudwatch-agent/bin/amazon-cloudwatch-agent-ctl \
  -a fetch-config \
  -m ec2 \
  -s \
  -c file:/opt/aws/amazon-cloudwatch-agent/etc/config.json
```

---

### 2.2 Metrics and Telemetry

#### Built-in Metrics

CliForge CLIs can emit telemetry data:

**Configuration**:

```yaml
# config.yaml
api:
  telemetry_url: https://telemetry.example.com/v1/metrics

behaviors:
  telemetry:
    enabled: true
    sample_rate: 1.0  # 100% sampling
    batch_size: 100
    flush_interval: 30s
    include_user_id: false  # Privacy consideration
```

**Metrics Emitted**:

```json
{
  "timestamp": "2025-11-25T14:30:15Z",
  "metrics": [
    {
      "name": "command.duration",
      "type": "histogram",
      "value": 245,
      "unit": "ms",
      "tags": {
        "command": "users.list",
        "status": "success",
        "client_version": "1.2.3"
      }
    },
    {
      "name": "api.request.count",
      "type": "counter",
      "value": 1,
      "tags": {
        "endpoint": "/v1/users",
        "method": "GET",
        "status_code": "200"
      }
    },
    {
      "name": "cache.hit_rate",
      "type": "gauge",
      "value": 0.87,
      "tags": {
        "cache_type": "spec"
      }
    }
  ],
  "client": {
    "version": "1.2.3",
    "platform": "linux-amd64",
    "go_version": "1.21.0"
  }
}
```

#### Prometheus Integration

**Metrics Exporter**:

```bash
# Create metrics exporter service
cat > /etc/systemd/system/mycli-exporter.service <<EOF
[Unit]
Description=MyCLI Metrics Exporter
After=network.target

[Service]
Type=simple
User=mycli
ExecStart=/usr/local/bin/mycli-exporter \
  --log-file /var/log/mycli/mycli.log \
  --metrics-port 9090
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now mycli-exporter
```

**Prometheus Configuration**:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'mycli'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s
    metrics_path: /metrics
```

**Example Prometheus Metrics**:

```prometheus
# HELP mycli_command_duration_seconds Command execution duration
# TYPE mycli_command_duration_seconds histogram
mycli_command_duration_seconds_bucket{command="users.list",le="0.1"} 150
mycli_command_duration_seconds_bucket{command="users.list",le="0.5"} 890
mycli_command_duration_seconds_bucket{command="users.list",le="1"} 950
mycli_command_duration_seconds_sum{command="users.list"} 245.3
mycli_command_duration_seconds_count{command="users.list"} 1000

# HELP mycli_api_requests_total Total API requests
# TYPE mycli_api_requests_total counter
mycli_api_requests_total{endpoint="/v1/users",method="GET",status="200"} 1234

# HELP mycli_cache_hit_ratio Cache hit ratio
# TYPE mycli_cache_hit_ratio gauge
mycli_cache_hit_ratio{cache="spec"} 0.87
```

---

### 2.3 Error Tracking

#### Sentry Integration

**Configuration**:

```yaml
# config.yaml
behaviors:
  error_tracking:
    enabled: true
    sentry:
      dsn: https://abc123@sentry.io/456789
      environment: production
      release: mycli@1.2.3
      sample_rate: 1.0
      traces_sample_rate: 0.1
      attach_stacktrace: true
      send_default_pii: false
```

**Error Report Structure**:

```json
{
  "event_id": "abc123def456",
  "timestamp": "2025-11-25T14:30:15Z",
  "level": "error",
  "message": "API request failed: 500 Internal Server Error",
  "exception": {
    "type": "APIError",
    "value": "500 Internal Server Error",
    "stacktrace": [...]
  },
  "contexts": {
    "runtime": {
      "name": "go",
      "version": "1.21.0"
    },
    "app": {
      "app_version": "1.2.3",
      "app_name": "mycli"
    },
    "os": {
      "name": "linux",
      "version": "5.15.0",
      "kernel_version": "5.15.0-generic"
    }
  },
  "tags": {
    "command": "users.list",
    "endpoint": "/v1/users",
    "http.status_code": "500"
  },
  "extra": {
    "request_id": "req-abc123",
    "user_agent": "mycli/1.2.3"
  }
}
```

---

### 2.4 Performance Monitoring

#### Application Performance Monitoring (APM)

**Datadog APM**:

```yaml
# config.yaml
behaviors:
  apm:
    enabled: true
    datadog:
      agent_host: localhost
      agent_port: 8126
      service_name: mycli
      env: production
      version: 1.2.3
```

**New Relic**:

```yaml
behaviors:
  apm:
    enabled: true
    newrelic:
      license_key: ${NEW_RELIC_LICENSE_KEY}
      app_name: mycli
      labels: "environment:production;team:platform"
```

#### Custom Performance Tracking

**Command Execution Tracking**:

```bash
# Enable performance tracking
export MYCLI_PERF_TRACKING=1

# Run command with profiling
mycli --profile users list

# Output
Command: users list
┌─────────────────────────────────────┬──────────┐
│ Phase                               │ Duration │
├─────────────────────────────────────┼──────────┤
│ Config Load                         │ 12ms     │
│ Spec Fetch (cache hit)              │ 1ms      │
│ Command Parse                       │ 3ms      │
│ Authentication                      │ 45ms     │
│ API Request                         │ 178ms    │
│ Response Parse                      │ 6ms      │
│ Output Format                       │ 2ms      │
├─────────────────────────────────────┼──────────┤
│ Total                               │ 247ms    │
└─────────────────────────────────────┴──────────┘
```

---

## Update Management

### Overview

CliForge CLIs support self-updating to deliver security patches and new features without requiring manual reinstallation.

### 3.1 Self-Update Strategies

#### Automatic Updates

**Configuration**:

```yaml
# config.yaml
updates:
  enabled: true
  update_url: https://releases.example.com/latest/version.json
  check_interval: 24h
  auto_update: false  # Require confirmation
  allow_prerelease: false
  update_on_start: true  # Check on every invocation
```

**Update Check Flow**:

```
┌────────────────────────────────────────────────────────────┐
│                    UPDATE CHECK FLOW                       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. User runs: mycli users list                           │
│                                                            │
│  2. Check last update check timestamp                     │
│     └─→ If < 24h ago → Skip check                         │
│     └─→ If ≥ 24h ago → Continue                           │
│                                                            │
│  3. Fetch version.json from update server                 │
│     └─→ Compare versions (current vs latest)              │
│                                                            │
│  4. If newer version available:                           │
│     ┌─────────────────────────────────────────────┐       │
│     │ ⓘ Update available: 1.2.3 → 1.2.4         │       │
│     │   Security fix: CVE-2025-1234              │       │
│     │                                             │       │
│     │   Run 'mycli update' to install            │       │
│     │   Run 'mycli update --skip-version' to     │       │
│     │   suppress this notification                │       │
│     └─────────────────────────────────────────────┘       │
│                                                            │
│  5. Execute command                                        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

**Update Installation**:

```bash
# Manual update
$ mycli update

Checking for updates...
✓ New version available: 1.2.3 → 1.2.4

Changelog:
- Security fix: Updated OpenSSL dependencies (CVE-2025-1234)
- Feature: Added pagination support for large datasets

Download size: 15.2 MB

Do you want to update? [Y/n]: y

Downloading mycli-1.2.4-linux-amd64... ████████████ 100%
Verifying checksum... ✓
Verifying signature... ✓
Installing update... ✓

✓ Update complete! Restart mycli to use version 1.2.4.

# Automatic (non-interactive)
$ mycli update --yes

# Skip specific version
$ mycli update --skip-version 1.2.4
```

#### Controlled Rollouts (Canary/Staged)

**Configuration**:

```yaml
# config.yaml
updates:
  rollout:
    strategy: staged  # immediate, canary, staged
    stages:
      - percentage: 10
        duration: 24h
      - percentage: 50
        duration: 48h
      - percentage: 100
```

**Server-Side Rollout Control** (`version.json`):

```json
{
  "version": "1.2.4",
  "rollout": {
    "strategy": "staged",
    "current_stage": 2,
    "stages": [
      {
        "percentage": 10,
        "started_at": "2025-11-24T00:00:00Z",
        "ends_at": "2025-11-25T00:00:00Z"
      },
      {
        "percentage": 50,
        "started_at": "2025-11-25T00:00:00Z",
        "ends_at": "2025-11-27T00:00:00Z"
      },
      {
        "percentage": 100,
        "started_at": "2025-11-27T00:00:00Z"
      }
    ]
  },
  "platforms": {...}
}
```

**Client Selection Logic**:

```go
// Pseudo-code for staged rollout
func shouldReceiveUpdate(clientID string, rollout Rollout) bool {
    // Hash client ID to get consistent percentage
    hash := sha256.Sum256([]byte(clientID))
    percentile := int(hash[0]) % 100

    currentStage := rollout.Stages[rollout.CurrentStage]
    return percentile < currentStage.Percentage
}
```

---

### 3.2 Version Rollback

**Rollback to Previous Version**:

```bash
# List available versions
$ mycli update --list
Available versions:
  1.2.4 (current)
  1.2.3
  1.2.2
  1.2.1

# Rollback to specific version
$ mycli update --version 1.2.3

⚠️  Downgrading to older version: 1.2.4 → 1.2.3

Are you sure? [y/N]: y

Downloading mycli-1.2.3-linux-amd64... ████████████ 100%
Verifying checksum... ✓
Installing version 1.2.3... ✓

✓ Rollback complete! Version 1.2.3 is now active.
```

**Automatic Rollback on Failure**:

```bash
# Update with automatic rollback
$ mycli update --auto-rollback

Downloading update... ✓
Installing update... ✓
Validating update...

Running health checks:
  ✓ Binary executable
  ✓ Configuration valid
  ✓ API connectivity
  ✗ Authentication failed

✗ Health checks failed. Rolling back to 1.2.3...
✓ Rollback successful.

Error: Update failed health validation.
Please report this issue.
```

---

### 3.3 Testing Updates Before Deployment

#### Staging Environment

**Configuration**:

```yaml
# staging-config.yaml
metadata:
  name: mycli
  version: 1.2.4-staging

api:
  base_url: https://api.staging.example.com
  openapi_url: https://api.staging.example.com/openapi.yaml

updates:
  enabled: true
  update_url: https://releases-staging.example.com/latest/version.json
  channel: staging
```

**Multi-Channel Updates**:

```bash
# Production channel (default)
$ mycli update

# Beta channel
$ mycli update --channel beta

# Development channel
$ mycli update --channel dev
```

**Version Channels** (`version.json`):

```json
{
  "stable": {
    "version": "1.2.3",
    "url": "https://releases.example.com/v1.2.3/mycli-linux-amd64"
  },
  "beta": {
    "version": "1.2.4-beta.1",
    "url": "https://releases.example.com/v1.2.4-beta.1/mycli-linux-amd64"
  },
  "dev": {
    "version": "1.3.0-dev.42",
    "url": "https://releases.example.com/v1.3.0-dev.42/mycli-linux-amd64"
  }
}
```

#### Integration Testing

**Automated Update Testing**:

```bash
#!/bin/bash
# test-update.sh - CI/CD integration test

set -euo pipefail

# Install current version
wget https://releases.example.com/v1.2.3/mycli-linux-amd64 -O mycli
chmod +x mycli

# Verify current version
VERSION=$(./mycli --version | grep -oP 'v\K[0-9.]+')
if [ "$VERSION" != "1.2.3" ]; then
    echo "Version mismatch: expected 1.2.3, got $VERSION"
    exit 1
fi

# Perform update
./mycli update --yes --channel staging

# Verify new version
NEW_VERSION=$(./mycli --version | grep -oP 'v\K[0-9.]+')
if [ "$NEW_VERSION" != "1.2.4" ]; then
    echo "Update failed: expected 1.2.4, got $NEW_VERSION"
    exit 1
fi

# Run smoke tests
./mycli config show
./mycli users list --limit 1

echo "✓ Update test passed!"
```

---

## Enterprise Configuration

### 4.1 Centralized Configuration Management

#### Configuration Hierarchy

```
┌────────────────────────────────────────────────────────────┐
│              CONFIGURATION HIERARCHY                       │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  Priority (highest to lowest):                            │
│                                                            │
│  1. Environment Variables                                 │
│     └─ MYCLI_OUTPUT_FORMAT=json                           │
│                                                            │
│  2. CLI Flags                                             │
│     └─ mycli users list --output yaml                     │
│                                                            │
│  3. User Config (~/.config/mycli/config.yaml)             │
│     └─ Per-user customizations                            │
│                                                            │
│  4. System Config (/etc/mycli/config.yaml)                │
│     └─ Organization-wide defaults                         │
│                                                            │
│  5. Embedded Config (in binary)                           │
│     └─ Vendor defaults                                    │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

#### System-Wide Configuration

**Setup**:

```bash
# Create system config directory
sudo mkdir -p /etc/mycli

# Create organization-wide config
sudo tee /etc/mycli/config.yaml <<EOF
# Organization-wide MyCLI configuration

api:
  base_url: https://api.corp.example.com
  openapi_url: https://api.corp.example.com/openapi.yaml

defaults:
  output:
    format: json
    color: auto

  http:
    timeout: 60s
    # Corporate proxy
    proxy: http://proxy.corp.example.com:8080
    # Corporate CA bundle
    tls_ca_file: /etc/pki/tls/certs/corp-ca-bundle.crt

  caching:
    enabled: true
    ttl: 1h

behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: corp-mycli
      token_url: https://sso.corp.example.com/oauth/token
      auth_url: https://sso.corp.example.com/oauth/authorize
      flow: authorization_code
      scopes:
        - mycli:read
        - mycli:write
      # Use system keyring for token storage
      storage:
        type: keyring
        keyring:
          service: mycli-corp

updates:
  # Disable self-updates in corporate environment
  enabled: false

features:
  # Disable telemetry for privacy
  telemetry: false
EOF

# Set restrictive permissions
sudo chmod 644 /etc/mycli/config.yaml
```

**Group Policy (Windows)**:

```powershell
# Create GPO for MyCLI configuration
New-GPO -Name "MyCLI Configuration" -Comment "Corporate MyCLI settings"

# Set registry keys for configuration
$regPath = "HKLM:\SOFTWARE\Policies\MyCLI"
New-Item -Path $regPath -Force
Set-ItemProperty -Path $regPath -Name "ConfigFile" -Value "C:\ProgramData\MyCLI\config.yaml"
Set-ItemProperty -Path $regPath -Name "ProxyURL" -Value "http://proxy.corp.com:8080"
Set-ItemProperty -Path $regPath -Name "DisableUpdates" -Value 1
```

---

### 4.2 Proxy and Firewall Setup

#### HTTP/HTTPS Proxy Configuration

**Configuration**:

```yaml
# config.yaml
defaults:
  http:
    proxy: http://proxy.corp.example.com:8080
    no_proxy:
      - localhost
      - 127.0.0.1
      - .corp.example.com
```

**Environment Variables** (alternative):

```bash
export HTTP_PROXY=http://proxy.corp.example.com:8080
export HTTPS_PROXY=http://proxy.corp.example.com:8080
export NO_PROXY=localhost,127.0.0.1,.corp.example.com
```

**Authenticated Proxy**:

```yaml
defaults:
  http:
    proxy: http://username:password@proxy.corp.example.com:8080
```

**SOCKS Proxy**:

```yaml
defaults:
  http:
    proxy: socks5://proxy.corp.example.com:1080
```

#### Firewall Rules

**Required Outbound Access**:

```bash
# API endpoints
api.example.com:443          # HTTPS API access
api.corp.example.com:443     # Internal API

# Update servers
releases.example.com:443     # Binary updates
spec.example.com:443         # OpenAPI spec

# Authentication
oauth.example.com:443        # OAuth2 token endpoint
sso.corp.example.com:443     # Corporate SSO

# Optional (can be disabled)
telemetry.example.com:443    # Usage metrics
sentry.io:443                # Error tracking
```

**Firewall Configuration** (iptables):

```bash
# Allow outbound HTTPS to API
sudo iptables -A OUTPUT -p tcp -d api.corp.example.com --dport 443 -j ACCEPT

# Allow outbound HTTPS to update server
sudo iptables -A OUTPUT -p tcp -d releases.example.com --dport 443 -j ACCEPT

# Block all other outbound (if required)
sudo iptables -A OUTPUT -p tcp --dport 443 -j DROP
```

---

### 4.3 CA Bundle Configuration

#### Custom Certificate Authority

**Configuration**:

```yaml
# config.yaml
defaults:
  http:
    # Path to corporate CA bundle
    tls_ca_file: /etc/pki/tls/certs/corp-ca-bundle.crt

    # OR: Disable TLS verification (NOT RECOMMENDED)
    # tls_insecure_skip_verify: true
```

**Installing Corporate CA**:

```bash
# Copy corporate CA certificate
sudo cp corp-root-ca.crt /usr/local/share/ca-certificates/

# Update system CA bundle
sudo update-ca-certificates

# Verify
openssl s_client -connect api.corp.example.com:443 -CAfile /etc/ssl/certs/ca-certificates.crt
```

**Client Certificate Authentication** (mTLS):

```yaml
# config.yaml
defaults:
  http:
    tls_cert_file: /etc/mycli/client.crt
    tls_key_file: /etc/mycli/client.key
    tls_ca_file: /etc/pki/tls/certs/corp-ca-bundle.crt
```

---

### 4.4 SSO/SAML Integration

#### OAuth2 with Corporate SSO

**Configuration**:

```yaml
# config.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: mycli-corp
      client_secret: ${MYCLI_CLIENT_SECRET}
      token_url: https://sso.corp.example.com/oauth/token
      auth_url: https://sso.corp.example.com/oauth/authorize
      redirect_url: http://localhost:8080/callback
      flow: authorization_code
      scopes:
        - openid
        - profile
        - mycli:access
      pkce: true
```

**Authentication Flow**:

```bash
$ mycli login

Opening browser for authentication...
→ https://sso.corp.example.com/oauth/authorize?client_id=...

[Browser opens, user logs in via corporate SSO]

✓ Authentication successful!
Token stored in system keyring.

$ mycli users list
[Command executes with authenticated session]
```

#### SAML Authentication

**Configuration**:

```yaml
behaviors:
  auth:
    type: saml
    saml:
      idp_metadata_url: https://sso.corp.example.com/saml/metadata
      sp_entity_id: mycli-corp
      assertion_consumer_service_url: http://localhost:8080/saml/acs
      sign_authn_request: true
      cert_file: /etc/mycli/saml.crt
      key_file: /etc/mycli/saml.key
```

#### Active Directory / LDAP

**Configuration**:

```yaml
behaviors:
  auth:
    type: ldap
    ldap:
      url: ldaps://ldap.corp.example.com:636
      bind_dn: uid=%s,ou=users,dc=corp,dc=example,dc=com
      search_base: ou=users,dc=corp,dc=example,dc=com
      search_filter: (uid=%s)
      tls:
        ca_file: /etc/pki/tls/certs/corp-ca-bundle.crt
```

---

## Troubleshooting in Production

### 5.1 Common Production Issues

#### Issue: "Connection Timeout"

**Symptoms**:

```
Error: Failed to execute command
  Reason: connection timeout after 30s
  Endpoint: https://api.example.com/v1/users
```

**Diagnosis**:

```bash
# 1. Check network connectivity
curl -v https://api.example.com/v1/users

# 2. Check DNS resolution
nslookup api.example.com
dig api.example.com

# 3. Check proxy configuration
env | grep -i proxy
mycli config show | grep proxy

# 4. Test with increased timeout
MYCLI_TIMEOUT=120s mycli users list

# 5. Check firewall rules
sudo iptables -L -n | grep 443
```

**Resolution**:

```yaml
# Increase timeout in config
defaults:
  http:
    timeout: 120s
    retry:
      max_attempts: 3
      backoff: exponential
```

---

#### Issue: "Authentication Failed"

**Symptoms**:

```
Error: Authentication failed
  Reason: invalid_token
  Details: Token expired or revoked
```

**Diagnosis**:

```bash
# 1. Check token status
mycli auth status

# 2. Check token storage
ls -la ~/.config/mycli/auth/
# OR
security find-generic-password -s mycli  # macOS Keychain

# 3. Attempt re-authentication
mycli auth login --force

# 4. Enable debug logging
MYCLI_LOG_LEVEL=debug mycli users list
```

**Resolution**:

```bash
# Clear and re-authenticate
mycli auth logout
mycli auth login

# If persistent, check OAuth configuration
mycli config show | grep oauth
```

---

#### Issue: "Rate Limit Exceeded"

**Symptoms**:

```
Error: Rate limit exceeded
  Limit: 1000 requests/hour
  Current: 1023
  Reset: 2025-11-25T15:00:00Z (in 45m)
```

**Diagnosis**:

```bash
# Check rate limit headers
mycli --debug users list 2>&1 | grep -i "rate-limit"

# Review usage patterns
grep "api.request" /var/log/mycli/mycli.log | \
  awk '{print $1}' | uniq -c | sort -rn

# Enable request caching
mycli config set caching.enabled true
```

**Resolution**:

```yaml
# Enable aggressive caching
defaults:
  caching:
    enabled: true
    ttl: 5m

  retry:
    max_attempts: 5
    backoff: exponential
    backoff_multiplier: 2
```

---

### 5.2 Debug Mode in Production

#### Enabling Debug Mode Safely

**Temporary Debug (Single Command)**:

```bash
# Enable debug for one command
MYCLI_LOG_LEVEL=debug mycli users list 2> debug.log

# Review debug output
less debug.log
```

**Debug Configuration** (for troubleshooting):

```yaml
# /tmp/debug-config.yaml
defaults:
  logging:
    level: debug
    output: file
    file: /tmp/mycli-debug.log
```

**Use Debug Config**:

```bash
# Run with debug config
MYCLI_CONFIG=/tmp/debug-config.yaml mycli users list

# Review debug log
tail -f /tmp/mycli-debug.log
```

**Debug Build Warning**:

```bash
# If using debug build in production
$ mycli --version

╔════════════════════════════════════════════════════════════╗
║  🚨 DEBUG MODE ENABLED - SECURITY WARNING                 ║
╠════════════════════════════════════════════════════════════╣
║  This is a DEBUG BUILD.                                    ║
║  All embedded configuration can be overridden.             ║
║  ⚠️  DO NOT USE IN PRODUCTION                             ║
╚════════════════════════════════════════════════════════════╝

mycli version 1.2.3 (debug)
```

---

### 5.3 Log Analysis

#### Structured Log Queries

**Using jq** (JSON logs):

```bash
# Extract errors from last hour
cat /var/log/mycli/mycli.log | \
  jq -r 'select(.timestamp > (now - 3600)) | select(.level == "error")'

# Count errors by type
cat /var/log/mycli/mycli.log | \
  jq -r 'select(.level == "error") | .message' | \
  sort | uniq -c | sort -rn

# Analyze slow commands (>1s)
cat /var/log/mycli/mycli.log | \
  jq -r 'select(.fields.duration_ms > 1000) |
         "\(.fields.command): \(.fields.duration_ms)ms"'

# Extract API errors
cat /var/log/mycli/mycli.log | \
  jq -r 'select(.fields.status_code >= 400) |
         "\(.timestamp) \(.fields.command) \(.fields.status_code)"'
```

**Using grep** (text logs):

```bash
# Find errors in last 24 hours
grep -A 5 "ERROR" /var/log/mycli/mycli.log | tail -100

# Count errors by hour
grep "ERROR" /var/log/mycli/mycli.log | \
  awk '{print $1}' | cut -d'T' -f2 | cut -d':' -f1 | \
  sort | uniq -c

# Extract stack traces
grep -A 20 "panic:" /var/log/mycli/mycli.log
```

#### Centralized Log Analysis (ELK Stack)

**Kibana Query Examples**:

```
# Errors in last 24h
level:error AND @timestamp:[now-24h TO now]

# Slow requests
fields.duration_ms:>1000

# Failed authentication
message:"authentication failed"

# By user
fields.user:"alice" AND level:error

# API errors by endpoint
fields.endpoint:"/v1/users" AND fields.status_code:>=400
```

---

### 5.4 Performance Profiling

#### CPU Profiling

**Enable CPU Profiling**:

```bash
# Run with CPU profiling
MYCLI_CPU_PROFILE=/tmp/cpu.prof mycli users list

# Analyze profile
go tool pprof /tmp/cpu.prof
(pprof) top10
(pprof) web
```

#### Memory Profiling

**Enable Memory Profiling**:

```bash
# Run with memory profiling
MYCLI_MEM_PROFILE=/tmp/mem.prof mycli users list --all

# Analyze profile
go tool pprof -alloc_space /tmp/mem.prof
(pprof) top10
(pprof) list main.main
```

#### Request Tracing

**Distributed Tracing** (OpenTelemetry):

```yaml
# config.yaml
behaviors:
  tracing:
    enabled: true
    exporter: otlp
    otlp:
      endpoint: otel-collector.example.com:4317
    service_name: mycli
    sample_rate: 0.1
```

**Trace Example**:

```
Trace ID: abc123def456
Span: mycli users list (245ms)
  ├─ Span: load-config (12ms)
  ├─ Span: fetch-spec (1ms) [cache hit]
  ├─ Span: authenticate (45ms)
  │  ├─ Span: token-refresh (30ms)
  │  └─ Span: token-storage (5ms)
  ├─ Span: http-request (178ms)
  │  ├─ Span: dns-lookup (15ms)
  │  ├─ Span: tcp-connect (20ms)
  │  ├─ Span: tls-handshake (35ms)
  │  └─ Span: http-transfer (108ms)
  └─ Span: format-output (6ms)
```

---

## Security Hardening

### 6.1 Permissions and Access Control

#### File System Permissions

**Configuration Files**:

```bash
# User config (read/write by owner only)
chmod 600 ~/.config/mycli/config.yaml

# System config (read by all, write by root)
sudo chmod 644 /etc/mycli/config.yaml

# Auth tokens (read/write by owner only)
chmod 600 ~/.config/mycli/auth/token.json

# Binary (executable by all, write by root)
sudo chmod 755 /usr/local/bin/mycli
```

**Directory Structure**:

```bash
# Create directories with correct permissions
install -d -m 0700 ~/.config/mycli
install -d -m 0700 ~/.config/mycli/auth
install -d -m 0700 ~/.cache/mycli
install -d -m 0755 /etc/mycli

# Verify permissions
find ~/.config/mycli -type f -exec chmod 600 {} \;
find ~/.config/mycli -type d -exec chmod 700 {} \;
```

#### Role-Based Access Control

**System Groups**:

```bash
# Create mycli group
sudo groupadd mycli-users

# Add users to group
sudo usermod -aG mycli-users alice
sudo usermod -aG mycli-users bob

# Set group ownership
sudo chown root:mycli-users /etc/mycli/config.yaml
sudo chmod 640 /etc/mycli/config.yaml
```

**Capability-Based Security** (Linux):

```bash
# Grant specific capabilities instead of full root
# Example: Allow binding to privileged ports
sudo setcap 'cap_net_bind_service=+ep' /usr/local/bin/mycli

# Remove sudo requirement
sudo setcap 'cap_dac_override=+ep' /usr/local/bin/mycli
```

---

### 6.2 Secrets Management in Production

#### Environment Variables

**Recommended Approach**:

```bash
# Store secrets in environment
export MYCLI_API_KEY="sk-secret-key-here"
export MYCLI_CLIENT_SECRET="oauth-secret-here"

# Run CLI (secrets injected automatically)
mycli users list
```

**Systemd Service** (with secrets):

```ini
# /etc/systemd/system/mycli-worker.service
[Unit]
Description=MyCLI Background Worker
After=network.target

[Service]
Type=simple
User=mycli
EnvironmentFile=/etc/mycli/secrets.env
ExecStart=/usr/local/bin/mycli worker start
Restart=always

[Install]
WantedBy=multi-user.target
```

**Secrets File**:

```bash
# /etc/mycli/secrets.env (chmod 600)
MYCLI_API_KEY=sk-secret-key-here
MYCLI_CLIENT_SECRET=oauth-secret-here
```

#### HashiCorp Vault Integration

**Configuration**:

```yaml
# config.yaml
behaviors:
  secrets:
    vault:
      enabled: true
      address: https://vault.corp.example.com
      auth_method: token  # token, approle, kubernetes
      token_path: /var/run/secrets/vault-token
      secret_path: secret/data/mycli
```

**Usage**:

```bash
# Authenticate with Vault
export VAULT_ADDR=https://vault.corp.example.com
export VAULT_TOKEN=$(cat /var/run/secrets/vault-token)

# CLI automatically fetches secrets from Vault
mycli users list
```

#### AWS Secrets Manager

**Configuration**:

```yaml
behaviors:
  secrets:
    aws:
      enabled: true
      region: us-east-1
      secret_name: mycli/production
      auth: iam  # Use IAM role credentials
```

**IAM Policy**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue",
        "secretsmanager:DescribeSecret"
      ],
      "Resource": "arn:aws:secretsmanager:us-east-1:123456789012:secret:mycli/production-*"
    }
  ]
}
```

#### Kubernetes Secrets

**Configuration**:

```yaml
# config.yaml
behaviors:
  secrets:
    kubernetes:
      enabled: true
      namespace: default
      secret_name: mycli-secrets
```

**Kubernetes Secret**:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: mycli-secrets
  namespace: default
type: Opaque
stringData:
  api-key: sk-secret-key-here
  client-secret: oauth-secret-here
```

---

### 6.3 Audit Logging

#### Enable Audit Logging

**Configuration**:

```yaml
# config.yaml
defaults:
  audit:
    enabled: true
    output: syslog  # file, syslog, stdout
    file: /var/log/mycli/audit.log
    format: json
    include_request_body: false  # Privacy consideration
    include_response_body: false
```

**Audit Log Entry**:

```json
{
  "timestamp": "2025-11-25T14:30:15Z",
  "event": "command.executed",
  "user": "alice",
  "command": "users delete",
  "args": {
    "user_id": "12345"
  },
  "result": "success",
  "duration_ms": 450,
  "request_id": "req-abc123",
  "client_version": "1.2.3",
  "source_ip": "192.168.1.100",
  "session_id": "sess-xyz789"
}
```

#### Audit Log Analysis

**Query Recent Commands**:

```bash
# Last 100 commands by user
jq -r 'select(.user == "alice")' /var/log/mycli/audit.log | tail -100

# Failed commands
jq -r 'select(.result == "error")' /var/log/mycli/audit.log

# Destructive operations
jq -r 'select(.command | contains("delete"))' /var/log/mycli/audit.log
```

---

### 6.4 Compliance Considerations

#### GDPR Compliance

**Data Minimization**:

```yaml
# config.yaml
behaviors:
  telemetry:
    enabled: true
    include_user_id: false  # Don't send PII
    anonymize_ip: true

  audit:
    enabled: true
    pii_masking: true  # Mask sensitive fields
    retention_days: 90  # Auto-delete old logs
```

**User Data Export**:

```bash
# Export user's data
mycli gdpr export --user alice > alice-data.json

# Delete user's data
mycli gdpr delete --user alice --confirm
```

#### SOC 2 Compliance

**Required Configurations**:

```yaml
# Encryption in transit
defaults:
  http:
    tls_min_version: "1.2"

# Encryption at rest (for cached data)
behaviors:
  caching:
    encryption: true
    encryption_key: ${MYCLI_CACHE_ENCRYPTION_KEY}

# Audit logging
defaults:
  audit:
    enabled: true
    tamper_proof: true  # Immutable logs

# Access controls
behaviors:
  auth:
    require_mfa: true
    session_timeout: 8h
```

#### HIPAA Compliance

**Additional Requirements**:

```yaml
# PHI protection
behaviors:
  data_classification:
    enabled: true
    rules:
      - pattern: "ssn"
        classification: phi
        action: mask
      - pattern: "dob"
        classification: phi
        action: mask

# Audit all data access
defaults:
  audit:
    enabled: true
    log_data_access: true
    include_phi_indicator: true
```

---

## Disaster Recovery

### 7.1 Backup and Restore

#### Configuration Backup

**Backup Script**:

```bash
#!/bin/bash
# backup-mycli-config.sh

BACKUP_DIR="/var/backups/mycli"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup user configs
tar czf "$BACKUP_DIR/user-configs-$TIMESTAMP.tar.gz" \
  ~/.config/mycli/ \
  ~/.cache/mycli/ \
  2>/dev/null

# Backup system config
sudo tar czf "$BACKUP_DIR/system-config-$TIMESTAMP.tar.gz" \
  /etc/mycli/ \
  2>/dev/null

# Backup auth tokens (encrypted)
if command -v gpg &> /dev/null; then
  tar czf - ~/.config/mycli/auth/ 2>/dev/null | \
    gpg -e -r admin@example.com > "$BACKUP_DIR/auth-$TIMESTAMP.tar.gz.gpg"
fi

# Cleanup old backups (keep 30 days)
find "$BACKUP_DIR" -type f -mtime +30 -delete

echo "✓ Backup complete: $BACKUP_DIR"
```

**Automated Backup** (cron):

```bash
# Add to crontab
0 2 * * * /usr/local/bin/backup-mycli-config.sh
```

#### Restore from Backup

```bash
# List available backups
ls -lh /var/backups/mycli/

# Restore user config
tar xzf /var/backups/mycli/user-configs-20251125-020000.tar.gz -C ~/

# Restore system config
sudo tar xzf /var/backups/mycli/system-config-20251125-020000.tar.gz -C /

# Restore auth tokens (decrypt)
gpg -d /var/backups/mycli/auth-20251125-020000.tar.gz.gpg | \
  tar xzf - -C ~/
```

---

### 7.2 Failover Strategies

#### Multi-Region Deployment

**Configuration**:

```yaml
# config.yaml
api:
  # Primary endpoint
  base_url: https://api.us-east-1.example.com

  # Failover endpoints
  failover:
    - url: https://api.us-west-2.example.com
      priority: 1
    - url: https://api.eu-west-1.example.com
      priority: 2

  # Health check
  health_check:
    enabled: true
    endpoint: /health
    interval: 30s
    timeout: 5s
```

**Automatic Failover**:

```
┌────────────────────────────────────────────────────────────┐
│                   FAILOVER PROCESS                         │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  1. Client attempts request to primary                    │
│     → https://api.us-east-1.example.com                   │
│                                                            │
│  2. Request fails (timeout/5xx error)                     │
│     ✗ Connection timeout after 30s                        │
│                                                            │
│  3. Mark primary as unhealthy                             │
│     └─ Will retry after 60s                               │
│                                                            │
│  4. Attempt failover endpoint (priority 1)                │
│     → https://api.us-west-2.example.com                   │
│                                                            │
│  5. Request succeeds                                      │
│     ✓ Response received from us-west-2                    │
│                                                            │
│  6. Continue using failover endpoint                      │
│     └─ Periodically health-check primary                  │
│                                                            │
│  7. Primary recovers                                      │
│     ✓ Health check passed                                 │
│                                                            │
│  8. Fail back to primary                                  │
│     → Resume using https://api.us-east-1.example.com      │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## Performance Optimization

### 8.1 Caching Strategies

#### Spec Caching

**Configuration**:

```yaml
# config.yaml
behaviors:
  caching:
    spec:
      enabled: true
      ttl: 5m  # Cache spec for 5 minutes
      max_age: 24h  # Revalidate if older than 24h
```

**Cache Control**:

```bash
# Clear spec cache
mycli cache clear spec

# Disable cache for one command
mycli --no-cache users list

# Pre-warm cache
mycli cache warm
```

#### Response Caching

**Configuration**:

```yaml
behaviors:
  caching:
    responses:
      enabled: true
      ttl: 60s
      max_entries: 1000
      # Cache only safe methods
      methods:
        - GET
        - HEAD
```

---

### 8.2 Connection Pooling

**Configuration**:

```yaml
defaults:
  http:
    connection_pool:
      max_idle_conns: 100
      max_idle_conns_per_host: 10
      idle_conn_timeout: 90s
      keep_alive: 30s
```

---

### 8.3 Parallel Execution

**Batch Operations**:

```bash
# Sequential (slow)
mycli users list --all

# Parallel (faster)
mycli users list --all --parallel 10
```

---

## Summary

This operations guide covers:

1. **Production Deployment**: Multiple distribution strategies (binaries, packages, Docker, air-gapped)
2. **Monitoring**: Comprehensive logging, metrics, error tracking, and performance monitoring
3. **Updates**: Self-update mechanisms, rollback procedures, and controlled rollouts
4. **Enterprise**: Centralized configuration, proxy setup, CA bundles, and SSO integration
5. **Troubleshooting**: Common issues, debug mode, log analysis, and performance profiling
6. **Security**: Permissions, secrets management, audit logging, and compliance
7. **Disaster Recovery**: Backup/restore procedures and failover strategies
8. **Performance**: Caching, connection pooling, and parallel execution

**Next Steps**:

- Review your organization's specific requirements
- Customize configurations for your environment
- Set up monitoring and alerting
- Establish backup and disaster recovery procedures
- Conduct security audits and penetration testing

**Additional Resources**:

- [Configuration DSL Reference](configuration-dsl.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Security Best Practices](security-guide.md)
- [API Documentation](https://docs.example.com/api)
