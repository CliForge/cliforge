# File Operations Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Proposed

---

## Overview

### Problem Statement

ROSA CLI and similar tools accept file inputs for:
- Certificates (PEM, X.509)
- Authentication credentials (htpasswd)
- Configuration files (YAML, JSON)
- Trust bundles
- CloudFormation templates

CliForge's current design only handles primitive types in request bodies. File inputs require parsing, validation, and transformation before sending to APIs.

### Solution

A file operations framework that:
1. Handles file reading via `--from-file` and `--<param>-file` flags
2. Validates file formats and content
3. Transforms files to API-compatible formats
4. Integrates with workflow engine and plugin system

---

## Architecture

### File Input Flow

```
User provides file flag
      │
      ▼
┌─────────────────┐
│  Flag Parser    │ ── Detect file flag (--ca-file, --from-file)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  File Reader    │ ── Read file from filesystem
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Format Detector│ ── Detect format (PEM, JSON, YAML, htpasswd, etc.)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Validator      │ ── Validate content
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Transformer    │ ── Transform to API format
└────────┬────────┘
         │
         ▼
Include in API request body
```

---

## OpenAPI Extension: `x-cli-file-input`

**Location**: Request body property or parameter

**Purpose**: Declare that a field accepts file input

```yaml
requestBody:
  content:
    application/json:
      schema:
        properties:
          ca_certificate:
            type: string
            format: byte  # base64-encoded
            x-cli-file-input:
              flag: --ca-file
              format: pem
              type: x509-certificate
              validation:
                - type: pem-structure
                - type: certificate-validity
              transform: base64-encode

          htpasswd_users:
            type: array
            items:
              type: object
            x-cli-file-input:
              flag: --from-file
              format: htpasswd
              transform: parse-htpasswd-to-json
```

---

## Supported File Formats

### 1. PEM (Privacy-Enhanced Mail)

**Use Cases**: Certificates, private keys, certificate chains

**Validation**:
- PEM structure (BEGIN/END markers)
- Certificate validity (expiration, signature)
- Key pair matching

**Transformation**:
- Extract certificate data
- Base64 encode
- Convert to DER format

**Example**:
```yaml
x-cli-file-input:
  flag: --ca-file
  format: pem
  type: x509-certificate
  transform: base64-encode
```

### 2. htpasswd

**Use Cases**: HTTP basic auth user database

**Validation**:
- htpasswd format (username:hashed_password)
- Supported hash types (bcrypt, apr1)

**Transformation**:
- Parse to JSON array of users
- Extract usernames and hashes

**Example**:
```yaml
x-cli-file-input:
  flag: --from-file
  format: htpasswd
  transform: parse-htpasswd-to-json
  output-schema:
    type: array
    items:
      properties:
        username: string
        password: string
```

### 3. JSON/YAML

**Use Cases**: Configuration files, complex inputs

**Validation**:
- Valid JSON/YAML syntax
- Schema validation against OpenAPI schema

**Transformation**:
- Parse to object
- Merge with other flags

**Example**:
```yaml
x-cli-file-input:
  flag: --config-file
  format: [json, yaml]
  validation:
    - type: schema
      schema-ref: "#/components/schemas/ClusterConfig"
```

### 4. Plain Text

**Use Cases**: Scripts, templates, arbitrary content

**Validation**:
- File size limits
- Character encoding (UTF-8)

**Transformation**:
- Read as string
- Base64 encode if binary

---

## Security Considerations

### File Access Control

**Restrictions**:
- Only read files from allowed directories
- Respect file permissions
- No symlink traversal outside allowed paths

### Size Limits

**Defaults**:
- Certificate files: 10KB max
- Config files: 1MB max
- User can override with `--max-file-size`

### Sensitive Data

**Handling**:
- Private keys trigger security warnings
- Secrets masked in debug output
- Integrate with secrets-handling-design.md

---

## Integration with Plugins

File operations implemented as built-in plugin:

```yaml
x-cli-workflow:
  steps:
    - id: parse-certificate
      type: plugin
      plugin: file-ops
      input:
        operation: parse
        file: "{flags.ca_cert_file}"
        format: pem
        type: x509-certificate
      output:
        cert_data: "{result.base64}"
        issuer: "{result.issuer}"
        expiry: "{result.not_after}"

    - id: validate-cert
      type: plugin
      plugin: file-ops
      input:
        operation: validate
        data: "{parse-certificate.cert_data}"
        checks:
          - not-expired
          - trusted-issuer

    - id: upload-cert
      type: api-call
      endpoint: /api/v1/certificates
      body:
        certificate: "{parse-certificate.cert_data}"
```

---

## Related Documents

- **Plugin Architecture**: `plugin-architecture.md`
- **Workflow Orchestration**: `workflow-orchestration.md`
- **Secrets Handling**: `secrets-handling-design.md`
- **Gap Analysis**: `gap-analysis-rosa-requirements.md`

---

*⚒️ Forged with ❤️ by the CliForge team*
