<!-- markdownlint-disable-file -->

id: no-hardcoded-secrets
version: 2.0.0
status: stable
enforced_by:

- pre-commit: guardian-hook
- ci_linter: git-secrets
- runtime: vault-checker
  keywords:
- security
- secrets-management
- compliance
- production-risk

# Rule: Never Hardcode Secrets

## Definition

**No API keys, passwords, tokens, or private URLs hardcoded in version control or code.**

- ❌ `const API_KEY = "sk-abc123xyz"`
- ❌ `DATABASE_PASSWORD=prod_password`
- ❌ `privateKey: "-----BEGIN RSA KEY-----..."`
- ✅ `const API_KEY = process.env.OPENAI_API_KEY`
- ✅ Load from secure vault (Vault, AWS Secrets Manager, 1Password)

## Why It Exists

**Real Impact**:

- 2024 incident: Credentials leaked on GitHub → $6K AWS charges in 1 hour
- Compliance: OWASP A01:2021 (broken access control #1)
- Recovery cost: 10-100x development time
- Regulatory: GDPR fines up to 4% of revenue

**Data**: 40% of data breaches trace to hardcoded credentials (GitGuardian 2025)

## Enforcement Mechanisms

### Layer 1: Pre-Commit Hook (LOCAL)

```bash
# .git/hooks/pre-commit
# Blocks commit if secrets detected
git-secrets --scan

# Run: git commit -m "message"
# If secret found:
#   ❌ FAILED: hardcoded secret detected at line 23
#   ❌ Use: process.env.SECRET_NAME instead
#   ❌ Commit blocked
```

**By-pass**: Requires approval from security lead (logged)

### Layer 2: CI Linter (GITHUB)

```yaml
# .github/workflows/scan-secrets.yml
- name: Scan for Secrets
  run: |
    npm run secrets:scan
    # or: docker run zricethezav/gitleaks detect --source . --exit-code 1

# Result: If found, PR fails (cannot merge)
```

### Layer 3: Runtime Daemon (PRODUCTION)

```go
// Daemon startup check
secrets := scanForSecrets(codebase)
if len(secrets) > 0 {
    logger.Fatal("❌ HARDCODED SECRETS DETECTED", secrets)
    // Never starts if secrets found
}
```

## Pattern Examples

### ❌ WRONG (All will be CAUGHT)

```javascript
// File: src/config.js
const config = {
  OPENAI_KEY: "sk-abc123xyz", // ❌ BLOCKED
  DATABASE_PASSWORD: "prod_secret", // ❌ BLOCKED
  JWT_SECRET: "my-secret-key", // ❌ BLOCKED
  SLACK_WEBHOOK: "https://hooks.slack.com/services/XXX/YYY/ZZZ", // ❌ BLOCKED
};
```

```bash
# File: .env (COMMITTED)
# ❌ BLOCKED - Never commit .env
DB_USER=admin
DB_PASSWORD=secret123
API_KEY=sk-live-123
```

```python
# File: db.py
PASSWORD = "postgres_admin_password"  # ❌ BLOCKED
API_TOKEN = "ghp_abc123xyz"           # ❌ BLOCKED
```

### ✅ RIGHT (All will PASS)

```javascript
// File: src/config.js
const config = {
  OPENAI_KEY: process.env.OPENAI_API_KEY,
  DATABASE_PASSWORD: process.env.DATABASE_PASSWORD,
  JWT_SECRET: process.env.JWT_SECRET,
  SLACK_WEBHOOK: process.env.SLACK_WEBHOOK,
};
// If env var missing:
if (!config.OPENAI_KEY) {
  throw new Error("OPENAI_API_KEY not set. See .env.example");
}
```

```bash
# File: .env (NEVER COMMITTED, in .gitignore)
# ✅ ALLOWED - Git ignores this file
DB_USER=admin
DB_PASSWORD=secret123
API_KEY=sk-live-123
```

```
# File: .env.example (COMMITTED, with placeholders)
# ✅ ALLOWED - Shows structure without real values
DB_USER=
DB_PASSWORD=
API_KEY=sk-xxx-YOUR-KEY-HERE
```

```python
# File: db.py
PASSWORD = os.getenv("DATABASE_PASSWORD")
if not PASSWORD:
    raise ValueError("DATABASE_PASSWORD env var required. See .env.example")

API_TOKEN = os.getenv("API_TOKEN")
```

## Common Mistakes

| Mistake                                    | Why It Fails                      | Fix                                           |
| ------------------------------------------ | --------------------------------- | --------------------------------------------- |
| **Commit .env file**                       | Exposes all secrets               | Add to .gitignore: `echo .env >> .gitignore`  |
| **Think "local only is safe"**             | One `git push` exposes everything | Treat all development as potential production |
| **Example values that resemble real keys** | Attacker uses them                | Use obvious placeholders: "YOUR-KEY-HERE"     |
| **.env.example has real values**           | Defeats the point                 | Real values ONLY in local .env                |
| **Secrets in config files checked in**     | Entire team sees them             | Config from env vars only                     |

## Testing the Rule

### Verify Pre-Commit Hook Works

```bash
# Create test file with fake secret
echo "API_KEY = 'sk-test-secret'" > /tmp/test.py

# Try to stage & commit
git add /tmp/test.py
git commit -m "test"

# Expected: ❌ BLOCKED
# Message: "Secret key pattern detected"
```

### Verify CI Linter Works

```bash
# Push a commit with hardcoded secret to a PR
# Expected: CI fails with "Secrets detected"
```

### Verify Runtime Check Works

```bash
# Start daemon with hardcoded secret in codebase
# Expected: Daemon fails to start, logs FATAL
```

## Migration Path (If Found)

**For existing hardcoded secrets**:

1. **Week 1: Audit**
   - Scan entire codebase: `npm run audit:secrets`
   - List all findings
   - Evaluate risk (prod vs dev, old vs current)

2. **Week 2: Remediate**
   - Move to environment variables
   - Rotate all exposed secrets
   - Update tests to use mocks

3. **Week 3: Enforce**
   - Enable pre-commit hook
   - CI lint enabled
   - Team training

## Exception Process

**If you MUST hardcode a secret** (rare):

1. Get written approval from **Security Lead** + **Tech Lead**
2. Document rationale in code comment: `// Secret hardcoded for [reason], expires [date], approved by [name]`
3. Set calendar reminder to remove before expiry
4. Log in audit trail: who, when, why, expiry

**Example**:

```javascript
// ⚠️ EXCEPTION: Hardcoded for webhook testing, expires 2026-04-30
// Approved by: @security-lead, @tech-lead
// Reason: Integration test requires real webhook response
// To remove: Rotate secret on Slack settings
const SLACK_WEBHOOK = "https://hooks.slack.com/services/XXX/YYY/ZZZ";
```

## Related Rules

- **Use Environment Variables** (complementary)
- **Rotate Secrets Regularly** (complementary)
- **Never Log Secrets** (related, output protection)

## Success Metrics

| Metric                        | Target  | Current |
| ----------------------------- | ------- | ------- |
| Hardcoded secrets in codebase | 0       | ✅ 0    |
| .env files committed          | 0       | ✅ 0    |
| Pre-commit hook passing       | 100%    | ✅ 100% |
| CI secret scan passing        | 100%    | ✅ 100% |
| Exceptions approved           | <1/year | ✅ 0    |

## Tools Required

- **Local**: git-secrets, pre-commit hook
- **CI**: GitGuardian, TruffleHog, or gitleaks
- **Runtime**: Vault CLI, AWS Secrets SDK
- **Team**: Regular audit + training

---

**Status**: ✅ Enforced globally  
**Severity**: 🔴 CRITICAL (blocks production)  
**Last Updated**: 2026-04-03  
**Owner**: Security Team
