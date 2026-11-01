// Package walletclient talking to PanWallet, the service that owns customer
// money.
//
// Panway never touches the money itself. The flow is:
//
//  1. Customer pays a QRIS invoice in PanWallet.
//  2. PanWallet's webhook fires; it carries the wallet's public_id and the
//     amount in IDR.
//  3. Panway converts IDR into quota units and credits the matching user.
//
// PanWallet is the source of truth for the balance. Panway's quota column is
// a mirror used for rate limiting and pricing only.
package walletclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a PanWallet API client.
type Client struct {
	baseURL string
	apiKey  string
	hc      *http.Client
}

// New creates a client. baseURL points at the PanWallet API root, apiKey is
// the product key for this product (e.g. "panway").
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		hc:      &http.Client{Timeout: 15 * time.Second},
	}
}

// Balance is the wallet snapshot returned by PanWallet.
type Balance struct {
	PublicID  string `json:"public_id"`
	Balance   int64  `json:"balance"`
	Held      int64  `json:"held"`
	Available int64  `json:"available"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
}

// GetBalance reads a wallet by its short public id (8 characters).
func (c *Client) GetBalance(publicID string) (Balance, error) {
	var b Balance
	err := c.do("GET", "/v1/public/"+publicID, "", nil, &b)
	return b, err
}

// Hold locks an amount for one order. Amount is in IDR.
func (c *Client) Hold(walletUserID, referenceID string, amount int64, idem string) (map[string]any, error) {
	body := map[string]any{
		"wallet_user_id": walletUserID,
		"reference_id":   referenceID,
		"amount":         amount,
		"expires_at":     time.Now().Add(30 * time.Minute).UTC().Format(time.RFC3339),
	}
	var out map[string]any
	err := c.do("POST", "/v1/holds", idem, body, &out)
	return out, err
}

// Capture settles a hold. Without partial capture support in PanWallet, the
// full hold amount is charged.
func (c *Client) Capture(holdID, idem string, amount int64) (map[string]any, error) {
	var out map[string]any
	err := c.do("POST", "/v1/holds/"+holdID+"/capture", idem, nil, &out)
	return out, err
}

// Release cancels a hold without charging.
func (c *Client) Release(holdID, idem string) (map[string]any, error) {
	var out map[string]any
	err := c.do("POST", "/v1/holds/"+holdID+"/release", idem, nil, &out)
	return out, err
}

func (c *Client) do(method, path, idem string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.baseURL+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	if idem != "" {
		req.Header.Set("Idempotency-Key", idem)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("panwallet %s %s -> %d: %s", method, path, resp.StatusCode, string(raw))
	}
	if out != nil && len(raw) > 0 {
		return json.Unmarshal(raw, out)
	}
	return nil
}
