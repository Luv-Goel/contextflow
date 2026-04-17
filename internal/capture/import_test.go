package capture

import (
	"os"
	"testing"
	"time"
)

func TestParseLine_PlainBash(t *testing.T) {
	base := time.Unix(1000000, 0)
	cmd, ts := parseLine("git push origin main", base, 5)
	if cmd != "git push origin main" {
		t.Errorf("expected 'git push origin main', got %q", cmd)
	}
	if ts != base.Add(5*time.Second) {
		t.Errorf("expected base+5s, got %v", ts)
	}
}

func TestParseLine_ZshExtended(t *testing.T) {
	cmd, ts := parseLine(": 1713000000:0;docker build -t myapp .", time.Now(), 0)
	if cmd != "docker build -t myapp ." {
		t.Errorf("expected docker command, got %q", cmd)
	}
	if ts.Unix() != 1713000000 {
		t.Errorf("expected unix ts 1713000000, got %d", ts.Unix())
	}
}

func TestParseLine_ZshExtendedWithDuration(t *testing.T) {
	cmd, ts := parseLine(": 1713001234:42;npm run build", time.Now(), 0)
	if cmd != "npm run build" {
		t.Errorf("expected 'npm run build', got %q", cmd)
	}
	if ts.Unix() != 1713001234 {
		t.Errorf("expected ts 1713001234, got %d", ts.Unix())
	}
}

func TestParseLine_SkipsComments(t *testing.T) {
	cmd, _ := parseLine("# this is a comment", time.Now(), 0)
	// Should be returned as-is; caller checks for # prefix
	if cmd != "# this is a comment" {
		t.Errorf("unexpected: %q", cmd)
	}
}

func TestParseLine_EmptyLine(t *testing.T) {
	cmd, _ := parseLine("", time.Now(), 0)
	if cmd != "" {
		t.Errorf("expected empty, got %q", cmd)
	}
}

func TestImportHistoryFile_PlainFormat(t *testing.T) {
	// Write a temp bash history file
	f, err := os.CreateTemp("", "bash_history_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	lines := "git push origin main\nnpm run build\ndocker ps\ngit status\nls -la\n"
	f.WriteString(lines)
	f.Close()

	// Use temp db
	dir := t.TempDir()
	os.Setenv("HOME", dir)

	database := openTestDB(t)
	result, err := ImportHistoryFile(f.Name(), database)
	if err != nil {
		t.Fatalf("ImportHistoryFile: %v", err)
	}
	if result.Total != 5 {
		t.Errorf("expected 5 total, got %d", result.Total)
	}
	if result.Imported != 5 {
		t.Errorf("expected 5 imported, got %d", result.Imported)
	}
}

func TestImportHistoryFile_ZshFormat(t *testing.T) {
	f, err := os.CreateTemp("", "zsh_history_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	lines := ": 1713000000:0;git push origin main\n: 1713000060:5;npm test\n: 1713000120:0;docker build .\n"
	f.WriteString(lines)
	f.Close()

	dir := t.TempDir()
	os.Setenv("HOME", dir)

	database := openTestDB(t)
	result, err := ImportHistoryFile(f.Name(), database)
	if err != nil {
		t.Fatalf("ImportHistoryFile: %v", err)
	}
	if result.Total != 3 {
		t.Errorf("expected 3 total, got %d", result.Total)
	}
	if result.Imported != 3 {
		t.Errorf("expected 3 imported, got %d", result.Imported)
	}
}

func TestImportHistoryFile_SecretRedacted(t *testing.T) {
	f, err := os.CreateTemp("", "bash_history_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString("git push origin main\nexport GITHUB_TOKEN=ghp_abc123\nnpm run build\n")
	f.Close()

	dir := t.TempDir()
	os.Setenv("HOME", dir)

	database := openTestDB(t)
	result, err := ImportHistoryFile(f.Name(), database)
	if err != nil {
		t.Fatalf("ImportHistoryFile: %v", err)
	}
	// All 3 should be imported (secret is redacted, not skipped)
	if result.Total != 3 {
		t.Errorf("expected 3 total, got %d", result.Total)
	}
	if result.Imported != 3 {
		t.Errorf("expected 3 imported (redacted counts), got %d", result.Imported)
	}
}
