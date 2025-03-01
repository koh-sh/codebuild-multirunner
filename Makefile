.PHONY: test fmt cov tidy run lint dockerbuild dockerrun blackboxtest

COVFILE = coverage.out
COVHTML = cover.html

test:
	go test ./... -json | tparse -all

fmt:
	go tool gofumpt -l -w *.go

cov:
	go test -cover ./... -coverprofile=$(COVFILE)
	go tool cover -html=$(COVFILE) -o $(COVHTML)
	rm $(COVFILE)

tidy:
	go mod tidy -v

lint:
	go tool golangci-lint run -v

# for testing
dockerbuild:
	docker build . -t codebuild-multirunner:latest

dockerrun:
	docker run -it --rm -v ~/.aws:/root/.aws -v ~/.codebuild-multirunner.yaml:/.codebuild-multirunner.yaml codebuild-multirunner:latest -v

blackboxtest:
	./_testscripts/blackbox.sh
