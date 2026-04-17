package capture

import "testing"

func TestContainsSecret(t *testing.T) {
	cases := []struct {
		cmd      string
		expected bool
	}{
		{"git push origin main", false},
		{"npm run build", false},
		{"docker build -t myapp .", false},
		{"echo hello world", false},
		{"ls -la", false},

		// Should be redacted
		{"mysql -u root --password=secret123", true},
		{"curl -H 'Authorization: Bearer abc123token' https://api.example.com", true},
		{"export GITHUB_TOKEN=ghp_abc123", true},
		{"export AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", true},
		{"PGPASSWORD=mypass psql -U postgres", true},
		{"export API_KEY=supersecret", true},
		{"git clone https://user:password@github.com/repo.git", true},
		{"docker login --password mypassword registry.io", true},
	}

	for _, tc := range cases {
		got := containsSecret(tc.cmd)
		if got != tc.expected {
			t.Errorf("containsSecret(%q) = %v, want %v", tc.cmd, got, tc.expected)
		}
	}
}

func TestSanitize_RedactsSecrets(t *testing.T) {
	result := sanitize("export GITHUB_TOKEN=ghp_abc123")
	if result != "[redacted — command contained potential secret]" {
		t.Errorf("expected redaction, got %q", result)
	}
}

func TestSanitize_PassesThroughSafe(t *testing.T) {
	result := sanitize("git push origin main")
	if result != "git push origin main" {
		t.Errorf("expected passthrough, got %q", result)
	}
}
