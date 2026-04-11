package skills

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SecurityScanner performs 8-point security scans on skills
type SecurityScanner struct{}

// NewSecurityScanner creates a new SecurityScanner
func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{}
}

// Scan performs all 8 security checks on a skill
func (s *SecurityScanner) Scan(ctx context.Context, skill *Skill) (*SecurityScanResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	if skill == nil {
		return &SecurityScanResult{
			ScannedAt:   time.Now(),
			OverallPass: true,
			Checks:      make(map[string]SecurityCheck),
		}, nil
	}

	result := &SecurityScanResult{
		ScannedAt:   time.Now(),
		OverallPass: true,
		Checks:      make(map[string]SecurityCheck),
	}

	checks := []func(*Skill) SecurityCheck{
		s.scanFileStructure,
		s.scanDangerousCommands,
		s.scanHardcodedSecrets,
		s.scanEnvHarvesting,
		s.scanNetworkAccess,
		s.scanObfuscation,
		s.scanPromptInjection,
		s.scanPrivilegeEscalation,
	}

	for _, check := range checks {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("context cancelled during scan: %w", err)
		}
		c := check(skill)
		result.Checks[c.Name] = c
		if !c.Passed {
			result.OverallPass = false
		}
	}

	return result, nil
}

// scanFileStructure checks for unexpected files (hidden files, symlinks, etc.)
func (s *SecurityScanner) scanFileStructure(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "file_structure",
		Passed:      true,
		Severity:    "medium",
		Description: "Check for unexpected files such as hidden files, .env files, or symlinks",
		Findings:    []string{},
	}

	flaggedPatterns := []string{".bashrc", ".profile", ".env", ".bash_profile", ".zshrc"}

	for filename := range skill.Content {
		// Check for hidden files (starting with .)
		if strings.HasPrefix(filename, ".") {
			check.Passed = false
			check.Findings = append(check.Findings, fmt.Sprintf("hidden file detected: %s", filename))
		}

		// Check for flagged filenames
		for _, flagged := range flaggedPatterns {
			if filename == flagged || strings.HasSuffix(filename, "/"+flagged) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("flagged file detected: %s", filename))
			}
		}
	}

	return check
}

// scanDangerousCommands searches for dangerous command patterns
func (s *SecurityScanner) scanDangerousCommands(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "dangerous_commands",
		Passed:      true,
		Severity:    "critical",
		Description: "Check for dangerous shell commands that could harm the system",
		Findings:    []string{},
	}

	dangerousPatterns := []string{
		`rm\s+-rf`,
		`chmod\s+777`,
		`curl\s+\S+\s*\|\s*bash`,
		`curl\s+\S+\|sh`,
		`eval\s*\(`,
		`exec\s*\(`,
		`\bsudo\b`,
		`dd\s+if=`,
		`\bmkfs\b`,
		`:\(\)\{:\|:&\};:`,
	}

	for _, pattern := range dangerousPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("dangerous pattern '%s' found in %s", pattern, filename))
			}
		}
	}

	return check
}

// scanHardcodedSecrets searches for hardcoded credentials
func (s *SecurityScanner) scanHardcodedSecrets(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "hardcoded_secrets",
		Passed:      true,
		Severity:    "critical",
		Description: "Check for hardcoded secrets, API keys, and credentials",
		Findings:    []string{},
	}

	secretPatterns := []string{
		`AKIA[0-9A-Z]{16}`,
		`(?i)token["']?\s*[:=]\s*["'][a-zA-Z0-9]{20,}["']`,
		`(?i)password\s*[:=]\s*["'][^"']+["']`,
		`(?i)secret[_-]?key\s*[:=]\s*["'][a-zA-Z0-9]{16,}["']`,
		`(?i)api[_-]?key\s*[:=]\s*["'][a-zA-Z0-9]{16,}["']`,
	}

	for _, pattern := range secretPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("hardcoded secret pattern '%s' found in %s", pattern, filename))
			}
		}
	}

	return check
}

