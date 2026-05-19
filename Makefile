# Match the `go` directive in go.mod (used to avoid accidental toolchain / module bumps).
GOTOOLCHAIN ?= go$(shell awk '/^go /{print $$2}' go.mod)

.PHONY: all fmt install tidy bump test release

all: fmt install

fmt:
	go fmt ./...

install:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go install ./...

test:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go test ./...

# Intentional dependency maintenance only — not part of `make all`.
tidy:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go mod tidy

# Bumps dependencies in the **current directory**; do not use on this repo when developing gobump.
bump:
	GOTOOLCHAIN=$(GOTOOLCHAIN) go run .

# Tag next v1.<minor>.0, push, warm proxy (see scripts/release.sh).
release:
	@bash scripts/release.sh
