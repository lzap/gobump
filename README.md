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

## How it works

* Loads project `go.mod` and stores it in memory
* For each direct dependency, it performs `go get -u DEPENDENCY@latest`
* If the command fails, it reverts to the last version of `go.mod`
* If the command upgraded Go version in `go.mod`, it reverts to the last version of `go.mod`
* If the command succeeds, it parses the newly created `go.mod` and continues working on other dependencies

## Uncovered use cases

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

## Planned features

* Major version bumps (find if there is a `/v2`).
* Run `go test` between updates.
* Store whole history and show diffs.
