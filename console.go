package main

import (
	"os"

	"golang.org/x/term"
)

var (
	ColorReset     = "\033[0m"
	ColorBlack     = "\033[30m"
	ColorRed       = "\033[31m"
	ColorGreen     = "\033[32m"
	ColorYellow    = "\033[33m"
	ColorBlue      = "\033[34m"
	ColorMagenta   = "\033[35m"
	ColorCyan      = "\033[36m"
	ColorGray      = "\033[37m"
	ColorWhite     = "\033[97m"
	ColorBold      = "\033[1m"
	ColorItalic    = "\033[3m"
	ColorUnderline = "\033[4m"
	ColorInvert    = "\033[7m"
)

func color(input string, color ...string) string {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		return input
	}

	c := ""
	for i := range color {
		c = c + color[i]
	}
	return c + input + ColorReset
}
