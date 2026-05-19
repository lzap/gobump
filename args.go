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
	Version       bool
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
	NoGit         bool
	GitUserName   string
	GitUserEmail  string
	ModuleProxy   string
	FailOnError   bool
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
	flag.BoolVar(&config.Version, "version", false, "print Go binary debug info")
	flag.BoolVar(&config.DryRun, "dry-run", false, "revert to original go.mod after running")
	flag.BoolVar(&config.Verbose, "verbose", defaultVerbose, "echo go get and -exec command output (markdown: inside detail blocks); does not log git operations in git mode")
	flag.Var(&commands, "exec", "exec command for each individual bump, can be used multiple times")
	flag.Var(&exclude, "exclude", "comma-separated list of modules to exclude from update")
	flag.StringVar(&config.Format, "format", defaultFormat, "output format (console, markdown, none)")
	flag.StringVar(&config.GoModSrc, "src-go-mod", "go.mod", "path to go.mod source file (default: go.mod)")
	flag.StringVar(&config.GoModDst, "dst-go-mod", "go.mod", "path to go.mod destination file (default: go.mod)")
	flag.IntVar(&config.Retries, "retries", 5, "number of downgrade retries for each module (default: 5)")
	flag.BoolVar(&config.Changelog, "changelog", false, "fetch upstream git changelog for each updated module (embedded in per-dependency commit messages when git integration is enabled; otherwise aggregated at end per -changelog-dest)")
	flag.StringVar(&config.ChangelogDest, "changelog-dest", "stdout", "with -changelog and -no-git (or no usable git work tree): write aggregated changelogs to stdout (default), a file path, or \"gist\"; ignored when changelogs are committed per dependency")
	flag.BoolVar(&config.NoGit, "no-git", false, "if true, skip all git operations (no per-dependency commits or reset/clean on failure)")
	flag.StringVar(&config.GitUserName, "user-name", "Schutzbot", "git user.name for per-dependency commits (local repo config)")
	flag.StringVar(&config.GitUserEmail, "user-email", "schutzbot@gmail.com", "git user.email for per-dependency commits (local repo config)")
	flag.StringVar(&config.ModuleProxy, "proxy", "", "module proxy base URL (default: first usable $GOPROXY entry, else https://proxy.golang.org)")
	flag.BoolVar(&config.FailOnError, "fail-on-error", false, "exit with status 1 if any non-excluded module failed to update")
	flag.Parse()

	config.Commands = commands
	config.Dependencies = flag.Args()
	config.Exclude = exclude
}
