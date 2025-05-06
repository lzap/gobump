package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	executeTest := func(t *testing.T, input, output string) {
		t.Run(input, func(t *testing.T) {
			os.Args = []string{"test", "-dry-run", "-src-go-mod", input, "-dst-go-mod", output, "-exec", "echo ok", "-format", "none"}
			main()
		})
	}

	// Clean up the project go.mod as indirect dependencies will be added
	defer func() {
		t.Log("Cleaning up go.mod")
		err := exec.Command("go", "mod", "tidy").Run()
		if err != nil {
			t.Fatalf("failed to tidy go.mod: %v", err)
		}
	}()

	files, err := filepath.Glob("testdata/*.in")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		if s, err := os.Stat(file); err != nil || s.IsDir() {
			continue
		}

		executeTest(t, file, strings.Replace(file, ".in", ".out", 1))
	}
}
