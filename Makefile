all: build

VERSION=v1.6.0
GIT_SHA=$(shell git log --pretty=format:'%h' -n 1)
SHASUMCMD := $(shell command -v sha1sum || command -v shasum; 2> /dev/null)

.PHONY: build
build: 
	mkdir -p ./build/dist/darwin
	mkdir -p ./build/dist/linux
	GOOS=darwin GOARCH=amd64 go build -o ./build/dist/darwin/template-amd64 -ldflags "-X main.Version=${VERSION} -X blendlabs.com/template.GitVersion=${GIT_SHA}" cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o ./build/dist/linux/template-amd64  -ldflags "-X main.Version=${VERSION} -X blendlabs.com/template.GitVersion=${GIT_SHA}" cmd/main.go
	(${SHASUMCMD} ./build/dist/darwin/template-amd64 | cut -d' ' -f1) > ./build/dist/darwin/template-amd64.sha1
	(${SHASUMCMD} ./build/dist/linux/template-amd64 | cut -d' ' -f1) > ./build/dist/linux/template-amd64.sha1

.PHONY: release-tag
release-tag:
	@git tag ${VERSION}
	@git push --tags

.PHONY: test
test:
	@go test 
