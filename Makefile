.PHONY: install tidy tag

install:
	go install ./...

tidy:
	go mod tidy

# Create git tag v1.<X> (see scripts/next_v1_tag.sh).
tag:
	@bash scripts/next_v1_tag.sh
