# barebitcoin-go

> ⚠️ **WARNING**: This package has not been sufficiently tested, which in the worst case may lead to **LOSS OF FUNDS**. Proceed at your own risk and preferably use read-only API keys.

A Go client library for the [Bare Bitcoin API](https://dev.barebitcoin.no), with an included CLI.

## Authentication

Create API keys at [barebitcoin.no](https://barebitcoin.no/innlogget/profil/nokler), then set these environment variables:

```sh
export BAREBITCOIN_PUBLIC_KEY="bb/public/..."
export BAREBITCOIN_SECRET_KEY="bb/apisecret/..."
```

## Library

```go
package main

import (
	barebitcoin "github.com/sondregj/barebitcoin-go"
)

func main() {
	client := barebitcoin.NewHTTPClientWithKeys(publicKey, secretKey)

	// Fetch current price
	price, err := client.GetPrice(ctx, 0)
	if err != nil {
		panic(err)
	}

	// Fetch account balances
	accounts, err := client.GetBitcoinAccounts(ctx, false)
	if err != nil {
		panic(err)
	}
}
```

## CLI

### Install

```sh
go install github.com/sondregj/barebitcoin-go/cmd/barebitcoin@latest
```

### Examples

```sh
# Check the current price
barebitcoin price

# View all balances
barebitcoin holdings

# Buy 500 NOK worth of bitcoin
barebitcoin buy 500

# Show deposit addresses
barebitcoin receive
```

## Related

- [API docs](https://dev.barebitcoin.no)
- [Bare Bitcoin](https://barebitcoin.no)
