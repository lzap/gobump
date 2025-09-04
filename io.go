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

func cmd(cmd string, args ...string) error {
	if config.Verbose {
		out.Println(cmd, strings.Join(args, " "))
	}
	c := exec.Command(cmd, args...)
	c.Env = os.Environ()
	if config.Verbose {
		c.Stdout = out
		c.Stderr = out
	} else {
		c.Stdout = nil
		c.Stderr = nil
	}
	c.Stderr = out
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