// scanEnvHarvesting checks for reads of sensitive environment variables
func (s *SecurityScanner) scanEnvHarvesting(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "env_harvesting",
		Passed:      true,
		Severity:    "high",
		Description: "Check for reads of sensitive environment variables",
		Findings:    []string{},
	}

	sensitiveVars := []string{
		`AWS_SECRET_KEY`,
		`AWS_SECRET_ACCESS_KEY`,
		`AWS_ACCESS_KEY`,
		`AWS_ACCESS_KEY_ID`,
		`GITHUB_TOKEN`,
		`PRIVATE_KEY`,
		`DATABASE_URL`,
		`API_KEY`,
	}

	for _, envVar := range sensitiveVars {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(envVar))
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("sensitive env var '%s' accessed in %s", envVar, filename))
			}
		}
	}

	return check
}

// scanNetworkAccess looks for outbound HTTP/HTTPS calls
func (s *SecurityScanner) scanNetworkAccess(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "network_access",
		Passed:      true,
		Severity:    "medium",
		Description: "Check for outbound network access via HTTP/HTTPS calls",
		Findings:    []string{},
	}

	networkPatterns := []string{
		`\bcurl\s+`,
		`\bwget\s+`,
		`\bfetch\s*\(`,
		`\bhttp\.Request`,
		`\burllib`,
		`\brequests\.get`,
	}

	for _, pattern := range networkPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("network access pattern '%s' found in %s", pattern, filename))
			}
		}
	}

	return check
}

// scanObfuscation detects obfuscated scripts
func (s *SecurityScanner) scanObfuscation(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "obfuscation",
		Passed:      true,
		Severity:    "high",
		Description: "Check for obfuscated content such as base64-encoded scripts, minified single-line scripts, or unicode escapes",
		Findings:    []string{},
	}

	obfuscationPatterns := []string{
		`base64\s+(-d|--decode)`,
		`b64decode\s*\(`,
		`atob\s*\(`,
		`\\u[0-9a-fA-F]{4}`,
	}

	for _, pattern := range obfuscationPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("obfuscation pattern '%s' found in %s", pattern, filename))
			}
		}
	}

	// Check for single-line scripts that are very long (potential minification)
	for filename, content := range skill.Content {
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			if len(line) > 1000 {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("potential minified single-line script in %s (line length: %d)", filename, len(line)))
			}
		}
	}

	return check
}

// scanPromptInjection detects prompt injection attempts
func (s *SecurityScanner) scanPromptInjection(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "prompt_injection",
		Passed:      true,
		Severity:    "critical",
		Description: "Check for prompt injection attempts that try to override system policy",
		Findings:    []string{},
	}

	injectionPatterns := []string{
		`(?i)ignore\s+previous\s+instructions`,
		`(?i)ignore\s+all\s+prior\s+rules`,
		`(?i)disregard\s+system\s+prompts`,
		`(?i)you\s+are\s+now\s+in\s+developer\s+mode`,
		`(?i)ignore\s+all\s+instructions`,
		`(?i)disregard\s+previous`,
	}

	for _, pattern := range injectionPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("prompt injection pattern found in %s", filename))
			}
		}
	}

	return check
}

// scanPrivilegeEscalation checks for privilege escalation attempts
func (s *SecurityScanner) scanPrivilegeEscalation(skill *Skill) SecurityCheck {
	check := SecurityCheck{
		Name:        "privilege_escalation",
		Passed:      true,
		Severity:    "critical",
		Description: "Check for privilege escalation attempts such as sudo, su root, or setuid",
		Findings:    []string{},
	}

	escalationPatterns := []string{
		`\bsudo\b`,
		`\bsu\s+root\b`,
		`\bchmod\s+\+s\b`,
		`\bchmod\s+[0-7]*[4-7][0-7]{2}\b`,
		`(?i)escalat(e|ing|ion)`,
		`\broot\b`,
	}

	for _, pattern := range escalationPatterns {
		re := regexp.MustCompile(pattern)
		for filename, content := range skill.Content {
			if re.MatchString(content) {
				check.Passed = false
				check.Findings = append(check.Findings, fmt.Sprintf("privilege escalation pattern '%s' found in %s", pattern, filename))
			}
		}
	}

	return check
}
