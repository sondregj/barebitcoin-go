package barebitcoin

import (
	"testing"
)

func TestGenerateHMAC(t *testing.T) {
	// Test vector from https://dev.barebitcoin.no/#hmac-details
	secret := "bb/apisecret/ZZavHDgVRyGowg8blKfPDDRlN3+6h0/vOUA"
	nonce := 1733314678
	body := `{"type": "ORDER_TYPE_MARKET", "direction": "DIRECTION_BUY", "amount": 100}`
	uri := "/v1/orders"
	method := "POST"
	hmac := "e6pQ5w9AqVhwHRWXuwS7ZwzRd0kH2GYpHSmtP0cTlSU="

	client := &HTTPClient{
		secretKey: secret,
	}

	got, err := client.generateHMAC(method, uri, uint64(nonce), []byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := hmac
	if got != want {
		t.Errorf("HMAC mismatch\n got:  %s\nwant: %s", got, want)
	}
}
