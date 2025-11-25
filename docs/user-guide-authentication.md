# CliForge Authentication Guide

**Version:** 0.9.0
**Last Updated:** 2025-01-23

## Table of Contents

1. [Introduction](#introduction)
2. [Supported Authentication Types](#supported-authentication-types)
3. [API Key Authentication](#api-key-authentication)
4. [Basic Authentication](#basic-authentication)
5. [OAuth2 Authentication](#oauth2-authentication)
6. [Token Storage](#token-storage)
7. [Token Refresh](#token-refresh)
8. [Multiple Auth Providers](#multiple-auth-providers)
9. [Environment Variables for Credentials](#environment-variables-for-credentials)
10. [Security Best Practices](#security-best-practices)
11. [Troubleshooting](#troubleshooting)

---

## Introduction

CliForge provides comprehensive authentication support for CLI applications, including API keys, Basic authentication, and OAuth2 with multiple flow types. This guide explains how to configure and use each authentication method.

### Key Concepts

- **Authenticator**: Handles the authentication flow and token generation
- **Token Storage**: Persists authentication tokens securely
- **Token Refresh**: Automatically refreshes expired tokens
- **Multi-Provider**: Support for multiple authentication providers
- **Secure Defaults**: Environment variables for sensitive credentials

---

## Supported Authentication Types

CliForge supports four authentication types:

| Type | Description | Use Cases | Security |
|------|-------------|-----------|----------|
| **API Key** | Static key sent in header or query | Simple APIs, service accounts | Medium |
| **Basic** | Username/password in Authorization header | Legacy APIs, development | Low |
| **OAuth2** | Industry-standard token-based auth | Modern APIs, user auth | High |
| **None** | No authentication | Public APIs, testing | N/A |

### Choosing an Authentication Type

```
┌─────────────────────────────────────────────────────────────┐
│  Is the API public?                                         │
│  └─ Yes → Use "none"                                        │
│  └─ No  → Continue                                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Does the API support OAuth2?                               │
│  └─ Yes → Use "oauth2" (recommended)                        │
│  └─ No  → Continue                                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Does the API use API keys?                                 │
│  └─ Yes → Use "api_key"                                     │
│  └─ No  → Use "basic" (least secure)                        │
└─────────────────────────────────────────────────────────────┘
```

---

## API Key Authentication

API key authentication sends a static key with each request, either in a header or query parameter.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      location: header              # header or query
      name: X-API-Key               # Header name or query param name
      prefix: ""                    # Optional prefix (e.g., "Bearer ")
      env_var: PETSTORE_API_KEY     # Environment variable to read from
```

### Header-Based API Key

Most common pattern for API key authentication:

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      location: header
      name: X-API-Key
      env_var: PETSTORE_API_KEY
```

**Usage**:

```bash
# Set API key via environment variable
export PETSTORE_API_KEY=your-api-key-here

# CLI automatically adds header:
# X-API-Key: your-api-key-here
petstore pets list
```

### Query Parameter API Key

For APIs that require API key in URL:

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      location: query
      name: api_key
      env_var: PETSTORE_API_KEY
```

**Usage**:

```bash
export PETSTORE_API_KEY=your-api-key-here

# CLI automatically adds query parameter:
# https://api.example.com/pets?api_key=your-api-key-here
petstore pets list
```

### API Key with Prefix

For APIs that require a prefix (e.g., Bearer tokens):

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      location: header
      name: Authorization
      prefix: "Bearer "             # Note the trailing space
      env_var: PETSTORE_API_KEY
```

**Result**:
```
Authorization: Bearer your-api-key-here
```

### Interactive API Key Setup

Users can set API key interactively:

```bash
# First time setup
petstore login

# Prompts for API key
Enter your API key: **********************
✓ API key saved to keyring

# Subsequent commands use saved key
petstore pets list
```

### Security Notes

- Never embed API keys in configuration files
- Always use environment variables or keyring storage
- API keys are automatically masked in logs and output
- Keys are stored securely in OS keyring when available

---

## Basic Authentication

Basic authentication sends username and password with each request using the HTTP Basic authentication scheme.

**Security Warning**: Basic auth sends credentials with every request. Use HTTPS and consider OAuth2 for production.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: PETSTORE_USERNAME     # Environment variable for username
      password_env: PETSTORE_PASSWORD     # Environment variable for password
```

### Usage

```bash
# Set credentials via environment variables
export PETSTORE_USERNAME=myusername
export PETSTORE_PASSWORD=mypassword

# CLI automatically adds header:
# Authorization: Basic <base64(username:password)>
petstore pets list
```

### Interactive Login

```bash
# Login command prompts for credentials
petstore login

# Prompts
Username: myusername
Password: **********************
✓ Credentials saved to keyring

# Subsequent commands use saved credentials
petstore pets list
```

### Configuration Fields

| Field | Description | Required | Example |
|-------|-------------|----------|---------|
| `username_env` | Environment variable for username | Yes | `PETSTORE_USERNAME` |
| `password_env` | Environment variable for password | Yes | `PETSTORE_PASSWORD` |

### Security Notes

- Credentials are sent with every request (unlike OAuth2 tokens)
- Always use HTTPS to prevent credential interception
- Credentials are automatically masked in logs
- Consider migrating to OAuth2 for better security

---

## OAuth2 Authentication

OAuth2 is the industry-standard protocol for authorization. CliForge supports all major OAuth2 flows.

### OAuth2 Flows

CliForge supports four OAuth2 flows:

| Flow | Use Case | User Interaction | Client Secret |
|------|----------|------------------|---------------|
| **Authorization Code** | User authentication | Browser-based | Optional (PKCE) |
| **Client Credentials** | Service-to-service | None | Required |
| **Password** | Legacy user auth | CLI prompts | Optional |
| **Device Code** | Limited input devices | External device | Optional |

### Flow Selection Guide

```
┌─────────────────────────────────────────────────────────────┐
│  Is this a user-facing CLI?                                 │
│  └─ Yes → Use "authorization_code" with PKCE                │
│  └─ No  → Continue                                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Is this for server-to-server communication?                │
│  └─ Yes → Use "client_credentials"                          │
│  └─ No  → Continue                                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│  Does the device have limited input (TV, IoT)?              │
│  └─ Yes → Use "device_code"                                 │
│  └─ No  → Use "password" (legacy, not recommended)          │
└─────────────────────────────────────────────────────────────┘
```

---

## OAuth2 Flow: Authorization Code

The authorization code flow is the most secure OAuth2 flow for user authentication.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      client_id: petstore-cli
      client_secret: ${OAUTH_CLIENT_SECRET}  # Optional with PKCE
      auth_url: https://auth.petstore.example.com/authorize
      token_url: https://auth.petstore.example.com/token
      redirect_url: http://localhost:8085/callback
      scopes:
        - read:pets
        - write:pets
      pkce: true                            # Recommended for public clients

      # Token storage
      storage:
        type: keyring                       # keyring, file, memory
        keyring_service: petstore-cli
        keyring_user: default
```

### Configuration Fields

| Field | Description | Required | Example |
|-------|-------------|----------|---------|
| `client_id` | OAuth2 client identifier | Yes | `petstore-cli` |
| `client_secret` | OAuth2 client secret | No (with PKCE) | `secret123` |
| `auth_url` | Authorization endpoint | Yes | `https://auth.example.com/authorize` |
| `token_url` | Token endpoint | Yes | `https://auth.example.com/token` |
| `redirect_url` | Callback URL | Yes | `http://localhost:8085/callback` |
| `scopes` | Requested scopes | No | `[read:pets, write:pets]` |
| `pkce` | Enable PKCE | Recommended | `true` |

### Authorization Flow

1. **User initiates login**:
   ```bash
   petstore login
   ```

2. **CLI displays authorization URL**:
   ```
   Please visit this URL to authorize:

   https://auth.petstore.example.com/authorize?client_id=petstore-cli&...

   Waiting for authorization...
   ```

3. **User authorizes in browser**:
   - Browser opens authorization page
   - User logs in and grants permissions
   - Browser redirects to `http://localhost:8085/callback?code=...`

4. **CLI receives code and exchanges for token**:
   ```
   ✓ Authorization successful!
   ✓ Token saved to keyring
   ```

5. **Token used for API requests**:
   ```bash
   # CLI automatically adds header:
   # Authorization: Bearer <access_token>
   petstore pets list
   ```

### PKCE (Proof Key for Code Exchange)

PKCE is a security extension for OAuth2 that protects against authorization code interception.

**Enable PKCE**:

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      pkce: true                            # Enable PKCE
      client_secret: ""                     # Not needed with PKCE
```

**Benefits**:
- No client secret required
- Protection against code interception
- Recommended for all public clients (CLIs, mobile apps)

### Custom Redirect Port

Change the callback port if 8085 is in use:

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      redirect_url: http://localhost:9000/callback
```

**Note**: Must match the redirect URL registered with OAuth2 provider.

---

## OAuth2 Flow: Client Credentials

The client credentials flow is for server-to-server authentication without user interaction.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: client_credentials
      client_id: petstore-service
      client_secret: ${OAUTH_CLIENT_SECRET}  # Required
      token_url: https://auth.petstore.example.com/token
      scopes:
        - api:read
        - api:write

      storage:
        type: file
        path: ~/.petstore/token.json
```

### Configuration Fields

| Field | Description | Required | Example |
|-------|-------------|----------|---------|
| `client_id` | OAuth2 client identifier | Yes | `petstore-service` |
| `client_secret` | OAuth2 client secret | Yes | `secret123` |
| `token_url` | Token endpoint | Yes | `https://auth.example.com/token` |
| `scopes` | Requested scopes | No | `[api:read, api:write]` |

### Usage

```bash
# Set client secret
export OAUTH_CLIENT_SECRET=your-client-secret

# Login (non-interactive)
petstore login
# ✓ Token obtained successfully
# ✓ Token saved to ~/.petstore/token.json

# Use CLI
petstore pets list
```

### Service Account Best Practices

- Store client secret in environment variables or secrets manager
- Use minimal scopes required for the task
- Rotate client secrets regularly
- Monitor token usage and set up alerts

---

## OAuth2 Flow: Password

The resource owner password credentials flow allows users to provide username/password directly to the CLI.

**Security Warning**: This flow is considered legacy and less secure. Use authorization code flow when possible.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: password
      client_id: petstore-cli
      client_secret: ${OAUTH_CLIENT_SECRET}  # Optional
      token_url: https://auth.petstore.example.com/token
      scopes:
        - read:pets
        - write:pets

      storage:
        type: keyring
        keyring_service: petstore-cli
        keyring_user: default
```

### Usage

```bash
# Login prompts for credentials
petstore login

# Prompts
Username: myusername
Password: **********************
✓ Token obtained successfully
✓ Token saved to keyring

# Use CLI
petstore pets list
```

### When to Use

- Legacy APIs that don't support authorization code flow
- Internal tools where user experience is prioritized over security
- Migration path from Basic auth to OAuth2

### Security Considerations

- Credentials are sent directly to token endpoint
- Less secure than authorization code flow
- Should only be used over HTTPS
- Consider migrating to authorization code flow with PKCE

---

## OAuth2 Flow: Device Code

The device code flow is designed for devices with limited input capabilities (smart TVs, IoT devices, CLI tools).

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: device_code
      client_id: petstore-cli
      device_code_url: https://auth.petstore.example.com/device/code
      token_url: https://auth.petstore.example.com/token
      scopes:
        - read:pets
        - write:pets

      storage:
        type: keyring
        keyring_service: petstore-cli
        keyring_user: default
```

### Configuration Fields

| Field | Description | Required | Example |
|-------|-------------|----------|---------|
| `client_id` | OAuth2 client identifier | Yes | `petstore-cli` |
| `device_code_url` | Device authorization endpoint | Yes | `https://auth.example.com/device/code` |
| `token_url` | Token endpoint | Yes | `https://auth.example.com/token` |
| `scopes` | Requested scopes | No | `[read:pets, write:pets]` |

### Device Flow Process

1. **User initiates login**:
   ```bash
   petstore login
   ```

2. **CLI displays user code**:
   ```
   Device Authorization:
   User Code: ABCD-EFGH
   Verification URL: https://auth.petstore.example.com/device

   Or visit: https://auth.petstore.example.com/device?user_code=ABCD-EFGH

   Waiting for authorization...
   ```

3. **User authorizes on another device**:
   - Visit verification URL on phone/computer
   - Enter user code: `ABCD-EFGH`
   - Authorize the application

4. **CLI polls for completion**:
   ```
   ✓ Authorization successful!
   ✓ Token saved to keyring
   ```

5. **Token used for API requests**:
   ```bash
   petstore pets list
   ```

### Polling Behavior

- CLI polls token endpoint at intervals specified by authorization server
- Default interval: 5 seconds
- Respects `slow_down` responses by doubling interval
- Times out after device code expiration (typically 15 minutes)

---

## Token Storage

CliForge supports three token storage backends for persisting authentication tokens.

### Storage Types

| Type | Description | Security | Availability | Use Case |
|------|-------------|----------|--------------|----------|
| **Keyring** | OS credential manager | High | Most platforms | Production use |
| **File** | Encrypted JSON file | Medium | All platforms | Fallback option |
| **Memory** | In-memory only | Low | All platforms | Testing only |

### Storage Type Selection

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      # ... OAuth2 config ...

      storage:
        type: keyring               # keyring, file, or memory
        keyring_service: petstore-cli
        keyring_user: default
```

---

## Keyring Storage

Keyring storage uses the operating system's secure credential storage.

### Configuration

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: keyring
        keyring_service: petstore-cli         # Service name
        keyring_user: default                 # User/account name
```

### Platform Support

| Platform | Backend | Location |
|----------|---------|----------|
| **macOS** | Keychain | Keychain Access app |
| **Windows** | Credential Manager | Control Panel > Credential Manager |
| **Linux** | Secret Service | GNOME Keyring, KWallet, etc. |

### Advantages

- Operating system handles encryption
- Tokens survive system reboots
- Accessible from OS credential manager
- Most secure option

### Viewing Stored Tokens

**macOS**:
```bash
# Open Keychain Access
open "/Applications/Utilities/Keychain Access.app"

# Search for "petstore-cli"
```

**Windows**:
```
Control Panel → Credential Manager → Windows Credentials
Look for "petstore-cli"
```

**Linux**:
```bash
# GNOME Keyring
seahorse

# KWallet
kwalletmanager
```

---

## File Storage

File storage saves tokens to an encrypted JSON file.

### Configuration

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: file
        path: ~/.petstore/token.json          # File path
```

### File Location

Tokens are stored in XDG-compliant locations:

| Platform | Default Path |
|----------|-------------|
| **Linux/macOS** | `~/.local/share/petstore/token.json` |
| **Windows** | `%LOCALAPPDATA%\petstore\token.json` |

### File Structure

```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "v1.MRbKXCQ...",
  "token_type": "Bearer",
  "expires_at": "2025-01-23T12:00:00Z",
  "scopes": ["read:pets", "write:pets"],
  "extra": {
    "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### File Permissions

Tokens files are created with restrictive permissions:

```bash
# Permissions: 0600 (rw-------)
# Only the owner can read/write
-rw------- 1 user user 1234 Jan 23 12:00 token.json
```

### Advantages

- Works on all platforms
- No dependency on OS credential manager
- Easy to inspect and debug
- Portable across systems

### Disadvantages

- Less secure than keyring
- File can be read if permissions are changed
- Not integrated with OS credential management

---

## Memory Storage

Memory storage keeps tokens in memory only (not persisted).

### Configuration

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        type: memory
```

### Behavior

- Tokens are lost when CLI exits
- Must re-authenticate on every CLI invocation
- Useful for testing and CI/CD

### Use Cases

- Automated testing
- CI/CD pipelines (use short-lived tokens)
- Development environments
- Security-sensitive scenarios (no token persistence)

---

## Multi-Tier Storage

CliForge supports fallback storage for increased reliability.

### Configuration

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
          service: petstore-cli
          user: default

        file:
          path: ~/.petstore/token.json
```

### Behavior

**On Save**:
- Attempts to save to keyring
- Falls back to file if keyring unavailable
- Returns success if either succeeds

**On Load**:
- Tries keyring first
- Falls back to file if token not in keyring
- Returns first valid token found

### Use Cases

- Cross-platform CLIs (keyring not available on all systems)
- Graceful degradation when keyring service is unavailable
- Development machines without keyring configured

---

## Token Refresh

CliForge automatically refreshes expired tokens when refresh tokens are available.

### How It Works

1. **Token expires**: Access token has expired (or expires within 30 seconds)
2. **Check for refresh token**: If refresh token is available, attempt refresh
3. **Refresh token**: Call token endpoint with refresh token
4. **Update storage**: Save new access token and refresh token
5. **Continue**: Use new access token for API request

### Configuration

Refresh is automatic when using OAuth2 with refresh tokens:

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      # ... other config ...

      # Refresh happens automatically
      # No additional configuration needed
```

### Refresh Threshold

Tokens are refreshed when they expire within 30 seconds:

```go
// Built-in refresh logic
if token.ExpiresAt.Before(time.Now().Add(30 * time.Second)) {
    // Refresh token
}
```

### Manual Refresh

Users can manually refresh tokens:

```bash
# Force token refresh
petstore auth refresh

# Output
✓ Token refreshed successfully
✓ New token saved to keyring
```

### Refresh Token Expiration

If refresh token has also expired:

```bash
petstore pets list

# Output
✗ Error: refresh token expired
  Run 'petstore login' to re-authenticate
```

### Refresh Flow Diagram

```
┌─────────────┐
│ API Request │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ Token Valid?    │──Yes──> Use Token
└────────┬────────┘
         │
         No
         │
         ▼
┌─────────────────────┐
│ Refresh Token       │──No──> Re-authenticate
│ Available?          │
└────────┬────────────┘
         │
         Yes
         │
         ▼
┌─────────────────┐
│ Call Token      │
│ Endpoint with   │
│ Refresh Token   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Save New Token  │
└────────┬────────┘
         │
         ▼
    Use Token
```

---

## Multiple Auth Providers

CliForge supports multiple authentication providers in a single CLI.

### Configuration

**Embedded config** (`cli-config.yaml`):

```yaml
behaviors:
  auth:
    # Default provider
    type: oauth2
    oauth2:
      # ... default OAuth2 config ...

    # Additional providers
    providers:
      production:
        type: oauth2
        oauth2:
          client_id: petstore-prod
          auth_url: https://auth.petstore.example.com/authorize
          token_url: https://auth.petstore.example.com/token
          storage:
            type: keyring
            keyring_service: petstore-prod

      staging:
        type: oauth2
        oauth2:
          client_id: petstore-staging
          auth_url: https://staging-auth.petstore.example.com/authorize
          token_url: https://staging-auth.petstore.example.com/token
          storage:
            type: keyring
            keyring_service: petstore-staging

      service-account:
        type: oauth2
        oauth2:
          flow: client_credentials
          client_id: petstore-service
          client_secret: ${SERVICE_ACCOUNT_SECRET}
          token_url: https://auth.petstore.example.com/token
          storage:
            type: file
            path: ~/.petstore/service-token.json
```

### Usage

```bash
# Login to specific provider
petstore login --provider production
petstore login --provider staging
petstore login --provider service-account

# Use specific provider for commands
petstore --provider staging pets list

# View current provider
petstore auth status

# Output
Provider: production
Status: Authenticated
Token expires: 2025-01-23 12:00:00 UTC (in 2 hours)
Scopes: read:pets, write:pets
```

### Provider Selection Priority

1. `--provider` flag (highest)
2. Environment variable `PETSTORE_AUTH_PROVIDER`
3. User config `preferences.auth.default_provider`
4. Embedded config default provider (lowest)

### Use Cases

- Multiple environments (production, staging, development)
- Multiple accounts (personal, work)
- Different auth types (user OAuth2, service account)
- Testing different auth configurations

---

## Environment Variables for Credentials

Use environment variables to provide credentials securely.

### Variable Naming Convention

```
<CLI_NAME>_<CREDENTIAL_NAME>
```

Examples:
- `PETSTORE_API_KEY`
- `PETSTORE_USERNAME`
- `PETSTORE_PASSWORD`
- `OAUTH_CLIENT_SECRET`

### API Key Example

```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      env_var: PETSTORE_API_KEY
```

```bash
export PETSTORE_API_KEY=your-api-key
petstore pets list
```

### Basic Auth Example

```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: PETSTORE_USERNAME
      password_env: PETSTORE_PASSWORD
```

```bash
export PETSTORE_USERNAME=myuser
export PETSTORE_PASSWORD=mypass
petstore pets list
```

### OAuth2 Client Secret Example

```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_secret: ${OAUTH_CLIENT_SECRET}
```

```bash
export OAUTH_CLIENT_SECRET=your-client-secret
petstore login
```

### Environment Files

Create environment files for different contexts:

**~/.petstore-production.env**:
```bash
export PETSTORE_API_KEY=prod-key-here
export PETSTORE_ENV=production
```

**~/.petstore-staging.env**:
```bash
export PETSTORE_API_KEY=staging-key-here
export PETSTORE_ENV=staging
```

**Usage**:
```bash
# Source environment file
source ~/.petstore-production.env
petstore pets list

# Or use with subshell
(source ~/.petstore-staging.env && petstore pets list)
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Run CLI
  env:
    PETSTORE_API_KEY: ${{ secrets.PETSTORE_API_KEY }}
  run: |
    petstore pets list
```

```yaml
# GitLab CI example
test:
  variables:
    PETSTORE_API_KEY: $PETSTORE_API_KEY_SECRET
  script:
    - petstore pets list
```

---

## Security Best Practices

### For CLI Developers

1. **Never embed secrets in configuration**
   ```yaml
   # BAD - Don't do this
   behaviors:
     auth:
       api_key:
         key: "hardcoded-api-key-12345"

   # GOOD - Use environment variables
   behaviors:
     auth:
       api_key:
         env_var: PETSTORE_API_KEY
   ```

2. **Use OAuth2 when possible**
   - Prefer authorization code flow with PKCE
   - Avoid password flow unless necessary
   - Always use HTTPS

3. **Enable automatic secret masking**
   ```yaml
   behaviors:
     secrets:
       enabled: true
       masking:
         style: partial
   ```

4. **Provide secure defaults**
   ```yaml
   behaviors:
     auth:
       oauth2:
         storage:
           primary: keyring
           fallback: file
   ```

5. **Implement token rotation**
   - Support refresh tokens
   - Automatic token refresh before expiration
   - Clear expired tokens

6. **Validate authentication configuration**
   - Validate during CLI generation
   - Validate on runtime
   - Provide clear error messages

### For CLI Users

1. **Use environment variables for credentials**
   ```bash
   # GOOD
   export PETSTORE_API_KEY=secret

   # BAD - Don't store in config files
   # ~/.config/petstore/config.yaml
   # api_key: secret
   ```

2. **Secure config files**
   ```bash
   chmod 600 ~/.config/petstore/config.yaml
   chmod 600 ~/.petstore/token.json
   ```

3. **Use keyring storage when available**
   ```bash
   # Check if keyring is available
   petstore auth status --storage

   # Output
   Storage: keyring (macOS Keychain)
   Status: Available
   ```

4. **Regularly rotate credentials**
   ```bash
   # Logout and re-authenticate
   petstore logout
   petstore login
   ```

5. **Use minimal scopes**
   - Only request scopes you need
   - Review scope requests during authorization

6. **Monitor token usage**
   ```bash
   # Check token status
   petstore auth status

   # Output
   Token expires: 2025-01-23 12:00:00 UTC
   Scopes: read:pets
   ```

7. **Revoke tokens when done**
   ```bash
   # Logout to remove stored tokens
   petstore logout

   # Or logout from all providers
   petstore logout --all
   ```

### Credentials in Scripts

**Bad Practice**:
```bash
#!/bin/bash
export PETSTORE_API_KEY="hardcoded-secret"
petstore pets list
```

**Good Practice**:
```bash
#!/bin/bash
# Load from secure location
source "$HOME/.secrets/petstore.env"
petstore pets list
```

**Better Practice**:
```bash
#!/bin/bash
# Use secrets manager
export PETSTORE_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id petstore-api-key \
  --query SecretString \
  --output text)
petstore pets list
```

---

## Troubleshooting

### Authentication Failed

```
✗ Error: authentication failed: invalid_client
```

**Possible causes**:
- Incorrect client ID or secret
- Client not registered with OAuth2 provider
- Client credentials expired

**Solutions**:
1. Verify client credentials in config
2. Check environment variables
3. Confirm client registration with provider

### Token Expired

```
✗ Error: token expired
  Run 'petstore login' to re-authenticate
```

**Solutions**:
```bash
# Re-authenticate
petstore login

# Or force token refresh
petstore auth refresh
```

### Keyring Not Available

```
⚠️ Warning: keyring not available, falling back to file storage
✓ Token saved to ~/.petstore/token.json
```

**Causes**:
- Running in headless environment
- Keyring service not running
- No keyring installed

**Solutions**:

**macOS**: Keychain should be available by default

**Linux**:
```bash
# Install GNOME Keyring
sudo apt-get install gnome-keyring

# Or KWallet
sudo apt-get install kwalletmanager
```

**Windows**: Credential Manager should be available by default

### Authorization Timeout

```
✗ Error: authorization timeout
  Authorization was not completed within 5 minutes
```

**Solutions**:
1. Retry login
2. Check browser for authorization prompt
3. Verify redirect URL is accessible

### Redirect URL Mismatch

```
✗ Error: redirect_uri_mismatch
  The redirect URI in the request does not match the registered redirect URI
```

**Solutions**:
1. Check redirect URL in config matches registered URL
2. Ensure port matches (e.g., 8085 vs 8080)
3. Verify protocol (http vs https)

### Invalid Scopes

```
✗ Error: invalid_scope
  One or more requested scopes are invalid
```

**Solutions**:
1. Check scope names in config
2. Verify scopes are supported by API
3. Request only scopes you need

### Token Not Found

```
✗ Error: token not found
  Run 'petstore login' to authenticate
```

**Solutions**:
```bash
# Login to create token
petstore login
```

### Permission Denied

```
✗ Error: failed to save token: permission denied
```

**Solutions**:
```bash
# Fix permissions
chmod 700 ~/.config/petstore
chmod 600 ~/.config/petstore/token.json

# Or reset config directory
rm -rf ~/.config/petstore
mkdir -p ~/.config/petstore
chmod 700 ~/.config/petstore
```

### Debug Authentication Issues

Enable debug logging to diagnose issues:

```bash
# Enable debug mode
export PETSTORE_DEBUG=true

# Run command with verbose output
petstore --verbose login

# View authentication flow
petstore --debug pets list
```

---

## Command Reference

### Auth Commands

```bash
# Login (interactive)
petstore login
petstore login --provider staging

# Logout
petstore logout
petstore logout --all

# Check auth status
petstore auth status
petstore auth status --provider production

# Refresh token
petstore auth refresh

# View stored tokens (requires auth)
petstore auth list
```

---

## Summary

CliForge provides comprehensive authentication support:

- **Multiple auth types**: API Key, Basic, OAuth2, None
- **Four OAuth2 flows**: Authorization Code, Client Credentials, Password, Device Code
- **Secure token storage**: Keyring, File, Memory
- **Automatic token refresh**: Seamless expiration handling
- **Multi-provider support**: Multiple auth configurations
- **Environment variables**: Secure credential management
- **Security best practices**: Built-in secret masking and protection

By choosing the right authentication type and following security best practices, you can build secure and user-friendly CLIs with CliForge.

For general configuration guidance, see the [Configuration Guide](./user-guide-configuration.md).
