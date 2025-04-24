package main

import (
	"fmt"
	"os"
	"strings"
)

type OutputConsole struct{}
var _ Output = (*OutputConsole)(nil)

func (out *OutputConsole) Begin(text ...any) {
	if len(text) == 0 {
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputConsole) End(text ...any) {
	if len(text) == 0 {
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputConsole) Header(text string) {
	if len(text) == 0 {
		return
	}

	fmt.Println(color(text, ColorBold))
}

func (out *OutputConsole) BeginPreformatted(text ...any) {
	if len(text) == 0 {
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputConsole) EndPreformatted(text ...any) {
	if len(text) == 0 {
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputConsole) Write(buf []byte) (int, error) {
	return os.Stdout.Write(buf)
}

func (out *OutputConsole) Println(text ...string) {
	if len(text) == 0 {
		return
	}

	fmt.Println(strings.Join(text, " "))
}

func (out *OutputConsole) Error(str ...string) {
	fmt.Fprintln(os.Stderr, color(strings.Join(str, " "), ColorRed))
}

func (out *OutputConsole) Fatal(msg string, code ...int) {
	fmt.Fprintln(os.Stderr, msg)

	if len(code) == 0 {
		os.Exit(1)
	}

	os.Exit(code[0])
}

func (out *OutputConsole) PrintSummary(results []Result) {
	out.Println(color("summary:", ColorBold))

	for _, r := range results {
		action := "skipped"
		if r.Success {
			if r.VersionAfter == r.VersionBefore {
				action = "no action"
			} else {
				action = "upgraded"
			}
		}
		if r.VersionAfter != "" && r.VersionAfter != r.VersionBefore && action != "skipped" {
			out.Println(r.ModulePath, action, fmt.Sprintf("mingo:%s", r.MinGoVersion), r.VersionBefore, "->", r.VersionAfter)
		} else {
			out.Println(r.ModulePath, action, fmt.Sprintf("mingo:%s", r.MinGoVersion))

		}
	}
}
