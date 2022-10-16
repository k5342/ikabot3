.PHONY: build test clean

build:
	go build *.go

test:
	go test

clean:
	rm ikabot3
