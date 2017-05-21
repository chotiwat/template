all: build

TEMPLATE_RELEASE_VERSION=v1.0
TEMPLATE_CI_VERSION=v1.1-beta
GIT_SHA=$(shell git log --pretty=format:'%h' -n 1)

ifndef VERSION
  ifndef CI
    VERSION=${TEMPLATE_RELEASE_VERSION}
  else
    VERSION := ${TEMPLATE_CI_VERSION}+${GIT_SHA}
  endif
endif

SHASUMCMD := $(shell command -v sha1sum || command -v shasum; 2> /dev/null)

.PHONY: build
build: 
	mkdir -p ./build/dist
	GOOS=darwin GOARCH=amd64 go build -o ./build/dist/template-darwin-amd64 -ldflags "-X blendlabs.com/template.Version=${VERSION} -X blendlabs.com/template.GitVersion=${GIT_SHA}" main.go
	GOOS=linux GOARCH=amd64 go build -o ./build/dist/template-linux-amd64  -ldflags "-X blendlabs.com/template.Version=${VERSION} -X blendlabs.com/template.GitVersion=${GIT_SHA}" main.go
	(${SHASUMCMD} ./build/dist/template-darwin-amd64 | cut -d' ' -f1) > ./build/dist/template-darwin-amd64.sha1
	(${SHASUMCMD} ./build/dist/template-linux-amd64 | cut -d' ' -f1) > ./build/dist/template-linux-amd64.sha1


.PHONY: release-tag
release-tag:
	@git tag ${TEMPLATE_RELEASE_VERSION}
	@git push --tags

.PHONY: release-deps
release-deps:
	@go get -u github.com/kopeio/shipbot/cmd/shipbot

.PHONY: test
test:
	@go test ./template/...