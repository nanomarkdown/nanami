.PHONY: build test clean dev docker lint fmt deps

BINARY_NAME=nanami
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

build:
	go build -o bin/nanami ./cmd

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage: test
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
