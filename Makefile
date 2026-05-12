.PHONY: install tidy tag

install:
	go install ./...

tidy:
	go mod tidy

# Create git tag v1.<X> (see scripts/next_v1_tag.py).
tag:
	@python3 scripts/next_v1_tag.py
