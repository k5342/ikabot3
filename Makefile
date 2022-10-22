.PHONY: build test clean

.DEFAULT_GOAL := build

build: go.sum
	go build -o ikabot3

test: go.sum
	go test

clean:
	rm ikabot3
