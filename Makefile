.PHONY: all install tidy release

all: fmt tidy install

fmt:
	go fmt ./...

install:
	go install ./...

tidy:
	go mod tidy

bump:
	go run .

# Tag next v1.<minor>.0, push, warm proxy (see scripts/release.sh).
release:
	@bash scripts/release.sh
