.PHONY: build test clean install lint

build:
	go build -o bin/nanami ./cmd

clean:
	rm -rf bin/
