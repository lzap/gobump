.PHONY: install tidy release

install:
	go install ./...

tidy:
	go mod tidy

# Tag next v1.<minor>.0, push tags, warm proxy (see scripts/release.sh).
release:
	@bash scripts/release.sh
