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
	if config.SingleCommit || config.DryRun {
		return false
	}
	if !gitInsideWorkTree() {
		return false
	}
	if gitHasStagedChanges() {
		return false
	}
	return true
}

func gitInsideWorkTree() bool {
	outBytes, err := exec.Command("git", "rev-parse", "--is-inside-work-tree").Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(outBytes)) == "true"
}

func gitHasStagedChanges() bool {
	err := exec.Command("git", "diff", "--cached", "--quiet").Run()
	if err == nil {
		return false
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return true
	}
	// Treat unexpected git errors as "unsafe": do not enable per-dependency git mode.
	return true
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
	return gitRun("reset", "--hard", "HEAD")
}

func gitCommitDependencyBump(modulePath, version string) error {
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
	msg := fmt.Sprintf("chore(deps): update %s to %s", modulePath, version)
	if err := gitRun("commit", "-m", msg); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}
