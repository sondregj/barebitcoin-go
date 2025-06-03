package barebitcoin

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"
)

const (
	EnvPublicKey = "BAREBITCOIN_PUBLIC_KEY"
	EnvSecretKey = "BAREBITCOIN_SECRET_KEY"

	BaseURL = "https://api.bb.no"
)

type HTTPClient struct {
	apiKey    string
	secretKey string
	baseURL   string
	client    *http.Client
}

// NewHTTPClient creates a new HTTP client using environment variables for API keys.
func NewHTTPClient() *HTTPClient {
	if os.Getenv(EnvPublicKey) == "" || os.Getenv(EnvSecretKey) == "" {
		panic("BAREBITCOIN_PUBLIC_KEY and BAREBITCOIN_SECRET_KEY environment variables must be set")
	}
	return &HTTPClient{
		apiKey:    os.Getenv(EnvPublicKey),
		secretKey: os.Getenv(EnvSecretKey),
		baseURL:   BaseURL,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

// NewHTTPClientWithKeys creates a new HTTP client with the provided API keys.
func NewHTTPClientWithKeys(apiKey, secretKey string) *HTTPClient {
	return &HTTPClient{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   BaseURL,
		client:    &http.Client{Timeout: 30 * time.Second},
	}
}

type UserPrice struct {
	Buy  float64 `json:"buy"`
	Sell float64 `json:"sell"`
}

type Currency string

const (
	CurrencyUnspecified Currency = "CURRENCY_UNSPECIFIED"
	CurrencyNOK         Currency = "CURRENCY_NOK"
	CurrencyBTC         Currency = "CURRENCY_BTC"
)

type NewLightningInvoiceRequest struct {
	// If empty, the default account is used.
	AccountID string `json:"accountId,omitempty"`

	// The currency the invoice is denominated in. Once created, the invoice is
	// for a specific amount of bitcoin.
	Currency Currency `json:"currency"`

	// The amount of the invoice, in the requested currency.
	Amount float64 `json:"amount"`

	// Public description of the invoice. Shown to both the creator and the recipient of the invoice.
	// This is the so-called "memo" field of the Lightning invoice.
	PublicDescription string `json:"publicDescription,omitempty"`

	// Free-form text description of the invoice. Can be used to correlate with your own systems.
	// Only shown to the creator of this invoice, within the Bare Bitcoin systems.
	InternalDescription string `json:"internalDescription,omitempty"`
}

type NewLightningInvoiceResponse struct {
	DepositDestinationID string `json:"depositDestinationId"`
	Invoice              string `json:"invoice"`
}

type LightningInvoiceStatus string

const (
	LightningInvoiceStatusUnspecified LightningInvoiceStatus = "INVOICE_STATUS_UNSPECIFIED"
	LightningInvoiceStatusUnpaid      LightningInvoiceStatus = "INVOICE_STATUS_UNPAID"
	LightningInvoiceStatusPending     LightningInvoiceStatus = "INVOICE_STATUS_PENDING"
	LightningInvoiceStatusPaid        LightningInvoiceStatus = "INVOICE_STATUS_PAID"
	LightningInvoiceStatusExpired     LightningInvoiceStatus = "INVOICE_STATUS_EXPIRED"
	LightningInvoiceStatusCanceled    LightningInvoiceStatus = "INVOICE_STATUS_CANCELED"
)

type GetLightningInvoiceResponse struct {
	DepositDestinationID string                 `json:"depositDestinationId"`
	Invoice              string                 `json:"invoice"`
	Status               LightningInvoiceStatus `json:"status"`
}

type LedgerEntry struct {
	TransactionID string    `json:"transactionId"`
	Type          string    `json:"type"`
	Timestamp     time.Time `json:"timestamp"`
	Currency      string    `json:"currency"`
	Value         float64   `json:"value"`
	Fee           float64   `json:"fee"`
	Balance       float64   `json:"balance"`
}

type WithdrawalStatus string

const (
	WithdrawalStatusUnspecified WithdrawalStatus = "WITHDRAWAL_STATUS_UNSPECIFIED"
	WithdrawalStatusPending     WithdrawalStatus = "WITHDRAWAL_STATUS_PENDING"
	WithdrawalStatusCompleted   WithdrawalStatus = "WITHDRAWAL_STATUS_COMPLETED"
	WithdrawalStatusFailed      WithdrawalStatus = "WITHDRAWAL_STATUS_FAILED"
)

type GetBitcoinWithdrawalResponse struct {
	WithdrawalID string           `json:"withdrawalId"`
	Destination  string           `json:"destination"`
	Network      string           `json:"network"`
	AmountBTC    float64          `json:"amountBtc"`
	AmountNOK    float64          `json:"amountNok"`
	Status       WithdrawalStatus `json:"status"`
	CreatedAt    time.Time        `json:"createdAt"`
	SentAt       time.Time        `json:"sentAt"`
}

type PriceResponse struct {
	Price     float64   `json:"price"`
	Bid       float64   `json:"bid"`
	Ask       float64   `json:"ask"`
	Timestamp time.Time `json:"timestamp"`
	Card      UserPrice `json:"card"`
	Bank      UserPrice `json:"bank"`
}

func FetchBitcoinNOKPrice(ctx context.Context) (*PriceResponse, error) {
	client := NewHTTPClient()
	var response PriceResponse
	err := client.doGetRequest(ctx, "/v1/price/nok", &response)
	return &response, err
}

type DepositNetwork string

const (
	DepositNetworkUnspecified DepositNetwork = "NETWORK_UNSPECIFIED"
	DepositNetworkBitcoin     DepositNetwork = "NETWORK_BITCOIN"
	DepositNetworkLightning   DepositNetwork = "NETWORK_LIGHTNING"
)

type DepositDestination struct {
	Destination string         `json:"destination"`
	Network     DepositNetwork `json:"network"`
}

type DepositDestinationsResponse struct {
	OnchainAddress   *DepositDestination `json:"onchainAddress,omitempty"`
	LightningAddress *DepositDestination `json:"lightningAddress,omitempty"`
	LNURLPay         *DepositDestination `json:"lnurlPay,omitempty"`
}

func ListBitcoinDepositDestinations(ctx context.Context, accountID string) (*DepositDestinationsResponse, error) {
	client := NewHTTPClient()
	var response DepositDestinationsResponse
	path := "/v1/deposit-destinations/bitcoin"
	if accountID != "" {
		path += "?accountId=" + accountID
	}
	err := client.doGetRequest(ctx, path, &response)
	return &response, err
}

func (c *HTTPClient) CreateLightningInvoice(ctx context.Context, req *NewLightningInvoiceRequest) (*NewLightningInvoiceResponse, error) {
	var response NewLightningInvoiceResponse
	err := c.doPostRequest(ctx, "/v1/deposit-destinations/bitcoin/invoice", req, &response)
	return &response, err
}

func (c *HTTPClient) GetLightningInvoice(ctx context.Context, id string) (*GetLightningInvoiceResponse, error) {
	var response GetLightningInvoiceResponse
	err := c.doGetRequest(ctx, "/v1/deposit-destinations/bitcoin/invoice/"+id, &response)
	return &response, err
}

type GetLedgerResponse struct {
	Entries []LedgerEntry `json:"entries"`
}

func (c *HTTPClient) GetLedger(ctx context.Context) (*GetLedgerResponse, error) {
	var response GetLedgerResponse
	err := c.doGetRequest(ctx, "/v1/ledger", &response)
	return &response, err
}

type GetLedgerAccountResponse struct {
	Entries []LedgerEntry `json:"entries"`
}

func (c *HTTPClient) GetLedgerAccount(ctx context.Context, accountID string) (*GetLedgerAccountResponse, error) {
	var response GetLedgerAccountResponse
	err := c.doGetRequest(ctx, "/v1/ledger/"+accountID, &response)
	return &response, err
}

type Order struct {
	OrderID   string    `json:"orderId"`
	Type      string    `json:"type"`
	Direction string    `json:"direction"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"createdAt"`
}

type OpenOrdersResponse struct {
	Orders []Order `json:"orders"`
}

func (c *HTTPClient) GetOrders(ctx context.Context) (*OpenOrdersResponse, error) {
	var response OpenOrdersResponse
	err := c.doGetRequest(ctx, "/v1/orders", &response)
	return &response, err
}

type NewOrderRequest struct {
	Type        string  `json:"type"`
	Direction   string  `json:"direction"`
	Amount      float64 `json:"amount"`
	AccountID   string  `json:"accountId"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price,omitempty"`
}

type NewOrderResponse struct {
	OrderID string `json:"orderId"`
	TradeID string `json:"tradeId"`
}

func (c *HTTPClient) CreateOrder(ctx context.Context, req *NewOrderRequest) (*NewOrderResponse, error) {
	var response NewOrderResponse
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	err = c.doPostRequest(ctx, "/v1/orders", body, &response)
	return &response, err
}

func (c *HTTPClient) DeleteOrder(ctx context.Context, orderID string) error {
	return c.doDeleteRequest(ctx, "/v1/orders/"+orderID, nil)
}

// Public service

func (c *HTTPClient) GetPrice(ctx context.Context) (*PriceResponse, error) {
	var response PriceResponse
	path := "/v1/price/nok"
	// Does not require authentication
	err := c.doGetRequest(ctx, path, &response)
	return &response, err
}

type VolumeStats struct {
	AmountNOK     float64 `json:"amountNok"`
	AmountBTC     float64 `json:"amountBtc"`
	BuyPercentage float64 `json:"buyPercentage"`
}

type VolumeResponse struct {
	Daily   VolumeStats `json:"daily"`
	Monthly VolumeStats `json:"monthly"`
	Yearly  VolumeStats `json:"yearly"`
}

func (c *HTTPClient) GetVolume(ctx context.Context, currency string) (*VolumeResponse, error) {
	var response VolumeResponse
	path := "/v1/volume"
	if currency != "" {
		path += "?currency=" + currency
	}
	// Does not require authentication
	err := c.doGetRequest(ctx, path, &response)
	return &response, err
}

// User service

type ListBitcoinAccountsResponse struct {
	Accounts []BitcoinAccount `json:"accounts"`
}

type BitcoinAccount struct {
	ID               string  `json:"id"`
	AvailableBTC     float64 `json:"availableBtc"`
	PendingOrdersBTC float64 `json:"pendingOrdersBtc"`
}

func (c *HTTPClient) GetBitcoinAccounts(ctx context.Context) (*ListBitcoinAccountsResponse, error) {
	var response ListBitcoinAccountsResponse
	err := c.doGetRequest(ctx, "/v1/user/bitcoin-accounts", &response)
	return &response, err
}

type GetFiatAccountResponse struct {
	AvailableNOK     float64 `json:"availableNok"`
	PendingOrdersNOK float64 `json:"pendingOrdersNok"`
}

func (c *HTTPClient) GetFiatAccount(ctx context.Context) (*GetFiatAccountResponse, error) {
	var response GetFiatAccountResponse
	err := c.doGetRequest(ctx, "/v1/user/fiat-account", &response)
	return &response, err
}

type SendBitcoinRequest struct {
	AccountID   string  `json:"accountId"`
	Destination string  `json:"destination"`
	AmountBTC   float64 `json:"amountBtc"`
	Description string  `json:"description,omitempty"`
	IsPayment   bool    `json:"isPayment,omitempty"`
}

type SendBitcoinResponse struct {
	WithdrawalID string `json:"withdrawalId"`
	Network      string `json:"network"`
	Status       string `json:"status"`
}

func (c *HTTPClient) SendBitcoin(ctx context.Context, req *SendBitcoinRequest) (*SendBitcoinResponse, error) {
	var response SendBitcoinResponse
	err := c.doPostRequest(ctx, "/v1/withdrawals/bitcoin", req, &response)
	return &response, err
}

func (c *HTTPClient) GetBitcoinWithdrawal(ctx context.Context, withdrawalId string) (*GetBitcoinWithdrawalResponse, error) {
	var response GetBitcoinWithdrawalResponse
	err := c.doGetRequest(ctx, "/v1/withdrawals/bitcoin/"+withdrawalId, &response)
	return &response, err
}

const (
	headerAPIKey = "x-bb-api-key"
	headerNonce  = "x-bb-api-nonce"
	headerHMAC   = "x-bb-api-hmac"
)

func (c *HTTPClient) doGetRequest(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *HTTPClient) doPostRequest(ctx context.Context, path string, body, out any) error {
	return c.doRequest(ctx, http.MethodPost, path, body, out)
}

func (c *HTTPClient) doDeleteRequest(ctx context.Context, path string, out any) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, out)
}

