package main

import (
	"fmt"
	"os"
	"runtime/debug"
)

var (
	out Output
)

func main() {
	InitConfig()

	if config.Version {
		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Printf("%s %s\n", info.Main.Version, info.Main.Sum)
			os.Exit(0)
		}
		fmt.Println("(devel) HEAD")
		os.Exit(0)
	}

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

	original, err := parseMod(config.GoModSrc)
	if err != nil {
		out.Fatal(err.Error(), ERR_PARSE)
	}

	defer func() {
		if config.DryRun {
			if err := saveMod(config.GoModDst, original); err != nil {
				out.Fatal(err.Error(), ERR_WRITE)
			}
		}
	}()

	results := process(original)

	out.PrintSummary(results)
	if config.Changelog {
		PrintChangelogs(results)
	}
}
