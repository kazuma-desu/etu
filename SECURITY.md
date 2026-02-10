# Security Policy

## Supported Versions

The following versions of etu are currently supported with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1.0 | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability within etu, please report it responsibly.

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities via email to:

📧 **kmuthisha@gmail.com**

### What to Include

When reporting a vulnerability, please include:

1. **Description**: Clear description of the vulnerability
2. **Impact**: What could an attacker do with this vulnerability?
3. **Reproduction**: Step-by-step instructions to reproduce the issue
4. **Environment**: Version of etu, OS, Go version
5. **Screenshots/Logs**: If applicable, include relevant output
6. **Suggested Fix**: If you have one, include your suggested fix

### Response Timeline

We aim to respond to security reports within:

- **48 hours**: Acknowledgment of receipt
- **7 days**: Initial assessment and next steps
- **30 days**: Target for fix release (critical vulnerabilities)
- **90 days**: Target for fix release (non-critical vulnerabilities)

### Disclosure Policy

- We follow responsible disclosure practices
- We will keep your report confidential until a fix is released
- We will credit you in the release notes (unless you prefer to remain anonymous)
- We may request a CVE identifier for significant vulnerabilities

## Security Best Practices for Users

### Configuration Security

1. **Protect your config file**: The config file at `~/.config/etu/config.yaml` contains credentials
   - File permissions are automatically set to `0600` (owner read/write only)
   - Never commit this file to version control
   - Never share this file with others

2. **Use password-stdin for CI/CD**: When using etu in automated environments:
   ```bash
   echo "$ETCD_PASSWORD" | etu put /key value --password-stdin
   ```

3. **Rotate credentials regularly**: Change etcd passwords periodically

4. **Use TLS/mTLS in production**: Always enable TLS for production etcd clusters:
   ```bash
   etu login --context-name prod \
     --endpoints https://etcd:2379 \
     --cacert /path/to/ca.crt \
     --cert /path/to/client.crt \
     --key /path/to/client.key
   ```

### Operational Security

1. **Audit logging**: Enable etcd audit logging to track all operations
2. **Network segmentation**: Run etcd on isolated networks
3. **Principle of least privilege**: Use etcd RBAC to limit user permissions
4. **Monitor for anomalies**: Set up alerts for unusual access patterns

## Security Features

etu includes several security features:

- **Credential protection**: Config file uses 0600 permissions
- **TLS/mTLS support**: Full certificate validation
- **Key validation**: Regex-based format validation prevents injection
- **Password masking**: Passwords are masked in output and logs
- **Vulnerability scanning**: Regular govulncheck scans in CI

## Known Security Considerations

### Password Storage

**Current Status**: Passwords are stored in plaintext in the config file with 0600 permissions.

**Rationale**: 
- This is a local CLI tool (like kubectl, aws-cli)
- OS-level file permissions provide protection
- Encryption would require a master password, degrading UX

**Mitigations**:
- File permissions are strictly enforced
- User is warned if permissions are too open
- `--password-stdin` allows external secret management

**Future Considerations**:
- Integration with OS keyrings (keychain, libsecret)
- Support for external secret managers (HashiCorp Vault, AWS Secrets Manager)

## Acknowledgments

We thank the following security researchers who have responsibly disclosed vulnerabilities:

*No vulnerabilities reported yet*

## Security Updates

Security updates are released as patch versions (e.g., 0.1.1). Subscribe to GitHub releases to be notified:

- Watch this repository on GitHub
- Enable "Releases only" notifications

---

Last updated: 2026-02-10
