.PHONY: install tidy release

install:
	go install ./...

tidy:
	go mod tidy

# Tag v1.<X+1>, push all tags, warm proxy.golang.org for this module (see scripts/release.sh).
release:
	@bash scripts/release.sh
