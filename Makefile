.PHONY: build test clean

go.sum:
	go mod tidy

build: go.sum
	go build *.go

test: go.sum
	go test

clean:
	rm ikabot3
