package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/mod/modfile"
)

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	dryRun   bool
	verbose  bool
	format   string
	gomodsrc string
	gomoddst string
	retries  int

	out Output
)

func main() {
	goBinary := os.Getenv("GOVERSION")
	if goBinary == "" {
		goBinary = "go"
	}

	defaultFormat := "console"
	defaultVerbose := false
	if os.Getenv("GITHUB_ACTIONS")+os.Getenv("GITLAB_CI")+os.Getenv("CIRCLECI") != "" {
		defaultFormat = "markdown"
		defaultVerbose = true
	}

	var commands stringSlice
	flag.BoolVar(&dryRun, "dry-run", false, "revert to original go.mod after running")
	flag.BoolVar(&verbose, "verbose", defaultVerbose, "print more information including stderr of executed commands")
	flag.Var(&commands, "exec", "exec command for each individual bump, can be used multiple times")
	flag.StringVar(&format, "format", defaultFormat, "output format (console, markdown, none)")
	flag.StringVar(&gomodsrc, "src-go-mod", "go.mod", "path to go.mod source file (default: go.mod)")
	flag.StringVar(&gomoddst, "dst-go-mod", "go.mod", "path to go.mod destination file (default: go.mod)")
	flag.IntVar(&retries, "retries", 5, "number of downgrade retries for each module (default: 5)")
	flag.Parse()

	switch format {
	case "markdown":
		out = NewOutputMarkdown(os.Stdout)
	case "console":
		out = &OutputConsole{}
	default:
		out = &OutputNone{}
	}

	out.Begin()
	defer out.End()

	original := parse(gomodsrc)
	modules := []*modfile.File{original}
	var results []Result

	defer func() {
		if dryRun {
			save(gomoddst, original)
		}
	}()

	proxy := NewGoProxy("")

	var newMod *modfile.File
	for _, r := range original.Require {
		if !r.Indirect {
			success := true
			lastMod := modules[len(modules)-1]
			out.BeginPreformatted(goBinary, "get", r.Mod.Path)
			versions, err := proxy.FetchVersions(r.Mod.Path)
			if err != nil {
				out.Error("failed to fetch versions:", err.Error())
				out.EndPreformatted(false)
				continue
			}
			for vi, version := range versions {
				if vi >= retries {
					out.Error("too many failed attempts, giving up")
					break
				}
				err := cmd(goBinary, "get", r.Mod.Path+"@"+version.Version)
				newMod = parse(gomodsrc)
				if err != nil {
					out.Error("upgrade unsuccessful, reverting go.mod")
					save(gomoddst, lastMod)
					success = false
				} else if strings.TrimSuffix(lastMod.Go.Version, ".0") != strings.TrimSuffix(newMod.Go.Version, ".0") {
					out.Error("upgrade changes required Go version, reverting go.mod")
					save(gomoddst, lastMod)
					success = false
				} else {
					success = true
					break
				}
			}
			out.EndPreformattedCond(!success)

			if success {
				for _, c := range commands {
					if c == "" {
						continue
					}

					out.BeginPreformatted(c)
					if err := cmds(c); err != nil {
						out.Error("tests failed, reverting go.mod")
						save(gomoddst, lastMod)
						success = false
					}
					out.EndPreformattedCond(!success)
				}
			}

			mi := slices.IndexFunc(newMod.Require, func(re *modfile.Require) bool {
				return re.Mod.Path == r.Mod.Path
			})
			newRequire := newMod.Require[mi]

			result := Result{
				ModulePath:    r.Mod.Path,
				VersionBefore: r.Mod.Version,
				VersionAfter:  newRequire.Mod.Version,
			}

			if success {
				modules = append(modules, newMod)
				result.Success = true
			} else {
				result.Success = false
			}

			results = append(results, result)
		}
	}

	slices.SortFunc(results, func(a, b Result) int {
		return strings.Compare(a.ModulePath, b.ModulePath)
	})

	out.PrintSummary(results)
}
