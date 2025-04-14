.PHONY: tidy build

all: tidy build

tidy:
	go mod tidy

build:
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -linkmode external -extldflags '-static'" -o ./bin/server ./container-server
	env GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/incusTest ./test-runner
