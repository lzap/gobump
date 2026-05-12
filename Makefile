.PHONY: tidy
tidy:
	go mod tidy

.PHONY: install
install:
	go install ./...
