package main

import (
	"os"
)

var (
	out Output
)

func main() {
	ParseArgs()

	switch config.Format {
	case "markdown":
		out = NewOutputMarkdown(os.Stdout)
	case "console":
		out = &OutputConsole{}
	default:
		out = &OutputNone{}
	}

	out.Begin()
	defer out.End()

	original := parseMod(config.GoModSrc)

	defer func() {
		if config.DryRun {
			saveMod(config.GoModDst, original)
		}
	}()

	results := process(original)

	parseMod(config.GoModSrc)

	out.PrintSummary(results)
}
