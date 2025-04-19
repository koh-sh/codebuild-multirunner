.PHONY: test fmt cov tidy run lint dockerbuild dockerrun blackboxtest modernize modernize-fix

COVFILE = coverage.out
COVHTML = cover.html

test:
	go test ./... -json | tparse -all

fmt:
	go tool gofumpt -l -w .

cov:
	go test -cover ./... -coverprofile=$(COVFILE)
	go tool cover -html=$(COVFILE) -o $(COVHTML)
	rm $(COVFILE)

tidy:
	go mod tidy -v

lint:
	go tool golangci-lint run -v

ci: fmt modernize-fix lint test

# Go Modernize
modernize:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -test ./...

modernize-fix:
	go run golang.org/x/tools/gopls/internal/analysis/modernize/cmd/modernize@latest -fix ./...

# for testing
dockerbuild:
	docker build . -t codebuild-multirunner:latest

dockerrun:
	docker run -it --rm -v ~/.aws:/root/.aws -v ~/.codebuild-multirunner.yaml:/.codebuild-multirunner.yaml codebuild-multirunner:latest -v

blackboxtest:
	./_testscripts/blackbox.sh
