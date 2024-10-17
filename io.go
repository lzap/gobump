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
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	fmt.Println(cmd, strings.Join(args, " "))
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func print(str ...string) {
	fmt.Println(strings.Join(str, " "))
}

func printerr(str ...string) {
	fmt.Fprintln(os.Stderr, strings.Join(str, " "))
}
