# Secrets & Sensitive Data Handling

## Overview

CliForge needs to prevent sensitive data from being logged to stdout/stderr/logfiles. This includes API keys, passwords, tokens, and sensitive API response fields.

---

## Problem Statement

Sensitive data can leak through:
1. **Logs**: Debug output showing API requests/responses
2. **Error messages**: Stack traces with credential data
3. **Command output**: API responses containing secrets
4. **Audit logs**: Recording commands with sensitive flags

**Example of what we want to prevent:**
```bash
$ mycli users create --api-key sk_live_abc123 --email user@example.com --debug
DEBUG: POST /users
DEBUG: Headers: {"X-API-Key": "sk_live_abc123"}  ❌ LEAKED!
DEBUG: Response: {
  "id": "usr_123",
  "email": "user@example.com",
  "api_key": "sk_live_abc123"  ❌ LEAKED!
}
```

---

## Solution: Multi-Layer Secret Detection

### 1. OpenAPI Schema Marking (Primary)

Mark sensitive fields in the OpenAPI spec using `x-cli-secret`:

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: string
        email:
          type: string
        api_key:
          type: string
          x-cli-secret: true  # Mark as secret
        password:
          type: string
          writeOnly: true     # OpenAPI standard for passwords
          x-cli-secret: true

    CreateUserRequest:
      type: object
      properties:
        email:
          type: string
        password:
          type: string
          x-cli-secret: true

# Mark sensitive parameters
paths:
  /users:
    post:
      parameters:
        - name: X-API-Key
          in: header
          schema:
            type: string
          x-cli-secret: true  # Mark header as secret
```

### 2. Pattern-Based Detection (Secondary)

Automatically detect fields matching common secret patterns:

**Field name patterns** (case-insensitive):
```yaml
behaviors:
  secrets:
    # Field name patterns (case-insensitive, supports wildcards)
    field_patterns:
      - "*password*"
      - "*secret*"
      - "*token*"
      - "*key"        # api_key, access_key, etc.
      - "*credential*"
      - "auth"
      - "authorization"
      - "*bearer*"
      - "session"
      - "*ssn*"       # Social security number
      - "*cc_number*" # Credit card
      - "*cvv*"
      - "*pin*"
```

**Value patterns** (regex for detecting secret-like values):
```yaml
behaviors:
  secrets:
    # Value patterns (regex)
    value_patterns:
      - name: "AWS Access Key"
        pattern: 'AKIA[0-9A-Z]{16}'
      - name: "AWS Secret Key"
        pattern: '[A-Za-z0-9/+=]{40}'
      - name: "Generic API Key"
        pattern: '[sS][kK]_[a-zA-Z0-9]{32,}'
      - name: "Bearer Token"
        pattern: 'Bearer [A-Za-z0-9\-._~+/]+=*'
      - name: "JWT"
        pattern: 'eyJ[A-Za-z0-9-_=]+\.eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_.+/=]*'
      - name: "UUID (potential session ID)"
        pattern: '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
        enabled: false  # Too many false positives, disabled by default
```

### 3. Configuration DSL

```yaml
behaviors:
  secrets:
    # Enable/disable secret masking
    enabled: true

    # Masking strategy
    masking:
      style: partial  # partial, full, hash
      # partial: "sk_live_abc123..." -> "sk_liv***"
      # full: "sk_live_abc123..." -> "***"
      # hash: "sk_live_abc123..." -> "sha256:a3f2..."

      partial_show_chars: 6  # Show first 6 chars for partial masking
      replacement: "***"     # Replacement string

    # Field name patterns (auto-detect secrets)
    field_patterns:
      - "*password*"
      - "*secret*"
      - "*token*"
      - "*key"
      - "*credential*"
      - "auth*"
      - "*bearer*"

    # Value patterns (regex for detecting secret-like values)
    value_patterns:
      - name: "AWS Access Key"
        pattern: 'AKIA[0-9A-Z]{16}'
        enabled: true
      - name: "Generic API Key"
        pattern: '[sS][kK]_[a-zA-Z0-9]{32,}'
        enabled: true
      - name: "Bearer Token"
        pattern: 'Bearer [A-Za-z0-9\-._~+/]+=*'
        enabled: true

    # Explicit field paths to mask (JSONPath-like syntax)
    explicit_fields:
      - "$.user.api_key"
      - "$.user.password"
      - "$.credentials.*"
      - "$.*.secret"
      - "$.auth.token"

    # Headers to mask
    headers:
      - "Authorization"
      - "X-API-Key"
      - "X-Auth-Token"
      - "Cookie"
      - "Set-Cookie"

    # Where to apply masking
    mask_in:
      stdout: true      # Mask in command output
      stderr: true      # Mask in error messages
      logs: true        # Mask in log files
      audit: false      # Don't mask in audit logs (secure storage)
      debug_output: true  # Mask even in --debug output
