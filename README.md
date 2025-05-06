# gobump

Pins Go version when bumping dependencies.

A simple tool which upgrades all direct dependencies one by one ensuring the Go version statement in `go.mod` is never touched. This is useful if your build infrastructure lags behind the latest and greatest Go version and you are unable to upgrade yet, for example when using Red Hat Go Toolset for UBI.

It solves the following problem of `go get -u` pushing for the latest Go version, even if you explicitly use a specific version of Go:

```
$ go1.21.0 get -u golang.org/x/tools@latest
go: upgraded go 1.21.0 => 1.22.0
```

## Installation

```
go install github.com/lzap/gobump@latest
```

## Usage

```
cd ~/your_project
```

The utility currently does not take any arguments:

```
gobump
```

Example output:

```
go get golang.org/x/sys@latest
go get golang.org/x/tools@latest
go: upgraded go 1.21.0 => 1.22.0
go: upgraded toolchain go1.22.5 => go1.22.7
go: upgraded golang.org/x/mod v0.20.0 => v0.21.0
go: upgraded golang.org/x/tools v0.24.0 => v0.26.0
upgrade changes required Go version, reverting go.mod
go get google.golang.org/api@latest
go get cloud.google.com/go/compute@latest
go get cloud.google.com/go/storage@latest
```

The above command upgraded all dependencies except `golang.org/x/tools` which would have increased Go requirement.

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
          #setup_go: true
          #exec: "go test ./..."
          #tidy: true
          #pr: true
          token: ${{ secrets.GITHUB_TOKEN }}
          #labels: "gobump"
```

Action input:

* `exec`: optional command to execute for each dependency update
* `exec2`: second optional command to execute for each dependency update
* `tidy`: set to `false` to avoid executing `go mod tidy` after `gobump`
* `setup_go`: set to `false` to avoid `setup-go` action (e.g. when container with Go is used)
* `pr`: set to `false` to avoid creation of a PR
* `token`: github token
* `labels`: comma-separated github PR labels

## How it works

* Loads project `go.mod` and stores it in memory.
* For each direct dependency, it performs `go get -u DEPENDENCY@latest`.
* If the `go get` command fails or modifies Go version in `go.mod`, it reverts to the last version of `go.mod`.
* If user provides one or more optional `exec` argument, it executes it and if any of the commands fails, it reverts to the last `go.mod` version too.
* Repeats for every other direct dependency.

## Executing build or tests

For every single updated dependency, it is possible to run one or more commands to ensure the project builds or tests are passing. Use `-exec` option multiple times to do that, when such command returns non-zero value it is considered as a failure and that update is rolled back.

```
gobump -exec "go build ./..." -exec "go test ./..."
```

Commands are not executed via shell.

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

It is possible to use different binary than `go`, set `GOVERSION=go1.21.0` environment variable to use a different Go version that is available through `PATH`.
