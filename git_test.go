package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		GoModDst: "go.mod",
		NoGit:    false,
		DryRun:   false,
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
		t.Fatal("expected true in clean repo without uncommitted changes")
	}

	if err := os.WriteFile(readme, []byte("b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "add", "README").Run(); err != nil {
		t.Fatal(err)
	}
	if perDependencyGitEnabled() {
		t.Fatal("expected false when uncommitted changes exist")
	}

	if err := exec.Command("git", "reset", "HEAD").Run(); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "checkout", "--", "README").Run(); err != nil {
		t.Fatal(err)
	}

	config.NoGit = true
	if perDependencyGitEnabled() {
		t.Fatal("expected false when NoGit is true")
	}
	config.NoGit = false

	config.DryRun = true
	if perDependencyGitEnabled() {
		t.Fatal("expected false when DryRun is true")
	}
}

func TestGitEnsureUserIdentity(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command("git", "init").Run(); err != nil {
		t.Fatal(err)
	}

	config = &AppConfig{
		GitUserName:  "Schutzbot",
		GitUserEmail: "schutzbot@gmail.com",
	}
	if err := gitEnsureUserIdentity(); err != nil {
		t.Fatal(err)
	}

	name, err := exec.Command("git", "config", "user.name").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(name)); got != "Schutzbot" {
		t.Fatalf("user.name = %q, want Schutzbot", got)
	}
	email, err := exec.Command("git", "config", "user.email").Output()
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(email)); got != "schutzbot@gmail.com" {
		t.Fatalf("user.email = %q, want schutzbot@gmail.com", got)
	}
}

func TestErrIfUnsafeGitWorktree(t *testing.T) {
	tmp := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	config = &AppConfig{NoGit: false, DryRun: false}
	if err := errIfUnsafeGitWorktree(); err != nil {
		t.Fatalf("expected nil without git repo: %v", err)
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

	if err := errIfUnsafeGitWorktree(); err != nil {
		t.Fatalf("expected nil in clean repo: %v", err)
	}

	if err := os.WriteFile(readme, []byte("dirty\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := errIfUnsafeGitWorktree(); err == nil {
		t.Fatal("expected error when work tree has uncommitted changes")
	}

	config.NoGit = true
	if err := errIfUnsafeGitWorktree(); err != nil {
		t.Fatalf("expected nil with -no-git: %v", err)
	}
	config.NoGit = false

	config.DryRun = true
	if err := errIfUnsafeGitWorktree(); err != nil {
		t.Fatalf("expected nil with dry-run: %v", err)
	}
}
