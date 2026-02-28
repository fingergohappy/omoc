package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestListProfiles_EmptyWhenConfigDirMissing(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error: %v", err)
	}
	if len(profiles) != 0 {
		t.Fatalf("expected no profiles, got %v", profiles)
	}
}

func TestListProfiles_FindsAllValidProfiles(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	configDir := filepath.Join(tmp, "opencode")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}

	files := []string{
		DefaultProfileFile,
		TestProfileFile,
		"oh-my-opencode.claude.json",
		"oh-my-opencode.clau.json",
		"oh-my-opencode.Bad.json",
		"oh-my-opencode.test.json.20260101.bak",
		"not-oh-my-opencode.any.json",
	}
	for _, name := range files {
		path := filepath.Join(configDir, name)
		if err := os.WriteFile(path, []byte("{}\n"), 0644); err != nil {
			t.Fatalf("WriteFile %q: %v", name, err)
		}
	}

	if err := os.Symlink(TestProfileFile, filepath.Join(configDir, ActiveProfileLinkFile)); err != nil {
		t.Fatalf("Symlink: %v", err)
	}

	profiles, err := ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles() error: %v", err)
	}

	expected := []string{
		DefaultProfileFile,
		"oh-my-opencode.clau.json",
		"oh-my-opencode.claude.json",
		TestProfileFile,
	}
	if !reflect.DeepEqual(profiles, expected) {
		t.Fatalf("profiles mismatch\nexpected: %v\n     got: %v", expected, profiles)
	}
}
