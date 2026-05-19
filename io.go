package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	ERR_READ  = 2
	ERR_WRITE = 3
	ERR_PARSE = 4
	ERR_CMD   = 5
	ERR_GIT   = 6
)

// cmd runs a subprocess; when verbose, echoes the command and streams stdout/stderr to out
// (intended for go get and -exec inside a preformatted block).
func cmd(name string, args ...string) error {
	return runCmd(name, args, true)
}

// cmdQuiet runs a subprocess without writing to out (e.g. go mod tidy before git commit).
func cmdQuiet(name string, args ...string) error {
	return runCmd(name, args, false)
}

func runCmd(name string, args []string, logOutput bool) error {
	if logOutput && config.Verbose {
		out.Println(name, strings.Join(args, " "))
	}
	c := exec.Command(name, args...)
	c.Env = os.Environ()
	if logOutput && config.Verbose {
		c.Stdout = out
		c.Stderr = out
	}
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

var ErrCmd = fmt.Errorf("command error")

func cmds(str string) error {
	parts := strings.Fields(str)
	if len(parts) == 0 {
		return fmt.Errorf("%w: no command", ErrCmd)
	}

	if len(parts) == 1 {
		return cmd(parts[0])
	}

	p1 := parts[0]
	p2 := parts[1:]
	return cmd(p1, p2...)
}