```

---

## Masking Examples

### Partial Masking (Default)

```bash
$ mycli users create --api-key sk_live_abc123def456 --debug

DEBUG: POST /users
DEBUG: Headers: {"X-API-Key": "sk_liv***"}
DEBUG: Request Body: {
  "email": "user@example.com",
  "password": "super***"
}
DEBUG: Response: {
  "id": "usr_123",
  "email": "user@example.com",
  "api_key": "sk_liv***",
  "secret_token": "eyJhbG***"
}

✓ User created successfully
```

### Full Masking

```yaml
behaviors:
  secrets:
    masking:
      style: full
```

```bash
DEBUG: Headers: {"X-API-Key": "***"}
DEBUG: Response: {
  "api_key": "***",
  "password": "***"
}
```

### Hash Masking (for audit trails)

```yaml
behaviors:
  secrets:
    masking:
      style: hash
```

```bash
DEBUG: Headers: {"X-API-Key": "sha256:a3f2b8c1..."}
```

---

## Implementation

### Secret Detection Algorithm

```go
package secrets

import (
    "regexp"
    "strings"
)

type SecretDetector struct {
    config *SecretsConfig
    fieldPatterns []*regexp.Regexp
    valuePatterns []*regexp.Regexp
}

// IsSecretField checks if a field name indicates a secret
func (d *SecretDetector) IsSecretField(fieldName string) bool {
    lowerField := strings.ToLower(fieldName)

    // Check configured field patterns
    for _, pattern := range d.config.FieldPatterns {
        if matched, _ := filepath.Match(pattern, lowerField); matched {
            return true
        }
    }

    return false
}

// IsSecretValue checks if a value looks like a secret
func (d *SecretDetector) IsSecretValue(value string) (bool, string) {
    for _, pattern := range d.config.ValuePatterns {
        if !pattern.Enabled {
            continue
        }

        if pattern.Regex.MatchString(value) {
            return true, pattern.Name
        }
    }

    return false, ""
}

// MaskValue masks a sensitive value
func (d *SecretDetector) MaskValue(value string) string {
    switch d.config.Masking.Style {
    case "full":
        return d.config.Masking.Replacement
    case "hash":
        return fmt.Sprintf("sha256:%s", sha256Hash(value)[:16])
    case "partial":
        return partialMask(value, d.config.Masking.PartialShowChars, d.config.Masking.Replacement)
    default:
        return d.config.Masking.Replacement
    }
}

func partialMask(value string, showChars int, replacement string) string {
    if len(value) <= showChars {
        return replacement
    }
    return value[:showChars] + replacement
}

