package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type OutputMarkdown struct {
	Destination io.Writer
	w           io.Writer
}

var _ Output = (*OutputMarkdown)(nil)

func NewOutputMarkdown(w io.Writer) *OutputMarkdown {
	return &OutputMarkdown{
		Destination: w,
		w:           w,
	}
}

func (out *OutputMarkdown) Begin(text ...any) {
	if len(text) == 0 {
		fmt.Fprintf(out.w, "## Pinned Go version dependency update\n")
		return
	}

	fmt.Fprintln(out.w, joinAny(text...))
}

func (out *OutputMarkdown) End(text ...any) {
	defer fmt.Fprintf(out.w, "\n:pretzel: *Created with [gobump](https://github.com/lzap/gobump) (%s)* :pretzel:\n", BuildID())

	if len(text) == 0 {
		return
	}

	fmt.Fprintln(out.w, joinAny(text...))
}

func (out *OutputMarkdown) Header(text string) {
	if len(text) == 0 {
		return
	}

	fmt.Fprintf(out.w, "\n### %s\n", text)
}

func (out *OutputMarkdown) BeginPreformatted(text ...any) {
	out.w = bytes.NewBuffer(nil)

	if len(text) == 0 {
		return
	}

	fmt.Fprintf(out.w, "\n<details><summary>%s</summary>\n\n```\n", joinAny(text...))
}

func (out *OutputMarkdown) EndPreformatted(text ...any) {
	out.EndPreformattedCond(true, text...)
}

func (out *OutputMarkdown) endBuffer(render bool) {
	if out.w == nil {
		return
	}

	buf := out.w.(*bytes.Buffer)
	out.w = out.Destination

	if buf.Len() > 0 && render {
		fmt.Fprint(out.w, buf.String())
	}

	if render {
		fmt.Fprintf(out.w, "```\n</details>\n")
	}
}

func (out *OutputMarkdown) EndPreformattedCond(render bool, text ...any) {
	defer out.endBuffer(render)

	if len(text) == 0 {
		return
	}

	fmt.Fprintln(out.w, joinAny(text...))
}

func (out *OutputMarkdown) Write(buf []byte) (int, error) {
	return out.w.Write(buf)
}

func (out *OutputMarkdown) Println(text ...string) {
	if len(text) == 0 {
		return
	}

	fmt.Fprintln(out.w, strings.Join(text, " "))
}

func (out *OutputMarkdown) Error(str ...string) {
	fmt.Fprintln(out.w, strings.Join(str, " "))
}

func (out *OutputMarkdown) Fatal(msg string, code ...int) {
	fmt.Fprintln(out.w, msg)

	if len(code) == 0 {
		os.Exit(1)
	}

	os.Exit(code[0])
}

func (out *OutputMarkdown) PrintSummary(results []Result) {
	fmt.Fprintf(out.w, "\n## Summary\n\n")
	fmt.Fprintln(out.w, "|Module|Action|Min Go|Before|After|")
	fmt.Fprintln(out.w, "|---|---|---|---|---|")

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
