package main

import (
	"fmt"
	"os"
	"strings"
)

type OutputMarkdown struct{}

var _ Output = (*OutputMarkdown)(nil)

func (out *OutputMarkdown) Begin(text ...any) {
	if len(text) == 0 {
		fmt.Printf("## Pinned Go version dependency update\n")
		fmt.Println("")
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputMarkdown) End(text ...any) {
	if len(text) == 0 {
		fmt.Println("")
		fmt.Printf(":pretzel: *Created with [gobump](https://github.com/lzap/gobump) (%s)* :pretzel:\n", BuildID())
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputMarkdown) Header(text string) {
	if len(text) == 0 {
		return
	}

	fmt.Printf("\n### %s\n\n", text)
}

func (out *OutputMarkdown) BeginPreformatted(text ...any) {
	if len(text) == 0 {
		return
	}

	fmt.Printf("\n<details><summary>%s</summary>\n\n```\n", joinAny(text...))
}

func (out *OutputMarkdown) EndPreformatted(text ...any) {
	if len(text) == 0 {
		fmt.Printf("```\n</details>\n")
		return
	}

	fmt.Println(joinAny(text...))
}

func (out *OutputMarkdown) Write(buf []byte) (int, error) {
	return os.Stdout.Write(buf)
}

func (out *OutputMarkdown) Println(text ...string) {
	if len(text) == 0 {
		return
	}

	fmt.Println(strings.Join(text, " "))
}

func (out *OutputMarkdown) Error(str ...string) {
	fmt.Println(color(strings.Join(str, " "), ColorRed))
}

func (out *OutputMarkdown) Fatal(msg string, code ...int) {
	fmt.Println(msg)

	if len(code) == 0 {
		os.Exit(1)
	}

	os.Exit(code[0])
}

func (out *OutputMarkdown) PrintSummary(results []Result) {
	fmt.Printf("\n## Summary\n\n")
	fmt.Println("|Module|Action|Min Go|Before|After|")
	fmt.Println("|---|---|---|---|---|")

	for _, r := range results {
		action := "skipped"
		if r.Success {
			if r.VersionAfter == r.VersionBefore {
				action = "no action"
			} else {
				action = "upgraded"
			}
		}
		out.Println(strings.Join([]string{
			r.ModulePath,
			action,
			strOrDash(r.MinGoVersion),
			strOrDash(r.VersionBefore),
			strOrDash(r.VersionAfter),
		}, "|"))
	}
}
