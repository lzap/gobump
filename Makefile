.PHONY: install
install:
	go install ./...

.PHONY: tidy
tidy:
	go mod tidy
