package barebitcoin

import (
	"context"
	"errors"
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

func TestAPIError(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantCode    int
		wantMessage string
		wantError   string
	}{
		{
			name:        "gRPC status body",
			status:      http.StatusBadRequest,
			body:        `{"code":3, "message":"cannot fetch by \"x\"", "details":[]}`,
			wantCode:    3,
			wantMessage: `cannot fetch by "x"`,
			wantError:   `HTTP 400: cannot fetch by "x"`,
		},
		{
			name:      "body does not override status code",
			status:    http.StatusInternalServerError,
			body:      `{"statusCode":200}`,
			wantError: `HTTP 500: {"statusCode":200}`,
		},
		{
			name:      "non-JSON body",
			status:    http.StatusBadGateway,
			body:      "bad gateway",
			wantError: "HTTP 502: bad gateway",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := &HTTPClient{baseURL: server.URL, client: server.Client()}

			err := client.doGetRequest(context.Background(), "/v1/orders", nil)

			apiErr, ok := errors.AsType[*APIError](err)
			if !ok {
				t.Fatalf("expected *APIError, got %T: %v", err, err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, apiErr.StatusCode)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("expected code %d, got %d", tt.wantCode, apiErr.Code)
			}
			if apiErr.Message != tt.wantMessage {
				t.Errorf("expected message %q, got %q", tt.wantMessage, apiErr.Message)
			}
			if string(apiErr.Body) != tt.body {
				t.Errorf("expected body %q, got %q", tt.body, apiErr.Body)
			}
			if got := err.Error(); got != tt.wantError {
				t.Errorf("expected error %q, got %q", tt.wantError, got)
			}
		})
	}
}
