.PHONY: all build openapi

all: build

build:
	go build -o ./bin/barebitcoin ./cmd/barebitcoin

openapi:
	curl "https://dev.barebitcoin.no/_spec/api/openapi.yaml?download" -o ./openapi.yaml
	prettier --write ./openapi.yaml
