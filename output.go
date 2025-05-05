package main

import (
	"fmt"
	"strings"
)

type Output interface {
	Begin(text ...any)
	Header(text string)
	BeginPreformatted(text ...any)
	EndPreformatted(text ...any)
	EndPreformattedCond(render bool, text ...any)
	End(text ...any)
	Error(str ...string)
	Fatal(msg string, code ...int)
	Write(buf []byte) (int, error)
	Println(text ...string)
	PrintSummary(results []Result)
}

func joinAny(text ...any) string {
	if len(text) == 0 {
		return ""
	}

	var str []string
	for _, t := range text {
		str = append(str, fmt.Sprint(t))
	}
	return strings.Join(str, " ")
}

func strOrDash(str string) string {
	if str == "" {
		return "-"
	}
	return str
}
