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
$ go install github.com/lzap/gobump@latest
```

## Usage

```
$ cd ~/your_project

$ gobump
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

The above command upgraded all dependencies except `golang.org/x/tools` which would have increase Go requirement.

## How it works

* Loads project `go.mod` and stores it in memory
* For each direct dependency, it performs `go get -u DEPENDENCY@latest`
* If the command fails, it reverts to the last version of `go.mod`
* If the command upgraded Go version in `go.mod`, it reverts to the last version of `go.mod`
* If the command succeeds, it parses the newly created `go.mod` and continues working on other dependencies

## Configuration

It is possible to use different binary than `go`, set `GOVERSION=go1.21.0` environment variable to use a different Go version that is available through `PATH`.

## Planned features

* Major version bumps (find if there is a `/v2`).
* Run `go test` between updates.
* Store whole history and show diffs.
