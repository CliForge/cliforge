# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.9.x   | :white_check_mark: |
| < 0.9   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report vulnerabilities via email to:

**security@cliforge.com** (or create a private security advisory on GitHub)

### What to Include

Please include the following information:

1. **Type of vulnerability** (e.g., command injection, authentication bypass, etc.)
2. **Affected component** (which package or feature)
3. **Steps to reproduce** the issue
4. **Potential impact** of the vulnerability
5. **Suggested fix** (if you have one)

### Response Timeline

- **Initial Response:** Within 48 hours
- **Status Update:** Within 7 days
- **Fix Timeline:** Depends on severity
  - Critical: 1-7 days
  - High: 7-14 days
  - Medium: 14-30 days
  - Low: Best effort

### Security Update Process

1. **Acknowledgment:** We'll confirm receipt of your report
2. **Investigation:** We'll investigate and assess the severity
3. **Fix Development:** We'll develop and test a fix
4. **Coordinated Disclosure:** We'll coordinate release timing with you
5. **Public Disclosure:** After the fix is released, we'll publish a security advisory

### Bug Bounty

We currently do not offer a bug bounty program, but we deeply appreciate security researchers who help keep CliForge secure. Security contributors will be:

- Listed in our security acknowledgments
- Credited in release notes (with permission)
- Given priority support for future issues

## Security Best Practices

For security guidance on using CliForge, see:
- [Security Guide](../blob/main/docs/security-guide.md)
- [Operations Guide](../blob/main/docs/operations-guide.md)

## Known Security Considerations

### Debug Mode
- Debug builds (metadata.debug: true) allow full configuration override
- Never use debug builds in production
- Debug mode shows security warnings on every command

### Credential Storage
- Credentials stored in OS keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- Fallback to encrypted file storage (AES-256)
- Never commit credential files to version control

### Binary Verification
- Always verify binary checksums before installation
- Use signed binaries when available
- Download only from official sources

## Security Contacts

- **Email:** security@cliforge.com
- **Security Advisories:** https://github.com/CliForge/cliforge/security/advisories

Thank you for helping keep CliForge secure!
