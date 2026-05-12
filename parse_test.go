package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseModAndSaveModRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	content := "module example.com/m\n\ngo 1.23\n\nrequire github.com/foo/bar v1.0.0\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	mod, err := parseMod(path)
	if err != nil {
		t.Fatalf("parseMod: %v", err)
	}
	if mod == nil || mod.Module == nil {
		t.Fatal("expected parsed module")
	}
	if mod.Module.Mod.Path != "example.com/m" {
		t.Errorf("module path = %q", mod.Module.Mod.Path)
	}

	if err := saveMod(path, mod); err != nil {
		t.Fatalf("saveMod: %v", err)
	}
	mod2, err := parseMod(path)
	if err != nil {
		t.Fatalf("parseMod second: %v", err)
	}
	if diff := cmp.Diff(mod.Module.Mod.Path, mod2.Module.Mod.Path); diff != "" {
		t.Errorf("round trip module path (-want +got):\n%s", diff)
	}
}
