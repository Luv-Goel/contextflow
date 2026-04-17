package capture

import "regexp"

// secretPatterns are regexes that indicate a command likely contains a secret.
// If any match, the command is stored as "[redacted]" to protect credentials.
var secretPatterns = []*regexp.Regexp{
	// Explicit flags/args that suggest a secret value follows
	regexp.MustCompile(`(?i)--password[=\s]`),
	regexp.MustCompile(`(?i)--secret[=\s]`),
	regexp.MustCompile(`(?i)--token[=\s]`),
	regexp.MustCompile(`(?i)-p\s+\S+`), // docker -p style but also mysql -p<pass>
	regexp.MustCompile(`(?i)PGPASSWORD=`),
	regexp.MustCompile(`(?i)MYSQL_PWD=`),
	regexp.MustCompile(`(?i)AWS_SECRET`),
	regexp.MustCompile(`(?i)GITHUB_TOKEN=`),
	regexp.MustCompile(`(?i)API_KEY=`),
	// Common patterns: export SECRET=..., env VAR=...
	regexp.MustCompile(`(?i)export\s+\w*(KEY|TOKEN|SECRET|PASS|PWD|PASSWORD)\w*=`),
	// Inline env assignments before a command
	regexp.MustCompile(`(?i)\b\w*(KEY|TOKEN|SECRET|PASS|PASSWORD)\w*=\S{6,}`),
	// curl with auth headers or credentials
	regexp.MustCompile(`(?i)curl.*-[uH]\s+\S*:\S+`),
	regexp.MustCompile(`(?i)curl.*Authorization:`),
	// git credentials in URL
	regexp.MustCompile(`(?i)https?://[^@\s]+:[^@\s]+@`),
}

// containsSecret returns true if the command likely contains a credential.
func containsSecret(cmd string) bool {
	for _, re := range secretPatterns {
		if re.MatchString(cmd) {
			return true
		}
	}
	return false
}

// sanitize returns the command as-is, or a redacted placeholder if it looks
// like it contains a secret.
func sanitize(cmd string) string {
	if containsSecret(cmd) {
		return "[redacted — command contained potential secret]"
	}
	return cmd
}
