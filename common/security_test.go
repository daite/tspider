package common

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

func resetConfigForTest(t *testing.T, path string) {
	t.Helper()
	configPath = path
	config = nil
	configOnce = sync.Once{}
	TorrentURL = map[string]string{}
}

func TestSetSiteURLRejectsUnsafeURLs(t *testing.T) {
	resetConfigForTest(t, filepath.Join(t.TempDir(), "config.json"))

	tests := []string{
		"javascript:alert(1)",
		"file:///etc/passwd",
		"https://user:pass@example.com",
		"http://127.0.0.1:8080",
	}

	for _, tt := range tests {
		if err := SetSiteURL("nyaa", tt); err == nil {
			t.Fatalf("SetSiteURL(%q) succeeded, want error", tt)
		}
	}
}

func TestAddSiteAcceptsHTTPSURL(t *testing.T) {
	resetConfigForTest(t, filepath.Join(t.TempDir(), "config.json"))

	if err := AddSite("mirror", "https://example.com", "jp"); err != nil {
		t.Fatalf("AddSite returned error: %v", err)
	}
	if got := GetConfig().Sites["mirror"].URL; got != "https://example.com" {
		t.Fatalf("AddSite URL = %q, want %q", got, "https://example.com")
	}
}

func TestSaveConfigWritesPrivateFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX owner-only permission bits")
	}

	path := filepath.Join(t.TempDir(), "config.json")
	resetConfigForTest(t, path)

	if err := SaveConfig(DefaultConfig()); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config permissions = %o, want 0600", got)
	}
}
