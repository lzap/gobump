# gobump

Pins Go version when bumping dependencies.

A simple tool that upgrades all direct dependencies one by one, ensuring the Go version statement in `go.mod` is never touched. This is useful if your build infrastructure lags behind the latest Go version and you are unable to upgrade, for example, when using Go from Linux distribution packages or a container runtime like the Red Hat Go Toolset for UBI.

## The problem

When `go get -u ./...` is issued, at some point `go X.YY` in `go.mod` will be upgraded with the following message:

```
$ go get -u ./...
go: upgraded go 1.21.0 => 1.22.0
```

Using an explicit version of the Go binary does not change anything:

```
$ go1.21.0 get -u ./...
go: upgraded go 1.21.0 => 1.22.0
```

Starting from Go 1.21, the toolchain feature was added to solve some problems with tool versioning. It also skips upgrades when the toolchain version is explicitly set, but it has a different problem. When a single dependency cannot be upgraded, it skips the entire upgrade transaction, leading to no upgrades.

In the following scenario, the `github.com/google/go-cmp` package could have been upgraded as it was working on Go 1.21 at the time; however, nothing was upgraded:

```
$ GOTOOLCHAIN=go1.21.0 go get -u ./...
go: golang.org/x/mod@v0.24.0 requires go >= 1.23.0 (running go 1.21.0; GOTOOLCHAIN=go1.21.0)
go: golang.org/x/sys@v0.33.0 requires go >= 1.23.0 (running go 1.21.0; GOTOOLCHAIN=go1.21.0)
go: golang.org/x/term@v0.32.0 requires go >= 1.23.0 (running go 1.21.0; GOTOOLCHAIN=go1.21.0)
```

Only when dependencies are upgraded one by one, it works:

```
$ GOTOOLCHAIN=go1.21.0 go get -u github.com/google/go-cmp
go: upgraded github.com/google/go-cmp v0.3.0 => v0.7.0
```

This utility upgrades dependencies one by one, optionally running `go build` or `go test` when configured to ensure the project builds. This is useful for mass-upgrading dependencies to isolate those that break tests.

When a dependency cannot be upgraded (or optional commands fail to execute, e.g., `go test`), it retries several times (configurable) with older versions of the module until it succeeds. By default, it goes back up to 5 versions.

## Installation

```
go install github.com/lzap/gobump@latest
```

## Usage

```
  -changelog
    	print git changelog of all updated modules (default true)
  -changelog-gist string
    	GitHub token to create a Gist with the changelog
  -dry-run
    	revert to original go.mod after running
  -dst-go-mod string
    	path to go.mod destination file (default: go.mod) (default "go.mod")
  -exec value
    	exec command for each individual bump, can be used multiple times
  -exclude string
    	comma-separated list of modules to exclude from update
  -format string
    	output format (console, markdown, none) (default "console")
  -retries int
    	number of downgrade retries for each module (default: 5) (default 5)
  -src-go-mod string
    	path to go.mod source file (default: go.mod) (default "go.mod")
  -verbose
    	print more information including stderr of executed commands
```

The utility can also take one or more module paths as positional arguments. When provided, only those dependencies will be updated, ignoring others. This is useful for targeting specific dependency updates.

When no arguments are provided, `gobump` updates all direct dependencies. In this mode, it is important to always specify the `GOTOOLCHAIN` variable to match the version in your project's `go.mod` file. This is the version you want to pin and prevent from being upgraded.

```
GOTOOLCHAIN=go1.22.0 gobump
```

Example output:

```
go get github.com/google/go-cmp@latest
go get golang.org/x/mod@latest
go: golang.org/x/mod@latest: golang.org/x/mod@v0.24.0 requires go >= 1.23.0 (running go 1.22.0; GOTOOLCHAIN=go1.22.0)
upgrade unsuccessful, reverting go.mod
go get golang.org/x/term@latest
go: golang.org/x/term@latest: golang.org/x/term@v0.32.0 requires go >= 1.23.0 (running go 1.22.0; GOTOOLCHAIN=go1.22.0)
upgrade unsuccessful, reverting go.mod

Summary:
github.com/google/go-cmp keep
golang.org/x/mod err
golang.org/x/term err
```

Summary legend:

* `keep`: version is kept (no update available)
* `update`: module updated to newer version
* `err`: there was an error during the update; either the required Go version is too high, one of the `exec` commands failed, or another error occurred
* `excluded`: module was excluded from update