// HMAC-SHA256 of (URI path + SHA256(nonce + request body)) and base64 decoded secret API key
func (c *HTTPClient) generateHMAC(method, path string, nonce uint64, body []byte) (string, error) {
	encodedData := fmt.Sprintf("%d%s", nonce, string(body))
	summed := sha256.Sum256([]byte(encodedData))
	message := slices.Concat([]byte(method), []byte(path), summed[:])

	decodedSecret, err := base64.StdEncoding.DecodeString(c.secretKey)
	if err != nil {
		return "", fmt.Errorf("invalid HMAC secret: %w", err)
	}

	mac := hmac.New(sha256.New, decodedSecret)
	mac.Write(message)
	macsum := mac.Sum(nil)

	digest := base64.StdEncoding.EncodeToString(macsum)
	return digest, nil
}

func (c *HTTPClient) doRequest(ctx context.Context, method, path string, body, out any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	bodyString := ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
		bodyString = string(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	// Add authentication headers if keys are available
	if c.apiKey != "" && c.secretKey != "" {
		nonce := uint64(time.Now().UnixNano()) / 1000000
		signature, err := c.generateHMAC(method, path, nonce, []byte(bodyString))
		if err != nil {
			return fmt.Errorf("generating HMAC: %w", err)
		}

		req.Header.Set(headerAPIKey, c.apiKey)
		req.Header.Set(headerNonce, strconv.FormatUint(nonce, 10))
		req.Header.Set(headerHMAC, signature)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}
