package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPerDependencyGitEnabled(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	config = &AppConfig{
		GoModDst:     "go.mod",
		SingleCommit: false,
		DryRun:       false,
	}
	if perDependencyGitEnabled() {
		t.Fatal("expected false without git repository")
	}

	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "config", "user.email", "t@test").Run(); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "config", "user.name", "t").Run(); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(tmp, "README")
	if err := os.WriteFile(readme, []byte("a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "add", "README").Run(); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "commit", "-m", "init").Run(); err != nil {
		t.Fatal(err)
	}
	if !perDependencyGitEnabled() {
		t.Fatal("expected true in clean repo without staged changes")
	}

	if err := os.WriteFile(readme, []byte("b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "add", "README").Run(); err != nil {
		t.Fatal(err)
	}
	if perDependencyGitEnabled() {
		t.Fatal("expected false when staged changes exist")
	}

	if err := exec.Command("git", "reset", "HEAD").Run(); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "checkout", "--", "README").Run(); err != nil {
		t.Fatal(err)
	}

	config.SingleCommit = true
	if perDependencyGitEnabled() {
		t.Fatal("expected false when SingleCommit is true")
	}
	config.SingleCommit = false

	config.DryRun = true
	if perDependencyGitEnabled() {
		t.Fatal("expected false when DryRun is true")
	}
}