// MaskJSON recursively masks secrets in JSON
func (d *SecretDetector) MaskJSON(data map[string]interface{}) map[string]interface{} {
    result := make(map[string]interface{})

    for key, value := range data {
        // Check if field name indicates a secret
        if d.IsSecretField(key) {
            result[key] = d.MaskValue(fmt.Sprint(value))
            continue
        }

        // Check if value looks like a secret
        if strValue, ok := value.(string); ok {
            if isSecret, _ := d.IsSecretValue(strValue); isSecret {
                result[key] = d.MaskValue(strValue)
                continue
            }
        }

        // Recurse for nested objects
        if nested, ok := value.(map[string]interface{}); ok {
            result[key] = d.MaskJSON(nested)
            continue
        }

        // Recurse for arrays
        if arr, ok := value.([]interface{}); ok {
            result[key] = d.maskArray(arr)
            continue
        }

        result[key] = value
    }

    return result
}
```

### Integration with Logging

```go
// HTTP client wrapper that masks secrets
func (c *Client) Do(req *http.Request) (*http.Response, error) {
    if c.debug {
        // Mask secrets in debug output
        maskedReq := c.secretDetector.MaskRequest(req)
        log.Debugf("Request: %s %s", maskedReq.Method, maskedReq.URL)
        log.Debugf("Headers: %v", maskedReq.Header)
        if req.Body != nil {
            body, _ := io.ReadAll(req.Body)
            maskedBody := c.secretDetector.MaskJSON(parseJSON(body))
            log.Debugf("Body: %s", toJSON(maskedBody))
            req.Body = io.NopCloser(bytes.NewReader(body))
        }
    }

    resp, err := c.httpClient.Do(req)

    if c.debug && resp != nil {
        body, _ := io.ReadAll(resp.Body)
        maskedBody := c.secretDetector.MaskJSON(parseJSON(body))
        log.Debugf("Response: %s", toJSON(maskedBody))
        resp.Body = io.NopCloser(bytes.NewReader(body))
    }

    return resp, err
}
```

---

## OpenAPI Extension: `x-cli-secret`

### Schema-Level

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        api_key:
          type: string
          x-cli-secret: true
          description: "User's API key (will be masked in output)"
```

### Parameter-Level

```yaml
paths:
  /auth/login:
    post:
      parameters:
        - name: password
          in: query
          schema:
            type: string
          x-cli-secret: true
```

### Header-Level

```yaml
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
      x-cli-secret: true  # Mask in logs
```

---

## User Controls

Users can:

**1. Disable masking** (dangerous, but useful for debugging):
```bash
$ mycli users list --no-mask-secrets --debug
Warning: Secret masking disabled. Sensitive data may be exposed!
```

**2. Configure masking level:**
```yaml
# ~/.config/mycli/config.yaml
default:
  secrets:
    masking:
      style: partial  # or full, hash
```

**3. Add custom secret patterns:**
```yaml
# ~/.config/mycli/config.yaml
default:
  secrets:
    field_patterns:
      - "*internal_id*"  # Company-specific pattern
    value_patterns:
      - name: "Company Employee ID"
        pattern: 'EMP[0-9]{6}'
```

---

## Audit Logging

**Audit logs store HASHED secrets** (for compliance):

```yaml
# ~/.local/share/mycli/audit.log
timestamp: 2025-01-11T15:30:00Z
command: users create
flags:
  email: user@example.com
  api-key: sha256:a3f2b8c1d5e6f7a8...  # Hashed, not plaintext
response:
  user_id: usr_123
  api_key: sha256:f7e6d5c4b3a2f1e0...  # Hashed
status: success
```

This allows:
- Detecting if same API key was used multiple times
- Compliance/audit trails
- No plaintext secrets

---

## Best Practices

### For API Developers

1. **Mark secrets in OpenAPI spec:**
   ```yaml
   x-cli-secret: true
   ```

2. **Use `writeOnly` for passwords:**
   ```yaml
   password:
     type: string
     writeOnly: true
     x-cli-secret: true
   ```

3. **Document what's secret:**
   ```yaml
   description: "User's API key (will be masked in CLI output)"
   ```

### For CLI Users

1. **Never disable masking in production**
2. **Use `--debug` sparingly** (even masked, it shows structure)
3. **Review audit logs** for unexpected secret usage
4. **Use environment variables** instead of CLI flags for secrets:
   ```bash
   export MYCLI_API_KEY=sk_live_...
   mycli users list  # No secret in command history
   ```

---

## Summary

**CliForge provides three layers of secret protection:**

1. **Explicit marking**: `x-cli-secret` in OpenAPI spec
2. **Pattern detection**: Automatic detection via field names and value patterns
3. **User configuration**: Custom patterns and masking preferences

**Secrets are masked in:**
- ✅ Command output (stdout)
- ✅ Error messages (stderr)
- ✅ Debug logs
- ✅ Log files
- ✅ Audit logs (hashed, not plaintext)

**Users retain control:**
- Can disable masking (with warnings)
- Can customize masking style (partial, full, hash)
- Can add custom secret patterns

This prevents accidental secret leaks while maintaining usability.

---

**Version**: 0.7.0
**Last Updated**: 2025-01-11
**Project**: CliForge - Forge CLIs from APIs
