package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type commaSeparatedStringSlice []string

func (i *commaSeparatedStringSlice) String() string {
	return strings.Join(*i, ",")
}

func (i *commaSeparatedStringSlice) Set(value string) error {
	if value != "" {
		*i = strings.Split(value, ",")
	}
	return nil
}

type stringSlice []string

func (i *stringSlice) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *stringSlice) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// AppConfig holds the application configuration
type AppConfig struct {
	DryRun        bool
	Verbose       bool
	Format        string
	GoModSrc      string
	GoModDst      string
	Retries       int
	Commands      stringSlice
	GoBinary      string
	Changelog     bool
	ChangelogDest string
	Dependencies  []string
	Exclude       commaSeparatedStringSlice
}

var config *AppConfig

func isCI() bool {
	return os.Getenv("GITHUB_ACTIONS")+os.Getenv("GITLAB_CI")+os.Getenv("CIRCLECI") != ""
}

// InitConfig initializes the global configuration object.
func InitConfig() {
	config = &AppConfig{}

	goBinary := os.Getenv("GOVERSION")
	if goBinary == "" {
		goBinary = "go"
	}
	config.GoBinary = goBinary

	defaultFormat := "console"
	defaultVerbose := false
	if isCI() {
		defaultFormat = "markdown"
		defaultVerbose = true
	}

	var commands stringSlice
	var exclude commaSeparatedStringSlice
	flag.BoolVar(&config.DryRun, "dry-run", false, "revert to original go.mod after running")
	flag.BoolVar(&config.Verbose, "verbose", defaultVerbose, "print more information including stderr of executed commands")
	flag.Var(&commands, "exec", "exec command for each individual bump, can be used multiple times")
	flag.Var(&exclude, "exclude", "comma-separated list of modules to exclude from update")
	flag.StringVar(&config.Format, "format", defaultFormat, "output format (console, markdown, none)")
	flag.StringVar(&config.GoModSrc, "src-go-mod", "go.mod", "path to go.mod source file (default: go.mod)")
	flag.StringVar(&config.GoModDst, "dst-go-mod", "go.mod", "path to go.mod destination file (default: go.mod)")
	flag.IntVar(&config.Retries, "retries", 5, "number of downgrade retries for each module (default: 5)")
	flag.BoolVar(&config.Changelog, "changelog", false, "print git changelog of all updated modules")
	flag.StringVar(&config.ChangelogDest, "changelog-dest", "stdout", "Destination of the changelog (\"stdout\", \"gist\" or a filename)")
	flag.Parse()

	config.Commands = commands
	config.Dependencies = flag.Args()
	config.Exclude = exclude
}
