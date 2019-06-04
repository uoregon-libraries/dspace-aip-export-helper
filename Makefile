.PHONY: build format vet lint

exporter: vet lint build

build:
	go build -ldflags="-s -w"

format:
	gofmt -l -s -w .

vet:
	go vet

lint:
	golint
