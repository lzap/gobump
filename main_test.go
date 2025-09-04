package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	executeTest := func(t *testing.T, input, output string) {
		t.Run(input, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = []string{"test", "-dry-run", "-src-go-mod", input, "-dst-go-mod", output, "-exec", "echo ok", "-format", "none"}
			main()
		})
	}

	executeTestPositional := func(t *testing.T, input, output, dependency string) {
		t.Run(input, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = []string{"test", "-dry-run", "-src-go-mod", input, "-dst-go-mod", output, "-exec", "echo ok", "-format", "none", dependency}
			main()
		})
	}

	executeTestExclude := func(t *testing.T, input, output, exclude string) {
		t.Run(input, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = []string{"test", "-dry-run", "-src-go-mod", input, "-dst-go-mod", output, "-exec", "echo ok", "-format", "none", "-exclude", exclude}
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

		if strings.HasSuffix(file, "positional.in") {
			executeTestPositional(t, file, strings.Replace(file, ".in", ".out", 1), "github.com/sirupsen/logrus")
		} else if strings.HasSuffix(file, "exclude.in") {
			executeTestExclude(t, file, strings.Replace(file, ".in", ".out", 1), "github.com/sirupsen/logrus")
		} else if strings.HasSuffix(file, "exclude-no-positional.in") {
			executeTestExclude(t, file, strings.Replace(file, ".in", ".out", 1), "github.com/sirupsen/logrus")
		} else {
			executeTest(t, file, strings.Replace(file, ".in", ".out", 1))
		}
	}
}
