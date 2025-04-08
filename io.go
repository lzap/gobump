package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	_         = iota
	_         = iota
	ERR_READ  = iota
	ERR_WRITE = iota
	ERR_PARSE = iota
	ERR_CMD   = iota
)

func die(msg string, code ...int) {
	fmt.Fprintln(os.Stderr, msg)
	if len(code) == 0 {
		os.Exit(1)
	}
	os.Exit(code[0])
}

func cmd(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	if verbose {
		c.Stdout = os.Stdout
	} else {
		c.Stdout = nil
	}
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	fmt.Println(cmd, strings.Join(args, " "))
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

func println(str ...string) {
	fmt.Println(strings.Join(str, " "))
}

func printerr(str ...string) {
	fmt.Fprintln(os.Stderr, strings.Join(str, " "))
}
