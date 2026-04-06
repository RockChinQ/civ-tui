.PHONY: build run test lint clean

build:
	go build ./...

run:
	go run main.go

test:
	go test ./game/...

lint:
	go vet ./...

clean:
	go clean ./...
