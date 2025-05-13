# gobump

Pins Go version when bumping dependencies.

A simple tool which upgrades all direct dependencies one by one ensuring the Go version statement in `go.mod` is never touched. This is useful if your build infrastructure lags behind the latest and greatest Go version and you are unable to upgrade yet, for example when using Go from Linux distribution packages or when using a container runtime like Red Hat Go Toolset for UBI.

## The problem

When `go get -u ./...` is issued, at some point `go X.YY` in `go.mod` will be upgraded with the following message:

```
$ go get -u ./...
go: upgraded go 1.21.0 => 1.22.0
```

Using explicit version of Go binary does not change a thing:

```
$ go1.21.0 get -u ./...
go: upgraded go 1.21.0 => 1.22.0
```

Starting from Go 1.21, Toolchain feature was added which tries to solve some of the problems with tool versioning and also skips upgrade when toolchain version is explicitly set, but it has a different problem. When a single dependency cannot be upgraded it skips the whole upgrade transaction leading to no upgrades.

In the following scenario, package `github.com/google/go-cmp` could be upgraded as it was working on Go 1.21 at the time, however, nothing was upgraded:

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

This is what this utility does, it upgrades dependencies one by one optionally running `go build` or `go test` when configured to ensure the project builds. This is useful for mass-upgrade of dependencies to isolate those which break tests.

When a dependency cannot be upgraded (or optional command(s) fail to execute e.g. `go test`), it retries several times (configurable) with older versions of the module until it succeeds. By default it goes back up to 5 versions.

## Installation

```
go install github.com/lzap/gobump@latest
```

## Usage

```
  -dry-run
        revert to original go.mod after running
  -dst-go-mod string
        path to go.mod destination file (default: go.mod) (default "go.mod")
  -exec value
        exec command for each individual bump, can be used multiple times
  -format string
        output format (console, markdown, none) (default "console")
  -retries int
        number of downgrade retries for each module (default: 5) (default 5)
  -src-go-mod string
        path to go.mod source file (default: go.mod) (default "go.mod")
  -verbose
        print more information including stderr of executed commands
```

The utility currently does not take any arguments, but it is important to always specify GOTOOLCHAIN variable. It must match the version in `go.mod` of the project and it is the version you want to pin and never update.

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
* `err`: there was an error during update, either required Go version is too high or one of the `exec` commands failed or other error

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
        uses: lzap/gobump@v1
        with:
          go_version: "1.22.0"
          #setup_go: true
          #exec: "go test -buildvcs=false ./..."
          #tidy: true
          #pr: true
          token: ${{ secrets.GITHUB_TOKEN }}
          #labels: "gobump"
```

Action input:

* `go_version`: **the version to use** and thus pin the project to (defaults to stable but always set this one)
* `setup_go`: set to `false` to avoid `setup-go` action (e.g. when container with specific Go is used)
* `exec`: optional command to execute for each dependency update
* `exec2`: second optional command to execute for each dependency update
* `tidy`: set to `false` to avoid executing `go mod tidy` after `gobump`
* `exec_pr`: optional command to execute before PR is made
* `pr`: set to `false` to avoid creation of a PR
* `token`: github token
* `labels`: comma-separated github PR labels

Tip: When building or testing in a container, use `-buildvcs=false` to avoid `git: detected dubious ownership in repository` permissions errors. Alternatively, set `git config --system --add safe.directory /path` config option.

## How it works

* Loads project `go.mod` and stores it in memory.
* For each direct dependency, it performs `go get -u DEPENDENCY@V` where `V` is one of the 5 latest versions reported by [proxy.golang.org](https://proxy.golang.org/github.com/lzap/gobump/@v/list).
* If the `go get` command fails (e.g. `GOTOOLCHAIN` is set) or modifies Go version in `go.mod`, it reverts to the last version of `go.mod` and tries again with lower version up to N tries (configurable, by default it is 5).
* If and only if a module succeeds to update and one or more optional `exec` arguments are passed, it executes them and if any of the commands fails, it reverts to the last `go.mod` version.
* Repeats for every other direct dependency.

It is recommended to set `GOTOOLCHAIN` to explicit Go version to speed up failure of `go get` because with specific Go version it immediately fails and does not even attempt to download and install packages leading to `go.mod` change.

## Custom commands

For every single updated dependency, it is possible to run one or more commands to ensure the project builds or tests are passing. Use `-exec` option multiple times to do that, when such command returns non-zero value it is considered as a failure and that update is rolled back.

```
gobump -exec "go build ./..." -exec "go test ./..."
```

Commands are not executed via shell. Subprocesses will inherit the `GOTOOLCHAIN` setting so it is fine to use just `go` command or any version of Go later than 1.21 and it will pickup the correct toolchain.

## Ambiguous imports

Sometimes, even `gobump` does not help. Specifically with ambiguous imports in transient dependencies:

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

It is possible to use different binary than `go`, set `GOVERSION=go1.21.0` environment variable to use a different Go version that is available through `PATH`. But the recommended way of using specific Go tooling is via `GOTOOLCHAIN` variable.

## Limitations

When module `latest` version cannot be upgraded, the tool currently does not attempt to lower its version and find the latest that works. This is a feature that will be implemented later.

## Discussion

I created a post on reddit if you need to reach out: https://www.reddit.com/r/golang/comments/1kfypws/gobump_update_dependencies_with_pinned_go_version/

Alternatively, create an issue if you find a problem or need a feature.
