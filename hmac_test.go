package barebitcoin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestHMACExcludesQuery(t *testing.T) {
	var gotQuery, gotNonce, gotHMAC string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotNonce = r.Header.Get(headerNonce)
		gotHMAC = r.Header.Get(headerHMAC)
	}))
	defer srv.Close()

	client := NewHTTPClientWithKeys("key", "bb/apisecret/ZZavHDgVRyGowg8blKfPDDRlN3+6h0/vOUA")
	client.baseURL = srv.URL
	if _, err := client.GetBitcoinAccounts(context.Background(), true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotQuery != "includeDeleted=true" {
		t.Fatalf("query not sent, got %q", gotQuery)
	}

	nonce, err := strconv.ParseUint(gotNonce, 10, 64)
	if err != nil {
		t.Fatalf("invalid nonce: %v", err)
	}
	want, err := client.generateHMAC(http.MethodGet, "/v1/user/bitcoin-accounts", nonce, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHMAC != want {
		t.Errorf("HMAC mismatch\n got:  %s\nwant: %s", gotHMAC, want)
	}
}
