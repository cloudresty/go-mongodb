# Security Policy

[Home](README.md) &nbsp;/&nbsp; Security Policy

&nbsp;

We take the security of go-mongodb seriously. This document outlines our
security practices and how to report security vulnerabilities.

&nbsp;

## Table of Contents

- [Supported Versions](#supported-versions)
- [Reporting a Vulnerability](#reporting-a-vulnerability)
- [Security Best Practices](#security-best-practices)
- [Known Security Considerations](#known-security-considerations)
- [Security Updates](#security-updates)
- [Dependencies](#dependencies)
- [Compliance and Auditing](#compliance-and-auditing)
- [Contact Information](#contact-information)

&nbsp;

## Supported Versions

We provide security updates for the following versions:

| Version | Supported |
|---------|----------|
| Latest | ✅ Yes |
| Previous Major | ✅ Yes (6 months) |
| Older Versions | ❌ No |

**Current Support:**

- **Latest Release**: Always supported with security updates
- **Previous Major Version**: Supported for 6 months after new major release
- **Development Branches**: Not supported for security updates

🔝 [back to top](#security-policy)

&nbsp;

## Reporting a Vulnerability

&nbsp;

### Please DO NOT report security vulnerabilities through public GitHub issues

&nbsp;

### How to Report

1. **Email**: Send details to [security@cloudresty.com](mailto:security@cloudresty.com)
2. **Subject**: Include "SECURITY" and brief description
3. **Encryption**: Use our PGP key for sensitive information (see below)

🔝 [back to top](#security-policy)

&nbsp;

### What to Include

- **Description**: Clear description of the vulnerability
- **Impact**: Potential impact and attack scenarios
- **Reproduction**: Step-by-step instructions to reproduce
- **Affected Versions**: Which versions are affected
- **Environment**: Go version, OS, MongoDB version
- **Proof of Concept**: Code example (if applicable)

🔝 [back to top](#security-policy)

&nbsp;

### Response Timeline

- **Initial Response**: Within 48 hours
- **Vulnerability Assessment**: Within 5 business days
- **Fix Development**: Timeline depends on severity
- **Public Disclosure**: After fix is available and users have time to update

🔝 [back to top](#security-policy)

&nbsp;

### PGP Key for Encryption

```text
-----BEGIN PGP PUBLIC KEY BLOCK-----
[PGP key would be here in a real security policy]
-----END PGP PUBLIC KEY BLOCK-----
```

Download: [security@cloudresty.com.asc](mailto:security@cloudresty.com)

🔝 [back to top](#security-policy)

&nbsp;

## Security Best Practices

&nbsp;

### Environment Configuration

&nbsp;

#### Secure Connection URLs

```bash
# ✅ Good: Use secure protocols and explicit configuration
export MONGODB_HOSTS=cluster.example.com:27017
export MONGODB_TLS_ENABLED=true
export MONGODB_DATABASE=myapp

# ✅ Good: Strong authentication
export MONGODB_USERNAME=app_user
export MONGODB_PASSWORD=strong_random_password

# ❌ Avoid: Default credentials
export MONGODB_USERNAME=""
export MONGODB_PASSWORD=""
```

🔝 [back to top](#security-policy)

&nbsp;

#### TLS Configuration

```go
// Enable TLS with proper verification
config := &mongodb.Config{
    Host:        "cluster.example.com",
    Port:        27017,
    Database:    "myapp",
    TLSEnabled:  true,
    TLSInsecure: false, // Always verify certificates in production
}

client, err := mongodb.NewClientWithConfig(config)
```

🔝 [back to top](#security-policy)

&nbsp;

### Access Control

&nbsp;

#### Connection Permissions

- Use dedicated service accounts for applications
- Follow principle of least privilege
- Regularly rotate credentials
- Use strong, unique passwords

🔝 [back to top](#security-policy)

&nbsp;

#### MongoDB User Configuration

```bash
# Create dedicated user for your application
use myapp
db.createUser({
  user: "myapp_user",
  pwd: "secure_password",
  roles: [
    { role: "readWrite", db: "myapp" }
  ]
})

# Remove default admin user in production environments
# Follow MongoDB security best practices
```

🔝 [back to top](#security-policy)

&nbsp;

### Network Security

&nbsp;

#### Firewall Configuration

```bash
# Allow only necessary ports
# MongoDB: 27017 (default) or custom port
# Only from trusted networks and IP ranges
```

🔝 [back to top](#security-policy)

&nbsp;

#### VPC/Network Isolation

- Deploy MongoDB in private subnets
- Use VPC endpoints where available
- Implement network segmentation
- Monitor network traffic

🔝 [back to top](#security-policy)

&nbsp;

### Message Security

&nbsp;

#### Sensitive Data Handling

```go
// ❌ Avoid: Putting sensitive data in document fields without encryption
document := bson.M{
    "credit_card": "4111-1111-1111-1111", // Never do this
}

// ✅ Good: Encrypt sensitive data before storing
encryptedData, err := encrypt(sensitiveData)
document := bson.M{
    "encrypted_data": encryptedData,
    "encryption_info": bson.M{
        "encrypted": true,
        "algorithm": "AES-256-GCM",
    },
}
```

🔝 [back to top](#security-policy)

&nbsp;

#### Document Encryption

```go
// Example: Encrypt document fields before inserting
func insertSecureDocument(collection *mongodb.Collection, data any) error {
    // Encrypt sensitive fields
    encryptedDoc, err := encryptSensitiveFields(data)
    if err != nil {
        return err
    }

    // Insert encrypted document
    result, err := collection.InsertOne(ctx, encryptedDoc)
    if err != nil {
        return err
    }

    return nil
}
```

🔝 [back to top](#security-policy)

&nbsp;

### Logging Security

&nbsp;

#### Automatic PII Protection

The package automatically sanitizes sensitive information in logs:

```go
// Connection URLs are automatically sanitized
emit.Info.StructuredFields("Connecting to MongoDB",
    emit.ZString("host", sanitizeHost(connectionHost))) // Passwords removed
```

🔝 [back to top](#security-policy)

&nbsp;

#### Custom Log Sanitization

```go
// Sanitize custom sensitive data
func logOperation(documentID string, userID string) {
    emit.Info.StructuredFields("Processing document",
        emit.ZString("document_id", documentID),
        emit.ZString("user_id", sanitizeUserID(userID))) // Custom sanitization
}

func sanitizeUserID(userID string) string {
    if len(userID) <= 4 {
        return "***"
    }
    return userID[:4] + "***"
}
```

🔝 [back to top](#security-policy)

&nbsp;

## Known Security Considerations

&nbsp;

### Document Persistence vs Performance

- **Acknowledged writes** are safer but slower
- **Unacknowledged writes** are faster but can be lost
- Choose based on your security/performance requirements

```go
// High security: Use acknowledged writes with journal
config := mongodb.InsertConfig{
    WriteConcern: writeconcern.Majority(),
    Journal:      true, // Wait for journal confirmation
}

// High performance: Use faster writes for non-critical data
config := mongodb.InsertConfig{
    WriteConcern: writeconcern.Unacknowledged(),
    Journal:      false, // Faster but can be lost
}
```

🔝 [back to top](#security-policy)

&nbsp;

### Connection Security

&nbsp;

#### Connection Limits

```go
// Configure connection limits to prevent resource exhaustion
config := &mongodb.Config{
    URI:            "mongodb://localhost:27017",
    MaxPoolSize:    100,   // Limit connections in pool
    MinPoolSize:    5,     // Minimum pool size
    MaxIdleTime:    time.Minute * 10, // Connection idle timeout
    ConnectTimeout: time.Second * 10, // Connection timeout
}
```

🔝 [back to top](#security-policy)

&nbsp;

### GridFS Security

- GridFS collections may contain sensitive file data
- Implement appropriate access controls for GridFS
- Consider data retention policies for stored files
- Monitor GridFS storage for security issues

🔝 [back to top](#security-policy)

&nbsp;

### ULID Security

ULIDs are designed to be safe for use as public identifiers:

- **No sensitive information**: ULIDs don't contain user data
- **Collision resistant**: Cryptographically secure randomness
- **Time-based**: Only timestamp is extractable (no other metadata)
- **URL safe**: No encoding issues or injection risks

🔝 [back to top](#security-policy)

&nbsp;

## Security Updates

&nbsp;

### Update Process

1. **Assessment**: We evaluate reported vulnerabilities
2. **Classification**: Severity assessment using CVSS
3. **Development**: Fix development with security review
4. **Testing**: Comprehensive security testing
5. **Release**: Coordinated disclosure and patch release
6. **Notification**: Security advisory and user notification

🔝 [back to top](#security-policy)

&nbsp;

### Severity Levels

| Severity | Description | Response Time |
|----------|-------------|---------------|
| Critical | Immediate exploitation possible | 24-48 hours |
| High | Significant security risk | 1 week |
| Medium | Moderate security impact | 2 weeks |
| Low | Minor security consideration | Next release |

🔝 [back to top](#security-policy)

&nbsp;

### Security Advisories

Security updates are published through:

- **GitHub Security Advisories**: Primary notification channel
- **Release Notes**: Detailed in version release notes
- **Mailing List**: [security-announce@cloudresty.com](mailto:security-announce@cloudresty.com)
- **Documentation**: Updated security documentation

🔝 [back to top](#security-policy)

&nbsp;

## Dependencies

&nbsp;

### Dependency Security

We regularly audit and update dependencies:

- **Automated Scanning**: Dependabot and security scanners
- **Regular Updates**: Dependencies updated in minor releases
- **Vulnerability Monitoring**: Continuous monitoring for known CVEs
- **Minimal Dependencies**: We keep dependencies minimal to reduce attack surface

🔝 [back to top](#security-policy)

&nbsp;

### Current Dependencies

Main dependencies (as of latest version):

```text
go.mongodb.org/mongo-driver/v2  # MongoDB Go driver v2
github.com/cloudresty/emit      # Structured logging
github.com/cloudresty/ulid      # ULID generation
github.com/cloudresty/go-env    # Environment configuration
```

🔝 [back to top](#security-policy)

&nbsp;

### Dependency Policy

- **Security Patches**: Applied as soon as available
- **Major Updates**: Evaluated for security improvements
- **New Dependencies**: Security review required
- **Deprecated Dependencies**: Replaced promptly

🔝 [back to top](#security-policy)

&nbsp;

## Compliance and Auditing

&nbsp;

### Security Standards

This package is designed to support:

- **SOC 2**: Audit logging and access controls
- **GDPR**: Data encryption and sanitization features
- **HIPAA**: Encryption and audit trail capabilities
- **PCI DSS**: Secure communication and data handling

🔝 [back to top](#security-policy)

&nbsp;

### Audit Trail

The package provides comprehensive audit capabilities:

```go
// All operations are automatically logged with structured data
// Including: connection attempts, document operations, queries
// Logs include: timestamps, operation details, performance metrics

// Example audit log entry:
{"timestamp": "2024-01-01T12:00:00Z", "level": "info", "message": "Document inserted", "collection": "orders", "document_id": "01HQRS...", "ulid": "01HQRS..."}
```

🔝 [back to top](#security-policy)

&nbsp;

## Contact Information

&nbsp;

### Security Team

- **Email**: [security@cloudresty.com](mailto:security@cloudresty.com)
- **Response Time**: 48 hours for initial response
- **Office Hours**: Monday-Friday, 9 AM - 5 PM UTC

🔝 [back to top](#security-policy)

&nbsp;

### For Security Researchers

We welcome responsible disclosure from security researchers:

- **Bug Bounty**: Contact us about our bug bounty program
- **Hall of Fame**: Public recognition for security contributors
- **Coordination**: We work with researchers on responsible disclosure

🔝 [back to top](#security-policy)

&nbsp;

## Acknowledgments

We thank the security community and all researchers who have contributed to
making go-mongodb more secure.

&nbsp;

---

&nbsp;

An open source project brought to you by the [Cloudresty](https://cloudresty.com/) team.

[Website](https://cloudresty.com/) | [LinkedIn](https://www.linkedin.com/company/cloudresty) | [BlueSky](https://bsky.app/profile/cloudresty.com) | [GitHub](https://github.com/cloudresty)

&nbsp;
