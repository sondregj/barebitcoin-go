# Bare Bitcoin client for Go

Bare Bitcoin is a norwegian service for exchanging and holding Bitcoin.
It provides an open API that this package implements a client library for.
Additionally, there is a CLI implemented in `cmd/barebitcoin` that wraps all the API operations.

## Guidelines

- Keep diffs minimal and to the point
- Don't make unnecessary changes without consent
- The client should be idiomatic to Go
  - Use Go style initialisms even when the API uses camel case

## Updating the client

`openapi.yaml` is the source of truth for the Bare Bitcoin API. Before updating
`client.go`, refresh the spec:

```sh
make openapi
```

Then update `client.go` to match any new or changed endpoints, types, and
fields. Keep `cmd/barebitcoin/main.go` in sync with any breaking signature
changes.
