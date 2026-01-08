.PHONY: lint vet test coverage

VERSION?=latest
GOARCH?=arm64
GOOS?=linux

build:
	GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=$(VERSION)" -o dist/
package: build
	nfpm pkg --packager deb --target dist/
lint:
	go fmt $(go list ./... | grep -v /vendor/)
vet:
	go vet $(go list ./... | grep -v /vendor/)
clean:
	rm -f dist/*
