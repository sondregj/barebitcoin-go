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
	"time"

	"github.com/AlecAivazis/survey/v2"
)

func main() {
	ctx := context.Background()
	session, err := initSession(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// fmt.Println("session", session.AccessToken)

	client := &Client{
		accessToken: session.AccessToken,
	}
	user, err := client.GetUser(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("user balance sat", user.BalanceSat)

	bitcoinHoldings, err := client.GetBitcoinHoldings(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("current balance", bitcoinHoldings.Entries[len(bitcoinHoldings.Entries)-1].BalanceSatoshi, "sat")
}

type Client struct {
	accessToken string
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
		Value: c.accessToken,
	})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		fmt.Println("response", string(b))
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
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
