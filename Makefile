.PHONY: setup test fmt cov tidy run lint

COVFILE = coverage.out
COVHTML = cover.html

setup:
	go install github.com/mfridman/tparse@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

test:
	go test ./... -json | tparse -all

fmt:
	gofumpt -l -w *.go

cov:
	go test -cover ./... -coverprofile=$(COVFILE)
	go tool cover -html=$(COVFILE) -o $(COVHTML)
	rm $(COVFILE)

tidy:
	go mod tidy -v

lint:
	golangci-lint run -v
