package util

import "testing"

func TestShortDir(t *testing.T) {
	tests := []struct {
		name     string
		dir      string
		expected string
	}{
		{
			name:     "shortens long path",
			dir:      "/home/user/projects/myapp",
			expected: "~/projects/myapp",
		},
		{
			name:     "keeps short path",
			dir:      "myapp",
			expected: "myapp",
		},
		{
			name:     "keeps two part path",
			dir:      "projects/myapp",
			expected: "projects/myapp",
		},
		{
			name:     "handles trailing slash",
			dir:      "/home/user/projects/myapp/",
			expected: "~/myapp/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortDir(tt.dir); got != tt.expected {
				t.Errorf("ShortDir(%q) = %q; want %q", tt.dir, got, tt.expected)
			}
		})
	}
}
