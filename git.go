package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func perDependencyGitEnabled() bool {
	if config.NoGit || config.DryRun {
		return false
	}
	if !gitInsideWorkTree() {
		return false
	}
	if gitHasUncommittedChanges() {
		return false
	}
	return true
}

// errIfUnsafeGitWorktree returns an error when the repository has local changes
// and gobump would use git, so the user can opt out explicitly with -no-git.
func errIfUnsafeGitWorktree() error {
	if config.NoGit || config.DryRun {
		return nil
	}
	if !gitInsideWorkTree() {
		return nil
	}
	if !gitHasUncommittedChanges() {
		return nil
	}
	return fmt.Errorf("refusing to run: git work tree has uncommitted changes. Commit or stash your changes first, or pass -no-git to skip all git integration (no commits, reset, or clean)")
}

func gitHasUncommittedChanges() bool {
	outBytes, err := exec.Command("git", "status", "--porcelain").Output()
	if err != nil {
		// If status fails inside a work tree, treat as unsafe.
		return true
	}
	return strings.TrimSpace(string(outBytes)) != ""
}

func gitInsideWorkTree() bool {
	outBytes, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(outBytes)) == "true"
}

func gitRun(args ...string) error {
	if config.Verbose {
		out.Println(append([]string{"git"}, args...)...)
	}
	c := exec.Command("git", args...)
	c.Env = os.Environ()
	if config.Verbose {
		c.Stdout = out
	}
	c.Stderr = out
	return c.Run()
}

func goModSumPathsForGit() []string {
	modPath := config.GoModDst
	sumPath := strings.TrimSuffix(modPath, ".mod") + ".sum"
	paths := []string{modPath}
	if st, err := os.Stat(sumPath); err == nil && !st.IsDir() {
		paths = append(paths, sumPath)
	}
	return paths
}

func gitWorktreeDiffersFromHEAD() bool {
	paths := goModSumPathsForGit()
	args := append([]string{"diff", "--quiet", "HEAD", "--"}, paths...)
	err := exec.Command("git", args...).Run()
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true
	}
	return false
}

func gitResetHardHEAD() error {
	if err := gitRun("reset", "--hard", "HEAD"); err != nil {
		return fmt.Errorf("git reset --hard HEAD: %w", err)
	}
	if err := gitRun("clean", "-fdq"); err != nil {
		return fmt.Errorf("git clean -fdq: %w", err)
	}
	return nil
}

func gitEnsureUserIdentity() error {
	if err := gitRun("config", "user.name", config.GitUserName); err != nil {
		return fmt.Errorf("git config user.name: %w", err)
	}
	if err := gitRun("config", "user.email", config.GitUserEmail); err != nil {
		return fmt.Errorf("git config user.email: %w", err)
	}
	return nil
}

func goModTidy() error {
	if err := cmd(config.GoBinary, "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func gitCommitDependencyBump(modulePath, versionBefore, versionAfter string) error {
	if err := gitEnsureUserIdentity(); err != nil {
		return err
	}
	if err := goModTidy(); err != nil {
		return err
	}
	paths := goModSumPathsForGit()
	for _, p := range paths {
		if _, err := os.Stat(p); err != nil {
			return fmt.Errorf("git add: %w", err)
		}
	}
	absPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		absPaths = append(absPaths, abs)
	}
	topBytes, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return fmt.Errorf("git rev-parse --show-toplevel: %w", err)
	}
	top := strings.TrimSpace(string(topBytes))
	relPaths := make([]string, len(absPaths))
	for i, ap := range absPaths {
		rel, err := filepath.Rel(top, ap)
		if err != nil {
			return err
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("go.mod path %s is outside git top-level %s", ap, top)
		}
		relPaths[i] = filepath.ToSlash(rel)
	}
	addArgs := append([]string{"add", "--"}, relPaths...)
	if err := gitRun(addArgs...); err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	msg := fmt.Sprintf("chore(deps): update %s to %s", modulePath, versionAfter)
	if config.Changelog {
		msg += formatModuleChangelog(modulePath, versionBefore, versionAfter)
	}
	if err := gitRun("commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}
