package main

import (
	"strconv"
	"strings"
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

func color(input any, color ...string) string {
	var s string
	c := ""
	for i := range color {
		c = c + color[i]
	}
	switch v := input.(type) {
	case int:
		s = c + strconv.Itoa(v) + ColorReset
	case bool:
		s = c + strconv.FormatBool(v) + ColorReset
	case []string:
		s = c + strings.Join(v, ", ") + ColorReset
	case string:
		s = c + v + ColorReset
	default:
		panic("unsupported color type")
	}
	return s
}