## GitHub Action

The GitHub Action executes `gobump`, then performs `go mod tidy` and files an update PR to the project. Example PR: https://github.com/lzap/gobump/pull/7

Example action:

```
name: "Weekly gobump"
on:
  schedule:
    - cron: '13 13 * * SUN'
  workflow_dispatch:

jobs:
  bump-deps-ubuntu:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout repository
        uses: actions/checkout@v4

      - name: Run gobump-deps action
        uses: lzap/gobump@main
        with:
          go_version: "1.22.0"
          token: ${{ secrets.GITHUB_TOKEN }}
```

Action inputs:

* `go_version`: **The version to use** and pin the project to (defaults to stable, but always set this).
* `setup_go`: Set to `false` to avoid the `setup-go` action (e.g., when a container with a specific Go version is used).
* `exec`: An optional command to execute for each dependency update.
* `exec2`: A second optional command to execute for each dependency update.
* `exclude`: A comma-separated list of modules to exclude from the update.
* `tidy`: Set to `false` to avoid executing `go mod tidy` after `gobump`.
* `exec_pr`: An optional command to execute before a PR is made.
* `pr`: Set to `false` to avoid the creation of a PR.
* `token`: The GitHub token.
* `labels`: Comma-separated GitHub PR labels.

Tip: When building or testing in a container, use `-buildvcs=false` to avoid `git: detected dubious ownership in repository` permissions errors. Alternatively, set the `git config --system --add safe.directory /path` config option.

## How it works

* Loads the project's `go.mod` and stores it in memory.
* For each direct dependency, it performs `go get -u DEPENDENCY@V`, where `V` is one of the 5 latest versions reported by [proxy.golang.org](https://proxy.golang.org/github.com/lzap/gobump/@v/list).
* If the `go get` command fails (e.g., `GOTOOLCHAIN` is set) or modifies the Go version in `go.mod`, it reverts to the last version of `go.mod` and tries again with a lower version up to N times (configurable, defaults to 5).
* If and only if a module succeeds in updating and one or more optional `exec` arguments are passed, it executes them. If any of the commands fail, it reverts to the last `go.mod` version.
* Repeats for every other direct dependency.

It is recommended to set `GOTOOLCHAIN` to an explicit Go version to speed up the failure of `go get` because, with a specific Go version, it immediately fails and does not even attempt to download and install packages, which would lead to a `go.mod` change.

## Custom commands

For every updated dependency, it is possible to run one or more commands to ensure the project builds or tests are passing. Use the `-exec` option multiple times to do that. When such a command returns a non-zero value, it is considered a failure, and that update is rolled back.

```
gobump -exec "go build ./..." -exec "go test ./..."
```

Commands are not executed via a shell. Subprocesses will inherit the `GOTOOLCHAIN` setting, so it is fine to use just the `go` command or any version of Go later than 1.21, and it will pick up the correct toolchain.

## Ambiguous imports

Sometimes, even `gobump` does not help, specifically with ambiguous imports in transient dependencies:

```
go: mypackage imports
 cloud.google.com/go/storage imports
  google.golang.org/grpc/stats/opentelemetry: ambiguous import: found package google.golang.org/grpc/stats/opentelemetry in multiple modules:
  google.golang.org/grpc v1.67.3 (/home/lzap/go/pkg/mod/google.golang.org/grpc@v1.67.3/stats/opentelemetry)
  google.golang.org/grpc/stats/opentelemetry v0.0.0-20240907200651-3ffb98b2c93a (/home/lzap/go/pkg/mo
```

In this case, use this trick:

```
go get google.golang.org/grpc/stats/opentelemetry@none
```

## Configuration

It is possible to use a different binary than `go`; set the `GOVERSION=go1.21.0` environment variable to use a different Go version that is available through the `PATH`. But the recommended way of using specific Go tooling is via the `GOTOOLCHAIN` variable.

## Limitations

When a module's `latest` version cannot be upgraded, the tool currently does not attempt to lower its version and find the latest that works. This is a feature that will be implemented later.

## Discussion

I created a post on Reddit if you need to reach out: https://www.reddit.com/r/golang/comments/1kfypws/gobump_update_dependencies_with_pinned_go_version/

Alternatively, create an issue if you find a problem or need a new feature.
