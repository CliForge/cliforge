# CliForge Security Guide

**Version:** 1.0.0
**Last Updated:** 2025-01-25
**Target Audience:** CLI Developers, Security Teams, DevOps Engineers, End Users

---

## Table of Contents

1. [Introduction](#introduction)
2. [Security Architecture](#security-architecture)
3. [Credential Management](#credential-management)
4. [Network Security](#network-security)
5. [Configuration Security](#configuration-security)
6. [Audit and Compliance](#audit-and-compliance)
7. [Security Best Practices](#security-best-practices)
8. [Vulnerability Management](#vulnerability-management)
9. [Security Checklist](#security-checklist)

---

## Introduction

CliForge is a hybrid CLI generation framework that combines static binary generation with dynamic API spec loading. This unique architecture requires careful attention to security at multiple levels: build time, distribution, and runtime.

### Security Philosophy

CliForge follows defense-in-depth principles:

1. **Secure by Default**: Security features are enabled by default
2. **Least Privilege**: Minimal permissions and access required
3. **Transparency**: Clear security boundaries and trust relationships
4. **User Control**: Users maintain control over their credentials and data
5. **Continuous Protection**: Multi-layer security from development to runtime

### Document Scope

This guide covers:

- **CLI Developers**: Security considerations when building CLIs with CliForge
- **API Owners**: Security requirements for APIs consumed by generated CLIs
- **Operations Teams**: Deployment, monitoring, and compliance considerations
- **End Users**: Secure usage of CliForge-generated CLIs

---

## Security Architecture

### Threat Model

CliForge-generated CLIs face threats at multiple stages:

#### Development Time Threats

| Threat | Impact | Mitigation |
|--------|--------|------------|
| **Malicious configuration injection** | Backdoors in generated binaries | Configuration validation, code review |
| **Embedded secrets** | Credential leakage in distributed binaries | Secret detection, build-time scanning |
| **Dependency vulnerabilities** | Compromised dependencies | Dependency scanning, pinned versions |
| **Supply chain attacks** | Compromised build toolchain | Reproducible builds, signed releases |

#### Distribution Threats

| Threat | Impact | Mitigation |
|--------|--------|------------|
| **Binary tampering** | Malicious binary distribution | Code signing, checksum verification |
| **Man-in-the-middle downloads** | Binary interception/replacement | HTTPS, certificate pinning |
| **Typosquatting** | Users download fake CLI | Official distribution channels, verification |
| **Outdated binaries** | Users run vulnerable versions | Update notifications, forced updates |

#### Runtime Threats

| Threat | Impact | Mitigation |
|--------|--------|------------|
| **Credential theft** | Unauthorized API access | Keyring storage, token encryption |
| **API spec poisoning** | Malicious commands injected | Spec signature verification, caching |
| **Man-in-the-middle attacks** | Traffic interception | TLS, certificate validation |
| **Token replay attacks** | Stolen tokens used | Short-lived tokens, token refresh |
| **Information disclosure** | Secrets logged or displayed | Secret masking, sanitized output |
| **Configuration override attacks** | Malicious settings override | Locked configurations, debug mode controls |

### Security Boundaries

CliForge establishes clear security boundaries:

```
┌─────────────────────────────────────────────────────────────────┐
│                       SECURITY BOUNDARIES                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ TRUSTED ZONE                                            │    │
│  │ - Embedded configuration (signed, locked)              │    │
│  │ - Binary code (code-signed)                            │    │
│  │ - Default behaviors (immutable in production)          │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ SEMI-TRUSTED ZONE                                       │    │
│  │ - User preferences (validated, limited override)       │    │
│  │ - Cached OpenAPI specs (TTL-limited, validated)        │    │
│  │ - Environment variables (sanitized)                    │    │
│  └────────────────────────────────────────────────────────┘    │
│                           │                                      │
│                           ▼                                      │
│  ┌────────────────────────────────────────────────────────┐    │
│  │ UNTRUSTED ZONE                                          │    │
│  │ - Runtime OpenAPI specs (from API endpoint)            │    │
│  │ - API responses (validated against schema)             │    │
│  │ - User input (sanitized, validated)                    │    │
│  └────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Trust Relationships

1. **CLI Developer → Generated Binary**
   - Developer signs embedded configuration
   - Binary includes only validated, secure defaults
   - No secrets embedded in binary

2. **End User → CLI Binary**
   - User verifies binary authenticity (checksums, signatures)
   - User trusts embedded configuration from developer
   - User maintains control over credentials

3. **CLI Binary → API Server**
   - TLS with certificate validation
   - Mutual authentication (API validates CLI tokens)
   - API spec served over HTTPS with optional signatures

4. **CLI Binary → Update Server**
   - HTTPS with certificate pinning
   - Binary signature verification before update
   - Rollback capability if update fails

### Attack Surface Analysis

#### Minimal Attack Surface

**Embedded Binary**:
- Single statically-linked binary (no dynamic libraries)
- No external dependencies at runtime
- Configuration embedded at compile time
- No runtime code loading

**Network Communication**:
- Only HTTPS connections (TLS 1.2+)
- Certificate validation enforced
- No insecure fallback mechanisms
- Configurable proxy support

**Local Storage**:
- XDG-compliant directories (predictable, user-controlled)
- Restrictive file permissions (0600 for sensitive files)
- No global/system directories required
- Optional keyring integration

#### Attack Surface Reduction Strategies

1. **Minimize Dependencies**
   - Pure Go implementation (no CGO by default)
   - Minimal third-party dependencies
   - Vetted, well-maintained libraries only

2. **Input Validation**
   - All user input validated against schemas
   - OpenAPI spec validation at load time
   - Configuration validation at build time

3. **Output Sanitization**
   - Automatic secret masking in logs
   - Error messages sanitized
   - Debug output controlled

4. **Privilege Minimization**
   - No root/admin privileges required
   - User-level file system access only
   - Optional keyring access (with fallback)

---

## Credential Management

### Overview

CliForge supports multiple authentication mechanisms with varying security characteristics. Proper credential management is critical to CLI security.

### Credential Storage Options

#### 1. OS Keyring Storage (Recommended)

**Security Level**: HIGH

**Platforms**:
- **macOS**: Keychain (Secure Enclave on supported hardware)
- **Windows**: Credential Manager (DPAPI encryption)
- **Linux**: Secret Service API (GNOME Keyring, KWallet)

**Configuration**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: keyring
        keyring_service: myapp-cli
        keyring_user: default
```

**Security Properties**:
- OS-managed encryption keys
- Hardware-backed storage (when available)
- Per-user isolation
- Access control via OS permissions
- Survives system reboots
- Not accessible to other users

**Threats Mitigated**:
- ✓ Credential theft from filesystem
- ✓ Credential extraction from memory dumps
- ✓ Cross-user credential access
- ✓ Accidental credential exposure

**Limitations**:
- Requires keyring service running
- May require user interaction (password unlock)
- Platform-specific implementations
- Not available in headless environments

#### 2. File Storage (Encrypted)

**Security Level**: MEDIUM

**Configuration**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: file
        path: ~/.myapp/token.json  # Will be encrypted
```

**File Location** (XDG-compliant):
- **Linux/macOS**: `~/.local/share/myapp/token.json`
- **Windows**: `%LOCALAPPDATA%\myapp\token.json`

**Security Properties**:
- File permissions: 0600 (owner read/write only)
- AES-256 encryption of token data
- Encryption key derived from machine ID + user ID
- Tamper detection via HMAC

**Threats Mitigated**:
- ✓ Accidental file disclosure
- ✓ Casual inspection attacks
- ○ Determined attacker with file access

**Limitations**:
- Encryption key stored on same system
- Vulnerable to privileged user attacks
- No hardware-backed protection
- Portable across systems if copied

**When to Use**:
- Keyring not available
- Headless/CI environments
- Development/testing
- Cross-platform portability needed

#### 3. Memory Storage (Ephemeral)

**Security Level**: LOW

**Configuration**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: memory
```

**Security Properties**:
- Tokens never written to disk
- Lost when CLI process exits
- No persistent storage
- Memory cleared on exit

**Threats Mitigated**:
- ✓ Token persistence attacks
- ✓ File-based token theft
- ○ Memory dumps while running

**Limitations**:
- Must re-authenticate each CLI invocation
- No token refresh across runs
- Inconvenient for regular use

**When to Use**:
- CI/CD pipelines with ephemeral tokens
- Security-sensitive environments
- Testing/development
- Single-use operations

### Multi-Tier Storage (Fallback Strategy)

**Configuration**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        # Try keyring first, fall back to file
        primary: keyring
        fallback: file

        keyring:
          service: myapp-cli
          user: default

        file:
          path: ~/.myapp/token.json
```

**Behavior**:
- Attempts keyring storage first
- Falls back to encrypted file if keyring unavailable
- Transparent to user
- Logs storage method used (for audit)

**Benefits**:
- Best security when available
- Graceful degradation
- Cross-platform compatibility
- User-friendly

### Token Lifecycle Management

#### Token Acquisition

**OAuth2 Authorization Code Flow (Recommended)**:

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      client_id: myapp-cli
      auth_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      redirect_url: http://localhost:8085/callback
      pkce: true  # CRITICAL: Enable PKCE for public clients
      scopes:
        - read:api
        - write:api
```

**Security Considerations**:

1. **PKCE (Proof Key for Code Exchange)**:
   - **ALWAYS ENABLE** for public clients (CLIs, mobile apps)
   - Prevents authorization code interception attacks
   - No client secret required (public clients can't keep secrets)
   - RFC 7636 compliant

2. **Redirect URL**:
   - Use `http://localhost:<random-port>/callback` (loopback only)
   - **Never** use `http://0.0.0.0` or public IPs
   - Port should be configurable (avoid conflicts)
   - Validate redirect in authorization server

3. **State Parameter**:
   - Automatically generated random value
   - Prevents CSRF attacks
   - Validated on callback

4. **Scope Minimization**:
   - Request only required scopes
   - User reviews requested permissions
   - Scope down-scoping not supported (by design)

**Client Credentials Flow (Service Accounts)**:

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: client_credentials
      client_id: myapp-service
      client_secret: ${OAUTH_CLIENT_SECRET}  # Environment variable
      token_url: https://auth.example.com/token
      scopes:
        - api:read
        - api:write
```

**Security Considerations**:

1. **Client Secret Protection**:
   - **NEVER** embed in configuration file
   - **ALWAYS** use environment variables
   - Rotate secrets regularly (90 days recommended)
   - Use secrets management system (Vault, AWS Secrets Manager)

2. **Token Storage**:
   - File storage acceptable (encrypted)
   - Short TTL preferred (< 1 hour)
   - Monitor token usage

3. **Audit Logging**:
   - Log all token acquisitions
   - Track token usage patterns
   - Alert on anomalies

#### Token Refresh

**Automatic Refresh**:

```go
// Built-in refresh logic (automatic)
if token.ExpiresAt.Before(time.Now().Add(30 * time.Second)) {
    // Refresh token if it expires within 30 seconds
    newToken, err := authenticator.RefreshToken(ctx, token)
    if err == nil {
        storage.SaveToken(ctx, newToken)
    }
}
```

**Security Properties**:
- Automatic refresh before expiration
- Refresh tokens rotated (if supported by server)
- Old tokens invalidated
- Refresh failures trigger re-authentication

**Threats Mitigated**:
- ✓ Expired token errors
- ✓ Long-lived token exposure
- ✓ Token replay attacks (with rotation)

**Best Practices**:
- Access token TTL: 15-60 minutes
- Refresh token TTL: 7-30 days
- Rotate refresh tokens on use
- Revoke old refresh tokens

#### Token Revocation

**Explicit Logout**:

```bash
# Revoke token and clear storage
myapp logout

# Revoke all tokens (all providers)
myapp logout --all

# Revoke token on server (if supported)
myapp logout --revoke
```

**Automatic Revocation Scenarios**:
- User-initiated logout
- Token refresh failure (expired refresh token)
- Authentication error (401 Unauthorized)
- Security policy violation

**Server-Side Revocation**:
- CLI calls revocation endpoint (if available)
- Revokes both access and refresh tokens
- Handles revocation errors gracefully

### Environment Variables for Credentials

**Secure Environment Variable Usage**:

```yaml
# API Key
behaviors:
  auth:
    type: api_key
    api_key:
      env_var: MYAPP_API_KEY  # NEVER hardcode the key
```

```bash
# Set environment variable
export MYAPP_API_KEY=sk_live_abc123def456

# Use CLI
myapp users list
```

**Security Best Practices**:

1. **Variable Naming Convention**:
   ```
   <CLI_NAME>_<CREDENTIAL_TYPE>
   Examples: MYAPP_API_KEY, MYAPP_TOKEN, OAUTH_CLIENT_SECRET
   ```

2. **Environment File Security**:
   ```bash
   # Create environment file
   cat > ~/.myapp.env <<EOF
   export MYAPP_API_KEY=sk_live_abc123
   EOF

   # Set restrictive permissions
   chmod 600 ~/.myapp.env

   # Source in session
   source ~/.myapp.env
   ```

3. **Shell History Protection**:
   ```bash
   # Prevent credential leakage in history
   HISTCONTROL=ignorespace
    export MYAPP_API_KEY=secret  # Leading space prevents history

   # Or disable history temporarily
   set +o history
   export MYAPP_API_KEY=secret
   set -o history
   ```

4. **CI/CD Secrets Management**:

   **GitHub Actions**:
   ```yaml
   - name: Run CLI
     env:
       MYAPP_API_KEY: ${{ secrets.MYAPP_API_KEY }}
     run: myapp users list
   ```

   **GitLab CI**:
   ```yaml
   test:
     variables:
       MYAPP_API_KEY: $MYAPP_API_KEY_SECRET
     script:
       - myapp users list
   ```

   **AWS**:
   ```bash
   # Use AWS Secrets Manager
   export MYAPP_API_KEY=$(aws secretsmanager get-secret-value \
     --secret-id myapp-api-key \
     --query SecretString \
     --output text)
   ```

**Threats Mitigated**:
- ✓ Hardcoded secrets in config files
- ✓ Secrets in version control
- ✓ Secrets in command history
- ○ Environment variable exposure (process listing)

**Limitations**:
- Visible in process environment (ps, /proc)
- Inherited by child processes
- May be logged by shells/IDEs
- No encryption at rest

### API Key Security

**API Key Configuration**:

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      location: header          # header or query
      name: X-API-Key           # Header name
      prefix: ""                # Optional prefix (e.g., "Bearer ")
      env_var: MYAPP_API_KEY    # Environment variable
```

**Security Considerations**:

1. **Header vs Query Parameter**:
   - **Prefer headers**: More secure (not logged in URL)
   - **Avoid query**: Logged in server logs, browser history, proxies
   - **Exception**: API design constraints

2. **API Key Format**:
   - Use high-entropy keys (>= 256 bits)
   - Include prefix for identification (e.g., `sk_live_`, `pk_test_`)
   - Versioning support (rotate keys gracefully)

3. **API Key Rotation**:
   ```bash
   # Generate new key
   NEW_KEY=$(myapp auth rotate-key)

   # Update environment
   export MYAPP_API_KEY=$NEW_KEY

   # Verify new key works
   myapp auth verify

   # Revoke old key (server-side)
   myapp auth revoke-old-keys
   ```

4. **API Key Scopes**:
   - Use scoped keys (read-only, write-only)
   - Separate keys for different environments
   - Limit key capabilities

**Threats Mitigated**:
- ✓ Key exposure in URLs (header-based)
- ✓ Unlimited key lifetime (with rotation)
- ✓ Excessive permissions (with scopes)

### Basic Authentication (Legacy)

**Configuration**:

```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: MYAPP_USERNAME
      password_env: MYAPP_PASSWORD
```

**Security Warning**:
```
⚠️  SECURITY WARNING: Basic Authentication
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Basic auth sends credentials with EVERY request.
This is less secure than OAuth2 token-based auth.

RECOMMENDATIONS:
1. Use OAuth2 instead (authorization_code or client_credentials)
2. If Basic auth required, ALWAYS use HTTPS
3. Never log credentials
4. Implement credential rotation
5. Monitor for credential leakage
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Security Requirements**:
- **HTTPS ONLY** (never HTTP)
- Short-lived credentials
- Credential rotation
- Rate limiting (prevent brute force)
- Account lockout policies

**Migration Path to OAuth2**:
```yaml
# Phase 1: Support both Basic and OAuth2
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: password  # Resource Owner Password Credentials
      # ... OAuth2 config ...

    # Fallback to Basic for legacy clients
    fallback_auth:
      type: basic

# Phase 2: OAuth2 only
behaviors:
  auth:
    type: oauth2
```

### Credential Masking and Protection

CliForge automatically masks credentials in logs, debug output, and error messages.

**Automatic Secret Detection**:

```yaml
behaviors:
  secrets:
    enabled: true

    # Field name patterns (case-insensitive)
    field_patterns:
      - "*password*"
      - "*secret*"
      - "*token*"
      - "*key"
      - "*credential*"
      - "auth*"

    # Value patterns (regex)
    value_patterns:
      - name: "AWS Access Key"
        pattern: 'AKIA[0-9A-Z]{16}'
        enabled: true
      - name: "Generic API Key"
        pattern: '[sS][kK]_[a-zA-Z0-9]{32,}'
        enabled: true
      - name: "JWT Token"
        pattern: 'eyJ[A-Za-z0-9-_=]+\.eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_.+/=]*'
        enabled: true
      - name: "GitHub Token"
        pattern: 'ghp_[a-zA-Z0-9]{36}'
        enabled: true

    # Masking style
    masking:
      style: partial  # partial, full, hash
      partial_show_chars: 6
      replacement: "***"

    # Where to mask
    mask_in:
      stdout: true
      stderr: true
      logs: true
      debug_output: true
      audit: false  # Audit logs use hash, not plaintext
```

**Masking Examples**:

```bash
# Original value
sk_live_abc123def456ghi789

# Partial masking (default)
sk_liv***

# Full masking
***

# Hash masking (audit logs)
sha256:a3f2b8c1...
```

**OpenAPI Extensions for Secret Fields**:

```yaml
# Mark field as secret in OpenAPI spec
components:
  schemas:
    User:
      type: object
      properties:
        api_key:
          type: string
          x-cli-secret: true  # Masked in output
        password:
          type: string
          writeOnly: true
          x-cli-secret: true
```

**Debug Output Protection**:

```bash
# Even with --debug, secrets are masked
myapp --debug users create --api-key sk_live_abc123

# Output:
DEBUG: POST /users
DEBUG: Headers: {"X-API-Key": "sk_liv***"}
DEBUG: Response: {"id": "usr_123", "api_key": "sk_liv***"}
```

**Threats Mitigated**:
- ✓ Credential leakage in logs
- ✓ Credential exposure in error messages
- ✓ Credential disclosure in debug output
- ✓ Accidental credential sharing (copy/paste logs)

---

## Network Security

### TLS/SSL Configuration

**Default TLS Settings**:

```yaml
behaviors:
  http:
    # TLS configuration
    tls:
      min_version: "1.2"        # Minimum TLS 1.2
      max_version: "1.3"        # Prefer TLS 1.3
      verify_certificates: true  # ALWAYS true in production

      # Certificate verification
      ca_certificates: []       # Additional trusted CAs
      client_cert: ""           # Mutual TLS client cert
      client_key: ""            # Mutual TLS client key

      # Cipher suites (secure defaults)
      cipher_suites:
        - TLS_AES_128_GCM_SHA256
        - TLS_AES_256_GCM_SHA384
        - TLS_CHACHA20_POLY1305_SHA256

      # HTTPS enforcement
      strict_https: true         # Reject HTTP connections
```

**Security Requirements**:

1. **TLS Version Requirements**:
   - Minimum: TLS 1.2
   - Recommended: TLS 1.3
   - Forbidden: TLS 1.0, TLS 1.1, SSL 3.0, SSL 2.0

2. **Certificate Validation**:
   - **ALWAYS** validate certificates in production
   - Verify hostname matches certificate
   - Check certificate expiration
   - Validate certificate chain
   - Check certificate revocation (OCSP, CRL)

3. **Cipher Suite Selection**:
   - Forward secrecy (ECDHE, DHE)
   - AEAD modes (GCM, ChaCha20-Poly1305)
   - No weak ciphers (RC4, DES, 3DES)
   - No export ciphers

**Insecure TLS (Development Only)**:

```yaml
# DANGER: Only for development/testing
behaviors:
  http:
    tls:
      verify_certificates: false  # ⚠️  INSECURE
      insecure_skip_verify: true  # ⚠️  INSECURE
```

```
⚠️  SECURITY WARNING: Certificate Verification Disabled
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
TLS certificate verification is DISABLED.
This makes you vulnerable to man-in-the-middle attacks.

ONLY use this in development/testing environments.
NEVER use in production.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### Certificate Validation

**Certificate Pinning (Advanced)**:

```yaml
behaviors:
  http:
    tls:
      # Pin specific certificates (SPKI hash)
      pinned_certificates:
        - sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=
        - sha256/BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=

      # Pin public keys (more flexible)
      pinned_public_keys:
        - sha256/CCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCCC=
```

**Certificate Pinning Considerations**:

**Pros**:
- Prevents CA compromise attacks
- Prevents rogue certificate issuance
- Additional layer of protection

**Cons**:
- Requires update when certificates rotate
- Can cause outages if not managed properly
- Maintenance burden

**Recommendation**:
- Use for high-security APIs
- Pin to intermediate CA, not leaf certificate
- Maintain backup pins
- Monitor certificate expiration

**Custom CA Certificates**:

```yaml
behaviors:
  http:
    tls:
      # Trust additional CA certificates
      ca_certificates:
        - /path/to/custom-ca.pem
        - /path/to/internal-ca.crt
```

**Use Cases**:
- Internal/private CAs
- Self-signed certificates (development)
- Corporate CA requirements

**Mutual TLS (mTLS)**:

```yaml
behaviors:
  http:
    tls:
      # Client certificate authentication
      client_cert: ~/.myapp/client.crt
      client_key: ~/.myapp/client.key
      client_cert_password: ${CLIENT_CERT_PASSWORD}
```

**Security Benefits**:
- Mutual authentication (server verifies client)
- Stronger authentication than bearer tokens
- Resistant to token theft

**Use Cases**:
- High-security APIs
- Service-to-service communication
- Regulated environments

### Man-in-the-Middle Protection

**MITM Attack Scenarios**:

1. **Rogue CA Certificates**:
   - Attacker installs malicious CA cert
   - Issues fake certificate for API domain
   - Intercepts TLS traffic

   **Protection**: Certificate pinning, monitor system CA store

2. **DNS Spoofing**:
   - Attacker redirects DNS queries
   - Points API domain to malicious server
   - Presents valid certificate for different domain

   **Protection**: DNSSEC, hostname verification, certificate pinning

3. **Proxy Attacks**:
   - Attacker intercepts proxy traffic
   - Downgrades to HTTP or weak TLS
   - Steals credentials

   **Protection**: Direct connections, HTTPS proxies only, proxy authentication

**MITM Detection**:

```go
// Built-in TLS connection validation
func (c *HTTPClient) validateConnection(conn *tls.Conn) error {
    state := conn.ConnectionState()

    // Check TLS version
    if state.Version < tls.VersionTLS12 {
        return fmt.Errorf("TLS version too old: %d", state.Version)
    }

    // Check cipher suite
    if !isSecureCipherSuite(state.CipherSuite) {
        return fmt.Errorf("insecure cipher suite: %d", state.CipherSuite)
    }

    // Verify certificate chain
    opts := x509.VerifyOptions{
        DNSName: conn.RemoteAddr().String(),
        Roots:   c.rootCAs,
    }
    if _, err := state.PeerCertificates[0].Verify(opts); err != nil {
        return fmt.Errorf("certificate verification failed: %w", err)
    }

    return nil
}
```

### Proxy Security

**Proxy Configuration**:

```yaml
preferences:
  http:
    # HTTP proxy
    proxy: http://proxy.example.com:8080

    # HTTPS proxy (preferred)
    https_proxy: https://proxy.example.com:8443

    # Proxy authentication
    proxy_auth:
      username: ${PROXY_USERNAME}
      password: ${PROXY_PASSWORD}

    # No proxy for specific hosts
    no_proxy:
      - localhost
      - 127.0.0.1
      - "*.internal.example.com"
```

**Proxy Security Requirements**:

1. **HTTPS Proxies Only**:
   - Proxy connection encrypted
   - Prevents proxy-level MITM
   - Protects credentials in transit

2. **Proxy Authentication**:
   - Authenticate proxy access
   - Prevent unauthorized proxy usage
   - Credentials from environment variables

3. **No Proxy Exceptions**:
   - Skip proxy for internal/local resources
   - Reduce attack surface
   - Improve performance

4. **Proxy Validation**:
   - Validate proxy server certificate
   - Verify proxy identity
   - Detect malicious proxies

**Environment Variables**:

```bash
# Proxy configuration via environment
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=https://proxy.example.com:8443
export NO_PROXY=localhost,127.0.0.1,*.internal.example.com

# Proxy authentication
export PROXY_USERNAME=user
export PROXY_PASSWORD=pass
```

**Threats Mitigated**:
- ✓ Unencrypted proxy traffic
- ✓ Unauthorized proxy access
- ✓ Proxy-based MITM attacks
- ○ Compromised proxy server (validate TLS)

---

## Configuration Security

### Debug Mode Risks

**Debug Mode** (`metadata.debug: true`) is a powerful development feature that introduces significant security risks.

**Security Impact of Debug Mode**:

| Feature | Production | Debug Mode | Risk |
|---------|-----------|-----------|------|
| Configuration override | ❌ Locked | ✅ Any setting | **CRITICAL** |
| Secret masking | ✅ Enforced | ⚠️  Can disable | **HIGH** |
| Verbose logging | ❌ Minimal | ✅ Full details | **HIGH** |
| TLS verification | ✅ Required | ⚠️  Can skip | **CRITICAL** |
| API URL override | ❌ Locked | ✅ Any URL | **HIGH** |

**Debug Mode Configuration**:

```yaml
# Embedded config (cli-config.yaml)
metadata:
  debug: false  # MUST be false for production builds
```

**What Debug Mode Enables**:

```yaml
# User config with debug_override (ONLY works if debug: true)
debug_override:
  # Override ANY embedded setting
  api:
    base_url: https://malicious-api.example.com  # ⚠️  DANGEROUS

  behaviors:
    auth:
      type: none  # ⚠️  Disable authentication

    secrets:
      enabled: false  # ⚠️  Disable secret masking

    http:
      tls:
        verify_certificates: false  # ⚠️  Disable TLS verification
```

**Security Controls**:

1. **Build-Time Protection**:
   ```bash
   # Validation during CLI generation
   cliforge build --config cli-config.yaml

   # Error if debug: true in production
   ERROR: Debug mode enabled in production build
   ERROR: Set metadata.debug: false for production
   ```

2. **Runtime Detection**:
   ```bash
   # CLI warns user if debug mode enabled
   myapp --version

   ⚠️  WARNING: DEBUG MODE ENABLED
   This CLI was built with debug mode enabled.
   Configuration can be overridden. NOT FOR PRODUCTION USE.
   ```

3. **Audit Logging**:
   ```
   # Audit log shows debug overrides
   [2025-01-25 10:30:00] DEBUG_OVERRIDE: api.base_url changed to https://dev.example.com
   [2025-01-25 10:30:00] DEBUG_OVERRIDE: behaviors.secrets.enabled set to false
   ```

**Recommendation**:
- **Development**: `debug: true` (flexibility)
- **Staging**: `debug: false` (production-like)
- **Production**: `debug: false` (REQUIRED)

### Configuration Validation

**Build-Time Validation**:

CliForge validates embedded configuration during CLI generation:

```go
// Configuration validation rules
func (v *Validator) Validate(config *cli.Config) error {
    // Metadata validation
    - Name: required, alphanumeric+hyphens, max 50 chars
    - Version: required, semantic versioning
    - Description: required, 10-200 chars

    // API validation
    - OpenAPI URL: required, valid URL or file path
    - Base URL: required, valid HTTPS URL
    - Environments: at least one default

    // Auth validation
    - Auth type: valid type (api_key, oauth2, basic, none)
    - Type-specific required fields present
    - OAuth2 URLs: valid HTTPS URLs
    - Storage config: valid storage type

    // Security validation
    - No embedded secrets detected
    - TLS verification enabled (if not debug mode)
    - HTTPS URLs (no HTTP in production)
    - Secret masking enabled

    // Behavior validation
    - Timeout: valid duration
    - Retry: reasonable max attempts (<= 10)
    - Pagination: reasonable limits
    - Cache TTL: valid duration
}
```

**Validation Errors**:

```bash
$ cliforge build --config cli-config.yaml

ERROR: Configuration validation failed:
  - metadata.name: name must contain only lowercase letters, numbers, and hyphens
  - api.base_url: base_url must be a valid URL
  - behaviors.auth.oauth2.auth_url: auth_url is required when type is oauth2
  - behaviors.secrets: embedded secret detected in config (line 45)
```

**Runtime Validation**:

```go
// User preferences validation
func (v *Validator) ValidateUserPreferences(prefs *cli.UserPreferences) error {
    // Only validate overridable settings
    - HTTP timeout: valid duration
    - Proxy URLs: valid URLs
    - Pagination limit: non-negative
    - Output format: valid format (json, yaml, table, csv)
    - Retry attempts: 0-10

    // Reject attempts to override locked settings
    if prefs.API != nil {
        return fmt.Errorf("cannot override API settings")
    }
    if prefs.Auth != nil {
        return fmt.Errorf("cannot override auth settings")
    }
}
```

### Locked vs Overridable Settings

**Configuration Override Matrix**:

| Setting | Embedded | User Preferences | Debug Override | Env Var | Flag |
|---------|----------|------------------|----------------|---------|------|
| **Locked (Never Override)** | | | | | |
| API Base URL | ✅ | ❌ | ⚠️  Debug only | ❌ | ❌ |
| OpenAPI URL | ✅ | ❌ | ⚠️  Debug only | ❌ | ❌ |
| Auth type | ✅ | ❌ | ⚠️  Debug only | ❌ | ❌ |
| OAuth2 URLs | ✅ | ❌ | ⚠️  Debug only | ❌ | ❌ |
| Branding | ✅ | ❌ | ⚠️  Debug only | ❌ | ❌ |
| **Overridable with Constraints** | | | | | |
| HTTP timeout | ✅ | ✅ | ✅ | ✅ | ✅ |
| Retry attempts | ✅ | ✅ | ✅ | ✅ | ✅ |
| Pagination limit | ✅ | ✅ | ✅ | ✅ | ✅ |
| Output format | ✅ | ✅ | ✅ | ✅ | ✅ |
| Cache TTL | ✅ | ✅ | ✅ | ✅ | ❌ |
| **Always Overridable** | | | | | |
| Verbosity | ✅ | ✅ | ✅ | ✅ | ✅ |
| Color output | ✅ | ✅ | ✅ | ✅ | ✅ |
| Proxy settings | ✅ | ✅ | ✅ | ✅ | ❌ |

**Locked Settings Rationale**:

1. **API Base URL**: Prevents redirect to malicious API
2. **OpenAPI URL**: Ensures correct API specification
3. **Auth Type**: Prevents authentication bypass
4. **OAuth2 URLs**: Prevents credential theft
5. **Branding**: Maintains CLI identity

**Override Enforcement**:

```go
// Configuration merge with override protection
func (m *Merger) MergeConfigs(embedded, user *Config, debugMode bool) (*Config, error) {
    result := embedded.DeepCopy()

    // Apply user preferences (overridable only)
    if user.Preferences != nil {
        result.Preferences = m.mergePreferences(result.Preferences, user.Preferences)
    }

    // Apply debug overrides (if debug mode enabled)
    if debugMode && user.DebugOverride != nil {
        log.Warn("Applying debug overrides - NOT FOR PRODUCTION")
        result = m.mergeDebugOverrides(result, user.DebugOverride)
    } else if user.DebugOverride != nil {
        log.Warn("Debug overrides ignored (debug mode disabled)")
    }

    return result, nil
}
```

### Secrets Detection and Masking

**Multi-Layer Secret Detection**:

1. **OpenAPI Schema Marking** (Primary):
   ```yaml
   # OpenAPI spec
   components:
     schemas:
       User:
         properties:
           api_key:
             type: string
             x-cli-secret: true  # Explicit marking
   ```

2. **Pattern-Based Detection** (Secondary):
   ```yaml
   # Field name patterns
   field_patterns:
     - "*password*"
     - "*secret*"
     - "*token*"
     - "*key"
     - "*credential*"

   # Value patterns (regex)
   value_patterns:
     - name: "AWS Access Key"
       pattern: 'AKIA[0-9A-Z]{16}'
     - name: "Generic API Key"
       pattern: '[sS][kK]_[a-zA-Z0-9]{32,}'
     - name: "JWT Token"
       pattern: 'eyJ[A-Za-z0-9-_=]+\..+'
   ```

3. **Explicit Field Paths**:
   ```yaml
   # JSONPath-like syntax
   explicit_fields:
     - "$.user.api_key"
     - "$.credentials.*"
     - "$.auth.token"
   ```

**Masking Configuration**:

```yaml
behaviors:
  secrets:
    enabled: true

    masking:
      style: partial  # partial, full, hash
      partial_show_chars: 6
      replacement: "***"

    mask_in:
      stdout: true       # Mask in command output
      stderr: true       # Mask in error messages
      logs: true         # Mask in log files
      debug_output: true # Mask even in --debug
      audit: false       # Audit uses hash, not masking
```

**Masking Examples**:

```bash
# Original API response
{
  "user_id": "usr_123",
  "api_key": "sk_live_abc123def456ghi789",
  "email": "user@example.com"
}

# Masked output (partial)
{
  "user_id": "usr_123",
  "api_key": "sk_liv***",
  "email": "user@example.com"
}

# Masked output (full)
{
  "user_id": "usr_123",
  "api_key": "***",
  "email": "user@example.com"
}

# Audit log (hash)
{
  "user_id": "usr_123",
  "api_key": "sha256:a3f2b8c1d5e6f7a8...",
  "email": "user@example.com"
}
```

**Secret Detection in Logs**:

```go
// HTTP request/response logging with secret masking
func (c *Client) Do(req *http.Request) (*http.Response, error) {
    if c.debug {
        // Mask secrets in headers
        maskedHeaders := c.detector.MaskHeaders(req.Header)
        log.Debugf("Request: %s %s", req.Method, req.URL)
        log.Debugf("Headers: %v", maskedHeaders)

        // Mask secrets in body
        if req.Body != nil {
            body, _ := io.ReadAll(req.Body)
            maskedBody := c.detector.MaskJSON(parseJSON(body))
            log.Debugf("Body: %s", toJSON(maskedBody))
            req.Body = io.NopCloser(bytes.NewReader(body))
        }
    }

    // Execute request...
}
```

**Disabling Masking (Dangerous)**:

```bash
# Disable secret masking (ONLY for debugging)
myapp --no-mask-secrets --debug users list

⚠️  WARNING: Secret masking disabled
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Sensitive data may be exposed in output.
Only use this for debugging purposes.
DO NOT share output containing secrets.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Audit and Compliance

### Audit Logging

**Audit Log Configuration**:

```yaml
behaviors:
  audit:
    enabled: true

    # Audit log location (XDG-compliant)
    log_file: ~/.local/share/myapp/audit.log

    # Log format
    format: json  # json, text

    # What to log
    log_commands: true
    log_api_calls: true
    log_auth_events: true
    log_errors: true

    # Log retention
    retention_days: 90
    max_size_mb: 100
    compress: true
```

**Audit Log Entries**:

```json
{
  "timestamp": "2025-01-25T10:30:00Z",
  "event_type": "command_executed",
  "cli_version": "1.2.3",
  "user": "alice",
  "command": "users create",
  "flags": {
    "email": "user@example.com",
    "api-key": "sha256:a3f2b8c1..."  // Hashed, not plaintext
  },
  "api_endpoint": "POST /api/users",
  "status_code": 201,
  "duration_ms": 145,
  "error": null
}

{
  "timestamp": "2025-01-25T10:31:00Z",
  "event_type": "authentication",
  "auth_type": "oauth2",
  "auth_flow": "authorization_code",
  "token_acquired": true,
  "scopes": ["read:users", "write:users"],
  "token_hash": "sha256:f7e6d5c4...",
  "storage": "keyring"
}

{
  "timestamp": "2025-01-25T10:32:00Z",
  "event_type": "api_error",
  "api_endpoint": "GET /api/users/123",
  "status_code": 404,
  "error": "User not found",
  "retry_count": 0
}
```

**Audit Log Security**:

1. **File Permissions**:
   ```bash
   # Audit log created with restrictive permissions
   -rw------- 1 user user 12345 Jan 25 10:30 audit.log
   ```

2. **Secret Hashing**:
   - Secrets hashed (SHA-256), not logged in plaintext
   - Hash allows correlation without exposure
   - One-way hash (cannot reverse)

3. **Tamper Detection**:
   ```json
   {
     "entry_id": 12345,
     "entry_hash": "sha256:9a8b7c6d...",
     "previous_hash": "sha256:8b7c6d5e...",
     "signature": "..."  // Optional: signed audit entries
   }
   ```

4. **Log Rotation**:
   ```bash
   # Automatic log rotation
   audit.log           # Current
   audit.log.1.gz      # Compressed archive
   audit.log.2.gz
   audit.log.3.gz
   ```

**Audit Queries**:

```bash
# View recent audit events
myapp audit show --last 50

# Filter by event type
myapp audit show --type auth_events

# Filter by date range
myapp audit show --since 2025-01-01 --until 2025-01-31

# Export audit logs
myapp audit export --format csv --output audit-report.csv
```

### SOC2 Considerations

**SOC2 Trust Service Criteria** coverage:

#### Security (Common Criteria)

| Criteria | CliForge Feature | Implementation |
|----------|-----------------|----------------|
| **CC6.1**: Logical access controls | Authentication, authorization | OAuth2, API keys, scoped access |
| **CC6.2**: Transmission of data | TLS 1.2+, certificate validation | HTTPS only, no HTTP fallback |
| **CC6.3**: Encryption at rest | Keyring storage, encrypted files | OS keyring, AES-256 encryption |
| **CC6.6**: Access removal | Token revocation, logout | Explicit logout, token deletion |
| **CC6.7**: Credential management | Secret masking, rotation | Automatic masking, rotation support |
| **CC7.2**: System monitoring | Audit logging | Comprehensive audit logs |

#### Availability

| Criteria | CliForge Feature | Implementation |
|----------|-----------------|----------------|
| **A1.2**: Backup and recovery | Cached specs, fallback storage | Local caching, multi-tier storage |
| **A1.3**: Error handling | Retry logic, graceful degradation | Exponential backoff, user-friendly errors |

#### Confidentiality

| Criteria | CliForge Feature | Implementation |
|----------|-----------------|----------------|
| **C1.1**: Sensitive data handling | Secret detection, masking | Multi-layer detection, configurable masking |
| **C1.2**: Data disposal | Secure deletion | Overwrite tokens on logout |

**SOC2 Compliance Checklist**:

- ✅ All authentication events logged
- ✅ Secrets hashed in audit logs (not plaintext)
- ✅ TLS 1.2+ required for all network communication
- ✅ Certificate validation enforced
- ✅ Access tokens are time-limited
- ✅ Token refresh mechanism implemented
- ✅ Explicit logout/revocation supported
- ✅ Automatic secret masking in logs
- ✅ File permissions enforced (0600 for sensitive files)
- ✅ Keyring storage available (platform-dependent)
- ✅ Audit log retention configurable
- ✅ Error messages sanitized

### HIPAA Compliance

**HIPAA Requirements** (for CLIs handling PHI):

#### Administrative Safeguards

| Requirement | Implementation | Configuration |
|------------|----------------|---------------|
| **Access control** | Role-based scopes | OAuth2 scopes for PHI access |
| **Audit controls** | Comprehensive logging | Audit all PHI access |
| **Workforce training** | Documentation | Security guide, best practices |

#### Physical Safeguards

| Requirement | Implementation | Configuration |
|------------|----------------|---------------|
| **Device controls** | Token storage | Keyring (hardware-backed when available) |
| **Media disposal** | Secure deletion | Overwrite tokens, clear cache |

#### Technical Safeguards

| Requirement | Implementation | Configuration |
|------------|----------------|---------------|
| **Access control** | Authentication required | No anonymous access to PHI |
| **Audit controls** | Detailed logging | Log all PHI access with user ID |
| **Transmission security** | TLS 1.2+ | HTTPS only, no HTTP |
| **Encryption** | At-rest encryption | Keyring or encrypted file storage |

**HIPAA Configuration**:

```yaml
# HIPAA-compliant configuration
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      pkce: true
      scopes:
        - read:phi  # Explicit PHI scope
      storage:
        type: keyring  # Required for HIPAA

  audit:
    enabled: true
    log_commands: true
    log_api_calls: true
    log_auth_events: true
    retention_days: 2555  # 7 years (HIPAA requirement)

  secrets:
    enabled: true
    masking:
      style: full  # Full masking for PHI

  http:
    tls:
      min_version: "1.2"
      verify_certificates: true
```

**HIPAA Compliance Checklist**:

- ✅ Unique user authentication required
- ✅ Automatic session timeout (token expiration)
- ✅ Audit logs for all PHI access
- ✅ Audit log retention >= 6 years
- ✅ Encryption in transit (TLS 1.2+)
- ✅ Encryption at rest (keyring/file)
- ✅ Secure credential storage (no plaintext)
- ✅ PHI masked in logs and output
- ✅ Access control via scopes
- ❌ Backup and disaster recovery (API responsibility)
- ❌ Physical security (device responsibility)

### GDPR Compliance

**GDPR Requirements** (for CLIs processing EU personal data):

#### Data Protection Principles

| Principle | Implementation | Notes |
|-----------|----------------|-------|
| **Lawfulness, fairness, transparency** | Clear auth consent | OAuth2 consent screens |
| **Purpose limitation** | Scoped access | Minimal scopes requested |
| **Data minimization** | Collect only required data | API-defined data fields |
| **Accuracy** | Data validation | Schema validation |
| **Storage limitation** | Configurable retention | Cache TTL, log retention |
| **Integrity and confidentiality** | Encryption, TLS | At-rest and in-transit |
| **Accountability** | Audit logging | Comprehensive audit trail |

**GDPR Configuration**:

```yaml
behaviors:
  # Data minimization: cache only necessary data
  caching:
    spec_ttl: 5m        # Short TTL
    response_ttl: 0s    # No response caching (optional)

  # Audit logging: required for accountability
  audit:
    enabled: true
    retention_days: 90  # Adjust based on legal requirements

  # Data protection: encrypt sensitive data
  auth:
    storage:
      type: keyring     # Encrypted storage

  # Right to erasure: support data deletion
  data_retention:
    clear_cache_on_logout: true
    delete_logs_on_uninstall: true  # User action
```

**GDPR Rights Support**:

```bash
# Right to access: User can view their data
myapp auth status
myapp audit show --user alice

# Right to erasure: User can delete their data
myapp logout --clear-all   # Delete tokens
myapp cache clear          # Delete cached data
myapp audit clear          # Delete audit logs (optional, may conflict with retention requirements)

# Right to data portability: Export data
myapp audit export --format json --output my-data.json
```

**GDPR Compliance Checklist**:

- ✅ User consent for data processing (OAuth2 consent)
- ✅ Minimal data collection (scoped access)
- ✅ Data encrypted in transit (TLS 1.2+)
- ✅ Data encrypted at rest (keyring/file)
- ✅ Audit logging for accountability
- ✅ Data retention policies (configurable)
- ✅ User can access their data (audit logs)
- ✅ User can delete their data (logout, clear cache)
- ✅ User can export their data (audit export)
- ❌ Data processing agreements (API owner responsibility)
- ❌ Privacy policy (CLI developer responsibility)

---

## Security Best Practices

### For CLI Developers (API Owners)

#### 1. Secure Configuration

**Do**:
```yaml
# ✅ GOOD: Secure defaults
metadata:
  debug: false  # Production build

behaviors:
  auth:
    type: oauth2
    oauth2:
      pkce: true
      storage:
        type: keyring

  secrets:
    enabled: true
    masking:
      style: partial

  http:
    tls:
      min_version: "1.2"
      verify_certificates: true
```

**Don't**:
```yaml
# ❌ BAD: Insecure configuration
metadata:
  debug: true  # Allows config override

behaviors:
  auth:
    type: basic  # Weak authentication
    basic:
      username: admin  # Hardcoded credentials
      password: password123

  secrets:
    enabled: false  # No secret protection

  http:
    tls:
      verify_certificates: false  # Insecure
```

#### 2. Authentication Security

**OAuth2 Best Practices**:

1. **Always Use PKCE**:
   ```yaml
   oauth2:
     pkce: true  # REQUIRED for public clients
     client_secret: ""  # Not needed with PKCE
   ```

2. **Request Minimal Scopes**:
   ```yaml
   oauth2:
     scopes:
       - read:users  # Only what's needed
     # Don't request: admin, write:*, etc.
   ```

3. **Short-Lived Tokens**:
   ```yaml
   # Configure on auth server
   access_token_ttl: 15m   # 15 minutes
   refresh_token_ttl: 30d  # 30 days
   ```

4. **Token Rotation**:
   ```yaml
   # Configure on auth server
   rotate_refresh_tokens: true
   revoke_old_refresh_tokens: true
   ```

#### 3. Secret Management

**Never Embed Secrets**:

```yaml
# ❌ BAD: Secrets in config
behaviors:
  auth:
    api_key:
      key: "sk_live_abc123..."  # NEVER do this

# ✅ GOOD: Environment variables
behaviors:
  auth:
    api_key:
      env_var: MYAPP_API_KEY
```

**Enable Secret Detection**:

```yaml
# Build-time secret scanning
# Use tools like:
# - gitleaks
# - truffleHog
# - git-secrets

# Example: .github/workflows/security.yml
- name: Scan for secrets
  uses: trufflesecurity/trufflehog@main
  with:
    path: ./
    base: main
```

**Mark Secrets in OpenAPI Spec**:

```yaml
components:
  schemas:
    User:
      properties:
        api_key:
          type: string
          x-cli-secret: true  # Mark as secret
          description: "User API key (masked in CLI output)"
```

#### 4. TLS Configuration

**Require TLS 1.2+**:

```yaml
behaviors:
  http:
    tls:
      min_version: "1.2"
      strict_https: true  # Reject HTTP
```

**Consider Certificate Pinning**:

```yaml
# For high-security APIs
http:
  tls:
    pinned_public_keys:
      - sha256/AAAA...  # Intermediate CA public key
      - sha256/BBBB...  # Backup intermediate CA
```

#### 5. Audit Logging

**Enable Comprehensive Logging**:

```yaml
behaviors:
  audit:
    enabled: true
    log_commands: true
    log_api_calls: true
    log_auth_events: true
    log_errors: true
    retention_days: 90  # Adjust for compliance
```

**Secure Audit Logs**:

```yaml
audit:
  log_file: ~/.local/share/myapp/audit.log
  format: json
  permissions: "0600"  # Owner read/write only
  compress: true       # Compress rotated logs
```

#### 6. Error Handling

**Sanitize Error Messages**:

```go
// ❌ BAD: Exposes internal details
return fmt.Errorf("database connection failed: host=db.internal.example.com, user=admin, password=%s", password)

// ✅ GOOD: Generic error
return fmt.Errorf("authentication failed")
```

**Implement Rate Limiting**:

```yaml
behaviors:
  rate_limiting:
    enabled: true
    max_requests: 100
    window: 1m
    backoff:
      initial: 1s
      max: 60s
```

#### 7. Dependency Management

**Pin Dependencies**:

```go
// go.mod
require (
    github.com/spf13/cobra v1.8.0    // Exact version
    golang.org/x/oauth2 v0.15.0
)
```

**Regular Security Audits**:

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

#### 8. Build Security

**Reproducible Builds**:

```bash
# Use consistent build environment
cliforge build \
  --config cli-config.yaml \
  --version 1.2.3 \
  --ldflags "-s -w" \
  --reproducible
```

**Sign Releases**:

```bash
# Sign binaries
gpg --detach-sign --armor myapp-v1.2.3-darwin-amd64

# Generate checksums
sha256sum myapp-* > checksums.txt
gpg --clearsign checksums.txt
```

### For CLI Users

#### 1. Verify Binary Authenticity

**Check Checksums**:

```bash
# Download binary and checksums
curl -LO https://releases.example.com/myapp-v1.2.3-darwin-amd64
curl -LO https://releases.example.com/checksums.txt

# Verify checksum
sha256sum -c checksums.txt
```

**Verify Signatures**:

```bash
# Import developer's public key
gpg --import developer-public-key.asc

# Verify signature
gpg --verify checksums.txt.asc checksums.txt
```

#### 2. Secure Credential Storage

**Use Keyring Storage**:

```bash
# Verify keyring is available
myapp auth status --storage

# Output:
# Storage: keyring (macOS Keychain)
# Status: Available
```

**Protect Environment Variables**:

```bash
# ✅ GOOD: Use environment files with restricted permissions
cat > ~/.myapp.env <<EOF
export MYAPP_API_KEY=sk_live_...
EOF
chmod 600 ~/.myapp.env

# ❌ BAD: Credentials in shell history
export MYAPP_API_KEY=sk_live_abc123  # Visible in history
```

#### 3. Secure Configuration Files

**Set Restrictive Permissions**:

```bash
# Config file permissions
chmod 600 ~/.config/myapp/config.yaml

# Token file permissions (if using file storage)
chmod 600 ~/.local/share/myapp/token.json

# Cache directory permissions
chmod 700 ~/.cache/myapp
```

**Review Configuration**:

```bash
# View effective configuration
myapp config show

# Verify locked settings
myapp config validate
```

#### 4. Regular Updates

**Enable Update Checks**:

```yaml
# ~/.config/myapp/config.yaml
preferences:
  updates:
    auto_check: true
    check_interval: 24h
```

**Update Promptly**:

```bash
# Check for updates
myapp update check

# Update to latest version
myapp update install

# Verify update
myapp --version
```

#### 5. Minimal Scopes

**Review Requested Scopes**:

```bash
# During OAuth2 login, review scopes
myapp login

# Output:
# This application requests the following permissions:
#   - read:users   (View user information)
#   - write:users  (Create and update users)
#
# Only grant if these permissions are necessary.
```

**Revoke Unnecessary Access**:

```bash
# Logout to revoke tokens
myapp logout

# Re-login with minimal scopes
myapp login --scopes read:users
```

#### 6. Monitor Activity

**Review Audit Logs**:

```bash
# View recent activity
myapp audit show --last 50

# Look for suspicious activity
myapp audit show --type auth_events
```

**Check Token Status**:

```bash
# View current authentication
myapp auth status

# Output:
# Status: Authenticated
# Token expires: 2025-01-25 12:00:00 UTC (in 45 minutes)
# Scopes: read:users, write:users
# Storage: keyring
```

#### 7. Secure Deletion

**Logout When Done**:

```bash
# Logout and clear tokens
myapp logout

# Clear all cached data
myapp cache clear

# Clear audit logs (if appropriate)
myapp audit clear --confirm
```

**Uninstall Cleanly**:

```bash
# Remove configuration
rm -rf ~/.config/myapp

# Remove cached data
rm -rf ~/.cache/myapp

# Remove data files
rm -rf ~/.local/share/myapp

# Remove keyring entries (manual)
# macOS: Keychain Access > search "myapp" > delete
# Linux: seahorse > search "myapp" > delete
# Windows: Credential Manager > search "myapp" > delete
```

### For Operations Teams

#### 1. Centralized Secrets Management

**Use Secrets Management Systems**:

```bash
# HashiCorp Vault
export MYAPP_API_KEY=$(vault kv get -field=api_key secret/myapp)

# AWS Secrets Manager
export MYAPP_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id myapp-api-key \
  --query SecretString \
  --output text)

# Azure Key Vault
export MYAPP_API_KEY=$(az keyvault secret show \
  --vault-name myapp-vault \
  --name api-key \
  --query value -o tsv)
```

**Rotate Credentials Regularly**:

```bash
# Automated credential rotation (example)
#!/bin/bash
# rotate-credentials.sh

# Generate new API key
NEW_KEY=$(myapp auth create-key)

# Update secrets manager
vault kv put secret/myapp api_key=$NEW_KEY

# Verify new key works
export MYAPP_API_KEY=$NEW_KEY
myapp auth verify

# Revoke old key (after verification)
myapp auth revoke-old-keys
```

#### 2. Network Security

**Restrict Network Access**:

```bash
# Firewall rules (allow HTTPS to API only)
# iptables example
iptables -A OUTPUT -p tcp --dport 443 -d api.example.com -j ACCEPT
iptables -A OUTPUT -p tcp --dport 443 -j DROP
```

**Use Dedicated Service Accounts**:

```yaml
# Service account OAuth2 config
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: client_credentials
      client_id: myapp-production-service
      client_secret: ${PROD_CLIENT_SECRET}
      scopes:
        - api:read  # Minimal scopes
```

#### 3. Compliance Monitoring

**Audit Log Aggregation**:

```bash
# Centralize audit logs
# Example: Ship to SIEM system

# Fluentd/Fluent Bit
<source>
  @type tail
  path /home/*/.local/share/myapp/audit.log
  format json
  tag myapp.audit
</source>

<match myapp.audit>
  @type elasticsearch
  host elasticsearch.internal
  port 9200
  index_name myapp-audit
</match>
```

**Compliance Reporting**:

```bash
# Generate compliance reports
myapp audit export \
  --format csv \
  --since 2025-01-01 \
  --until 2025-01-31 \
  --output audit-jan-2025.csv

# Filter for specific events
myapp audit export \
  --type auth_events \
  --format json \
  --output auth-events.json
```

#### 4. Incident Response

**Security Incident Playbook**:

1. **Credential Compromise**:
   ```bash
   # Immediately revoke compromised credentials
   myapp auth revoke --token <compromised-token>

   # Rotate all credentials
   ./rotate-credentials.sh

   # Review audit logs for unauthorized access
   myapp audit show --since <compromise-time> --type api_calls

   # Notify affected users
   # ...
   ```

2. **Binary Compromise**:
   ```bash
   # Verify binary integrity
   sha256sum myapp

   # Re-download from trusted source
   curl -LO https://releases.example.com/myapp-v1.2.3-darwin-amd64

   # Verify signature
   gpg --verify myapp-v1.2.3.sig myapp-v1.2.3-darwin-amd64
   ```

3. **API Compromise**:
   ```bash
   # Switch to backup API endpoint (if configured)
   export MYAPP_API_URL=https://backup-api.example.com

   # Or wait for API recovery
   # ...
   ```

#### 5. Security Monitoring

**Alert on Anomalies**:

```yaml
# Example: Prometheus alerting rules
groups:
  - name: myapp_security
    rules:
      - alert: HighAuthFailureRate
        expr: rate(myapp_auth_failures[5m]) > 0.1
        annotations:
          summary: "High authentication failure rate"

      - alert: UnusualAPIUsage
        expr: rate(myapp_api_calls[5m]) > 100
        annotations:
          summary: "Unusual API call volume"
```

**Dashboards**:

```bash
# Grafana dashboard (example)
# - Authentication events timeline
# - API call volume by endpoint
# - Error rate by type
# - Token expiration timeline
```

---

## Vulnerability Management

### Reporting Security Issues

**Security Vulnerability Disclosure**:

CliForge follows responsible disclosure principles. If you discover a security vulnerability, please report it privately.

**Reporting Channel**:

```
Email: security@cliforge.dev
PGP Key: https://cliforge.dev/security-pgp-key.asc
```

**What to Include in Report**:

1. **Vulnerability Description**:
   - Type of vulnerability (e.g., credential leakage, MITM)
   - Affected versions
   - Impact assessment

2. **Reproduction Steps**:
   - Detailed steps to reproduce
   - Configuration files (sanitized)
   - Commands executed
   - Expected vs actual behavior

3. **Proof of Concept**:
   - Code/scripts demonstrating the issue
   - Screenshots (if applicable)
   - Logs (with secrets masked)

4. **Suggested Fix** (if known):
   - Proposed solution
   - Patch/code changes
   - Configuration changes

**Example Report**:

```
Subject: [SECURITY] Credential leakage in debug output

Severity: HIGH
Affected Versions: v1.0.0 - v1.2.3

Description:
When using --debug flag, API keys are logged in plaintext to stderr,
even when secret masking is enabled.

Reproduction:
1. Configure API key authentication:
   behaviors.auth.type: api_key
   behaviors.secrets.enabled: true
2. Run: myapp --debug users list
3. Observe: API key visible in debug output

Expected: API key should be masked (e.g., "sk_liv***")
Actual: API key logged in plaintext (e.g., "sk_live_abc123...")

Impact:
- API keys exposed in debug logs
- Keys may be shared inadvertently (e.g., in bug reports)
- Violates principle of least privilege

Suggested Fix:
Apply secret masking to debug output, not just regular output.

Proof of Concept:
[Attached: debug-output.txt with keys redacted]
```

**Response Timeline**:

- **Initial Response**: Within 48 hours
- **Triage**: Within 5 business days
- **Fix Development**: Depends on severity
  - **Critical**: < 7 days
  - **High**: < 14 days
  - **Medium**: < 30 days
  - **Low**: Next release
- **Public Disclosure**: After fix released (coordinated disclosure)

### Security Update Process

**Severity Classification**:

| Severity | Criteria | Example | Response Time |
|----------|----------|---------|---------------|
| **Critical** | Remote code execution, credential theft | Hardcoded API key in binary | < 7 days |
| **High** | Privilege escalation, data exposure | Secret masking bypass | < 14 days |
| **Medium** | Denial of service, information disclosure | TLS downgrade attack | < 30 days |
| **Low** | Minor information leakage | Version info disclosure | Next release |

**Security Update Workflow**:

1. **Vulnerability Reported**:
   - Security team notified
   - Issue triaged and classified
   - CVE requested (if applicable)

2. **Fix Developed**:
   - Patch created on private branch
   - Security review conducted
   - Testing (unit, integration, security)

3. **Release Prepared**:
   - Version bump (patch for security fix)
   - Changelog entry (with CVE if applicable)
   - Binaries built and signed

4. **Release Published**:
   - Binaries uploaded to release server
   - Security advisory published
   - Users notified (update mechanism)

5. **Public Disclosure**:
   - Coordinated disclosure (14-30 days after fix)
   - CVE published
   - Blog post / announcement

**Security Advisory Format**:

```markdown
# Security Advisory: CVE-2025-XXXXX

**Severity**: HIGH
**Affected Versions**: v1.0.0 - v1.2.3
**Fixed in Version**: v1.2.4
**Published**: 2025-01-25

## Summary

A vulnerability in CliForge's secret masking implementation allows
API keys to be logged in plaintext when using the --debug flag.

## Impact

Attackers with access to debug logs can extract API keys, leading
to unauthorized API access.

## Affected Configurations

All configurations using API key authentication are affected.

## Mitigation

Upgrade to v1.2.4 or later immediately.

## Workaround

Avoid using --debug flag until upgraded.

## Timeline

- 2025-01-10: Vulnerability reported
- 2025-01-11: Fix developed
- 2025-01-15: Release v1.2.4
- 2025-01-25: Public disclosure

## Credits

Reported by: Alice Smith (@alice)
```

**Automatic Update Notifications**:

```bash
# CLI checks for security updates
myapp --version

⚠️  SECURITY UPDATE AVAILABLE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Current version: v1.2.3
Latest version:  v1.2.4 (security fix)

A critical security vulnerability has been fixed.
Please update immediately:

  myapp update install

Details: https://cliforge.dev/security/CVE-2025-XXXXX
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### CVE Tracking

**CVE Request Process**:

1. **Determine if CVE Needed**:
   - Publicly disclosed vulnerability
   - Affects released versions
   - Security impact (not just bug)

2. **Request CVE**:
   - Submit to MITRE via web form
   - Provide vulnerability details
   - Receive CVE ID

3. **CVE Entry**:
   ```
   CVE-2025-XXXXX

   Description:
   CliForge before 1.2.4 logs API keys in plaintext when using
   the --debug flag, allowing attackers with access to debug logs
   to extract credentials.

   Severity: HIGH (CVSS 7.5)

   Affected Versions: 1.0.0 - 1.2.3
   Fixed in: 1.2.4

   References:
   - https://cliforge.dev/security/CVE-2025-XXXXX
   - https://github.com/CliForge/cliforge/security/advisories/GHSA-xxxx-xxxx-xxxx
   ```

4. **CVE Database Updates**:
   - National Vulnerability Database (NVD)
   - GitHub Advisory Database
   - CVE.org

**CVSS Scoring**:

```
CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N

Attack Vector (AV): Network
Attack Complexity (AC): Low
Privileges Required (PR): None
User Interaction (UI): None
Scope (S): Unchanged
Confidentiality (C): High
Integrity (I): None
Availability (A): None

Score: 7.5 (HIGH)
```

**CVE Tracking in CLI**:

```bash
# Check for known CVEs
myapp security check

# Output:
# Checking for known vulnerabilities...
#
# ✓ No known CVEs for version v1.2.4
#
# Recent security advisories:
# - CVE-2025-XXXXX (HIGH) - Fixed in v1.2.4
#   API keys logged in debug output
```

---

## Security Checklist

### Pre-Release Security Checklist

**Before releasing a new CLI version**:

#### Build Security

- [ ] Debug mode disabled (`metadata.debug: false`)
- [ ] No embedded secrets detected
- [ ] Dependencies scanned for vulnerabilities
- [ ] Reproducible build verified
- [ ] Binaries signed with release key
- [ ] Checksums generated and signed

#### Configuration Security

- [ ] TLS 1.2+ required
- [ ] Certificate verification enabled
- [ ] HTTPS-only endpoints (no HTTP)
- [ ] Secret masking enabled
- [ ] Audit logging configured
- [ ] Secure storage defaults (keyring)

#### Authentication Security

- [ ] OAuth2 with PKCE enabled (if applicable)
- [ ] No hardcoded credentials
- [ ] Token storage secured (keyring/encrypted file)
- [ ] Token expiration configured
- [ ] Token refresh implemented

#### Code Security

- [ ] No sensitive data in logs
- [ ] Error messages sanitized
- [ ] Input validation implemented
- [ ] Output encoding applied
- [ ] Rate limiting configured

#### Documentation

- [ ] Security guide updated
- [ ] Threat model documented
- [ ] Security advisories published
- [ ] Incident response plan ready

### Deployment Security Checklist

**For organizations deploying CliForge-generated CLIs**:

#### Infrastructure

- [ ] Binaries downloaded from trusted sources
- [ ] Binary signatures verified
- [ ] Checksums validated
- [ ] Network access restricted (firewall rules)
- [ ] Proxy configured (if required)

#### Credential Management

- [ ] Secrets stored in secrets management system
- [ ] Credentials rotated regularly (90 days)
- [ ] Service accounts used (not personal accounts)
- [ ] Minimal scopes requested
- [ ] Token expiration enforced

#### Monitoring

- [ ] Audit logs collected and monitored
- [ ] Alerts configured for anomalies
- [ ] Compliance reports generated
- [ ] Security dashboards created

#### Compliance

- [ ] SOC2 requirements met (if applicable)
- [ ] HIPAA requirements met (if applicable)
- [ ] GDPR requirements met (if applicable)
- [ ] Data retention policies enforced
- [ ] Privacy policies reviewed

#### Incident Response

- [ ] Incident response plan documented
- [ ] Security contacts identified
- [ ] Escalation procedures defined
- [ ] Backup/recovery tested

### Runtime Security Checklist

**For end users running CliForge-generated CLIs**:

#### Initial Setup

- [ ] Binary downloaded from official source
- [ ] Signature/checksum verified
- [ ] CLI installed in user directory (not system-wide)
- [ ] Configuration file permissions set (0600)
- [ ] Keyring available and configured

#### Regular Usage

- [ ] Latest version installed
- [ ] Update checks enabled
- [ ] Audit logs reviewed periodically
- [ ] Unnecessary cached data cleared
- [ ] Token status checked (`myapp auth status`)

#### Security Hygiene

- [ ] Credentials not in shell history
- [ ] Environment files have restrictive permissions (0600)
- [ ] Logout when done (`myapp logout`)
- [ ] Debug mode only used when necessary
- [ ] Secrets never shared or logged

#### Incident Response

- [ ] Know how to revoke tokens (`myapp logout`)
- [ ] Know how to clear cached data (`myapp cache clear`)
- [ ] Know how to report security issues (email security team)
- [ ] Keep audit logs for investigation

---

## Conclusion

CliForge provides a comprehensive security framework for building and operating secure CLI applications. By following the guidelines in this document, CLI developers, API owners, operations teams, and end users can work together to maintain a strong security posture.

### Key Takeaways

1. **Defense in Depth**: Security is implemented at multiple layers (build, distribution, runtime)
2. **Secure by Default**: Security features are enabled by default, not opt-in
3. **User Control**: Users maintain control over their credentials and data
4. **Transparency**: Clear security boundaries and trust relationships
5. **Continuous Protection**: Ongoing security updates and monitoring

### Additional Resources

- **CliForge Documentation**: https://docs.cliforge.dev
- **Security Advisories**: https://cliforge.dev/security
- **Security Contact**: security@cliforge.dev
- **GitHub Security**: https://github.com/CliForge/cliforge/security

### Updates to This Guide

This guide is updated with each major release:

- **v1.0.0** (2025-01-25): Initial comprehensive security guide
- Future updates will be listed here

For the latest version of this guide, visit: https://docs.cliforge.dev/security-guide

---

**Document Version**: 1.0.0
**Last Updated**: 2025-01-25
**Maintained by**: CliForge Security Team
**License**: CC BY-SA 4.0
