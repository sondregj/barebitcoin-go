package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		fmt.Println("Usage: barebitcoin <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	if command == "logout" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		sessionPath := filepath.Join(home, ".config", "barebitcoin", "session.json")
		if err := os.Remove(sessionPath); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("logged out")
		os.Exit(0)
	}

	session, err := initSession(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	client := &Client{
		session: session,
	}
	// NOTE: only doing this until I can do it properly
	{
		accessToken, err := RefreshCookie(ctx, client.session.AccessToken, client.session.RefreshToken)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		client.session.AccessToken = *accessToken
	}

	switch command {
	case "user":
		user, err := client.GetUser(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("user balance sat", user.BalanceSat)

	case "holdings":
		bitcoinHoldings, err := client.GetBitcoinHoldings(ctx)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println("current balance", bitcoinHoldings.Entries[len(bitcoinHoldings.Entries)-1].BalanceSatoshi, "sat")

	case "stats":
		stats, err := client.GetStats(ctx)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("stats {")
		fmt.Println("  purchase total sat", stats.PurchaseTotalSats)
		fmt.Println("  purchase spent nok", stats.PurchaseSpentNOK)
		fmt.Println("  purchase current total value", stats.PurchaseCurrentTotalValue)
		fmt.Println("  purchase average rate", stats.PurchaseAverageRate)
		fmt.Println("  purchase min rate", stats.PurchaseMinRate)
		fmt.Println("  purchase max rate", stats.PurchaseMaxRate)
		fmt.Println("}")

	case "history":
		historicalPrices, err := client.HistoricalBitcoinPrices(ctx)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("historical prices {")
		fmt.Println(" day count", len(historicalPrices.Day))
		fmt.Println(" week count", len(historicalPrices.Week))
		fmt.Println(" month count", len(historicalPrices.Month))
		fmt.Println(" quarter count", len(historicalPrices.Quarter))
		fmt.Println(" half year count", len(historicalPrices.HalfYear))
		fmt.Println(" year count", len(historicalPrices.Year))
		fmt.Println(" current year count", len(historicalPrices.CurrentYear))
		fmt.Println(" start count", len(historicalPrices.Start))
		fmt.Println("}")

	case "invoice":
		var amountSatoshi int
		if len(os.Args) == 3 {
			amountSatoshi, err = strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		invoice, err := client.NewLightningInvoice(ctx, amountSatoshi)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(invoice)

	default:
		fmt.Printf("unknown command %q\n", command)
	}

	if err := saveSession(client.session); err != nil {
		fmt.Println(err)
	}
}

type Client struct {
	session *LoginResponse
}

type User struct {
	ID                     string           `json:"id"`
	PersonalInfo           UserPersonalInfo `json:"personalInfo"`
	PrepaidNOK             string           `json:"prepaidNok"`
	BalanceNOK             string           `json:"balanceNok"`
	BalanceSat             string           `json:"balanceSat"`
	BalanceSatWithdrawable string           `json:"balanceSatWithdrawable"`
	FeeRatePercent         float64          `json:"feeRatePercent"`
	FeeRate                struct {
		Buy  float64 `json:"buy"`
		Sell float64 `json:"sell"`
		Card float64 `json:"card"`
	} `json:"feeRate"`
	VerifiedIdentity bool   `json:"verifiedIdentity"`
	VerifiedPhone    bool   `json:"verifiedPhone"`
	SetPassword      bool   `json:"setPassword"`
	Denomination     string `json:"denomination"`
	RateFiatCurrency string `json:"rateFiatCurrency"`
	WithdrawalSpeed  string `json:"withdrawalSpeed"`
	ConfirmedEmail   bool   `json:"confirmedEmail"`
	ReferralCode     string `json:"referralCode"`
	ReferredBy       string `json:"referredBy"`
	CreateTime       string `json:"createTime"`
}

type UserPersonalInfo struct {
	Name             string `json:"name"`
	Phone            string `json:"phone"`
	Email            string `json:"email"`
	NationalIDNumber string `json:"nationalIdNumber"`
	Address          string `json:"address"`
	Zip              string `json:"zip"`
	Age              int    `json:"age"`
	BirthDate        struct {
		Year  int `json:"year"`
		Month int `json:"month"`
		Day   int `json:"day"`
	} `json:"birthDate"`
	Citizenship struct {
		Country string `json:"country"`
		Code    string `json:"code"`
	} `json:"citizenship"`
}

func (c *Client) GetUser(ctx context.Context) (*User, error) {
	var response struct {
		User User `json:"user"`
	}
	err := c.post(ctx, "https://barebitcoin.no/connect/bb.v1alpha.UserService/GetUser", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response.User, nil
}

type Stats struct {
	PurchaseAverageRate       string `json:"purchaseAverageRate"`
	PurchaseMaxRate           string `json:"purchaseMaxRate"`
	PurchaseMinRate           string `json:"purchaseMinRate"`
	PurchaseTotalSats         string `json:"purchaseTotalSats"`
	PurchaseSpentNOK          string `json:"purchaseSpentNok"`
	PurchaseCurrentTotalValue string `json:"purchaseCurrentTotalValue"`
}

func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	var response Stats
	err := c.post(ctx, "https://barebitcoin.no/connect/bb.v1alpha.UserService/GetStats", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type BalanceEntry struct {
	BalanceNOKValue float64   `json:"balanceNokValue"`
	BalanceSatoshi  string    `json:"balanceSatoshi"`
	BTCNOK          float64   `json:"btcnok"`
	PNLAbsolute     float64   `json:"pnlAbsolute,omitempty"`
	PNLPercent      float64   `json:"pnlPercent,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}
type AllTimeData struct {
	Entries []BalanceEntry `json:"allTime"`
}

func (c *Client) GetBitcoinHoldings(ctx context.Context) (*AllTimeData, error) {
	var response AllTimeData
	err := c.post(ctx, "https://barebitcoin.no/connect/bb.pnl.v1.PnlService/BitcoinHoldings", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type HistoricalPrices struct {
	Day         []HistoricalPrice `json:"day"`
	Week        []HistoricalPrice `json:"week"`
	Month       []HistoricalPrice `json:"month"`
	Quarter     []HistoricalPrice `json:"quarter"`
	HalfYear    []HistoricalPrice `json:"halfYear"`
	Year        []HistoricalPrice `json:"year"`
	CurrentYear []HistoricalPrice `json:"currentYear"`
	Start       []HistoricalPrice `json:"start"`
}
type HistoricalPrice struct {
	BTCNOK    float64              `json:"btcnok"`
	Timestamp string               `json:"timestamp"`
	Delta     HistoricalPriceDelta `json:"delta"`
}
type HistoricalPriceDelta struct {
	NOK     float64 `json:"nok"`
	Percent float64 `json:"percent"`
}

func (c *Client) HistoricalBitcoinPrices(ctx context.Context) (*HistoricalPrices, error) {
	var response HistoricalPrices
	err := c.post(ctx, "https://barebitcoin.no/connect/bb.v1alpha.BBService/HistoricalBitcoinPrices", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TODO: https://barebitcoin.no/connect/bb.deposits.v1.DepositsService/GetDestinations

func (c *Client) NewLightningInvoice(ctx context.Context, amountSatoshi int) (string, error) {
	var response struct {
		PaymentRequest string `json:"paymentRequest"`
	}
	body := map[string]string{}
	if amountSatoshi > 0 {
		body["amountSatoshi"] = strconv.Itoa(amountSatoshi)
	}
	err := c.post(ctx, "https://barebitcoin.no/connect/bb.deposits.v1.DepositsService/NewLightningInvoice", body, &response)
	if err != nil {
		return "", err
	}
	return response.PaymentRequest, nil
}

func (c *Client) post(ctx context.Context, path string, body, response any) error {
	if body == nil {
		body = map[string]any{}
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "bb_access_token",
		Value: c.session.AccessToken,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return fmt.Errorf("unexpected status code %d", resp.StatusCode)
		}
		if apiErr.Message == "expired access token" {
			// TODO: refresh token
		}
		fmt.Println(resp)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, apiErr.Message)
	}
	return json.NewDecoder(resp.Body).Decode(response)
}

func initSession(ctx context.Context) (*LoginResponse, error) {
	session, err := getSession()
	if err == nil {
		return session, nil
	}
	if !errors.Is(err, ErrNoSession) {
		return nil, err
	}

	email := os.Getenv("BAREBITCOIN_EMAIL")
	password := os.Getenv("BAREBITCOIN_PASSWORD")
	if email == "" || password == "" {
		return nil, errors.New("BAREBITCOIN_EMAIL and BAREBITCOIN_PASSWORD must be set")
	}
	login, err := Login(ctx, email, password)
	if err != nil {
		return nil, err
	}
	if login.RequiresMFA {
		var mfaCode string
		survey.AskOne(&survey.Input{Message: "Enter MFA code"}, &mfaCode)
		err = ValidateGenerator(ctx, login.AccessToken, mfaCode)
		if err != nil {
			return nil, err
		}
	}
	if !login.VerifiedEmail {
		return nil, errors.New("email not verified")
	}

	err = saveSession(login)
	if err != nil {
		return nil, err
	}
	session = login

	return session, nil
}

var ErrNoSession = errors.New("no session")

func getSession() (*LoginResponse, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filepath.Join(home, ".config", "barebitcoin", "session.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNoSession
	} else if err != nil {
		return nil, err
	}
	var session LoginResponse
	err = json.Unmarshal(b, &session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func saveSession(session *LoginResponse) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".config", "barebitcoin")
	if err := os.MkdirAll(configDir, os.ModePerm); err != nil {
		return err
	}
	sessionFile := filepath.Join(configDir, "session.json")
	b, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return os.WriteFile(sessionFile, b, 0600)
}

// bb_access_token=ota_...
type LoginResponse struct {
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	VerifiedEmail bool   `json:"verifiedEmail"`
	RequiresMFA   bool   `json:"requiresMfa"`
}

func Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	var request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	request.Email = email
	request.Password = password
	b, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(b)
	resp, err := http.Post("https://barebitcoin.no/connect/bb.v1alpha.AuthService/Login", "application/json", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("login failed")
	}
	var response LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	fmt.Println("login header", resp.Header)

	return &response, nil
}

// FIXME: this does not really work, the token must be refreshed while the existing
// access token is still valid
func RefreshCookie(ctx context.Context, accessToken, refreshToken string) (*string, error) {
	url := "https://barebitcoin.no/connect/bb.cookie.v1.CookieService/RefreshCookie"
	body := bytes.NewReader([]byte("{\"refresfhToken\": \"" + refreshToken + "\"}"))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("content-type", "application/json")
	// req.AddCookie(&http.Cookie{
	// 	Name:  "bb_refresh_token",
	// 	Value: refreshToken,
	// })
	req.AddCookie(&http.Cookie{
		Name:  "bb_access_token",
		Value: accessToken,
	})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		fmt.Println(string(b))
		return nil, errors.New("refresh cookie bad status: " + resp.Status)
	}
	var newAccessToken string
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "bb_access_token" {
			newAccessToken = cookie.Value
		}
	}
	return &newAccessToken, nil
}

func ValidateGenerator(ctx context.Context, accessToken, code string) error {
	var request struct {
		Code string `json:"code"`
	}
	request.Code = code
	b, err := json.Marshal(request)
	if err != nil {
		return err
	}
	body := bytes.NewReader(b)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://barebitcoin.no/connect/bb.auth.v1.AuthenticatorService/ValidateGenerator", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "bb_access_token",
		Value: accessToken,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println("validate generator", string(b))
	if resp.StatusCode != http.StatusOK {
		return errors.New("validate generator failed")
	}
	return nil
}

// {"code":"unauthenticated","message":"no authentication found"}

//  https://barebitcoin.no/connect/bb.v1alpha.BBService/Price
// // {
//     "buyBtcnok": 636279.26,
//     "spreadBuy": 0.25,
//     "sellBtcnok": 633061.95,
//     "spreadSell": 0.25,
//     "midBtcnok": 634670.605,
//     "usdnok": "10.95",
//     "timestamp": "2024-08-04T17:37:15.420413252Z"
// }

// https://barebitcoin.no/connect/bb.v1alpha.PurchaseService/ListAutoPurchases
// // {
//     "autoPurchases": [
//         {
//             "id": "uuid",
//             "userId": "uuid",
//             "frequency": "SAVINGS_FREQUENCY_DAILY",
//             "amount": "1",
//             "daysUntilNextPurchase": "1",
//             "createTime": "time",
//             "totalPurchaseVolumeSatoshi": "445594",
//             "name": "string",
//             "accountId": "acc_...",
//             "waivedFee": true
//         }
//     ],
//     "totalMonthlyPurchaseAmount": 0,
//     "prepaidEmptyInDays": 16,
//     "receivedFirstPayment": true
// }

// https://barebitcoin.no/connect/bb.auth.v1.AuthenticatorService/GetAuthenticators
// {
//     "app": {
//         "id": "ts_...",
//         "type": "AUTHENTICATOR_TYPE_APP",
//         "verified": true
//     }
// }
