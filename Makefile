.PHONY: all build

all: build

build:
	go build -o ./bin/barebitcoin ./cmd/barebitcoin
