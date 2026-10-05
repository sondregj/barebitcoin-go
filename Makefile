.PHONY: all build openapi

all: build

build:
	go build -o ./bin/barebitcoin ./cmd/barebitcoin

openapi:
	curl -fL "https://barebitcoin.no/developers/openapi.yaml" -o ./openapi.yaml
	prettier --write ./openapi.yaml
