package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// TelegramClient sends messages through the Telegram Bot API.
//
// Only what Panway needs: sending a verification code and checking that the
// chat id belongs to the bot (getMe + getChat).
type TelegramClient struct {
	BotToken string
	BaseURL  string
	hc       *http.Client
}

// NewTelegramClient creates a client. baseURL may be empty; the official API
// endpoint is used then.
func NewTelegramClient(botToken, baseURL string) *TelegramClient {
	if baseURL == "" {
		baseURL = "https://api.telegram.org"
	}
	return &TelegramClient{
		BotToken: botToken,
		BaseURL:  strings.TrimSuffix(baseURL, "/"),
		hc:       &http.Client{Timeout: 15 * time.Second},
	}
}

// Enabled reports whether the client can send messages.
func (t *TelegramClient) Enabled() bool {
	return t != nil && strings.TrimSpace(t.BotToken) != ""
}

// telegramResponse is the common envelope of the Bot API.
type telegramResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		ID       int64  `json:"id"`
		Username string `json:"username"`
		FirstName string `json:"first_name"`
	} `json:"result"`
}

// SendText sends a plain text message to a chat.
func (t *TelegramClient) SendText(chatID string, text string) error {
	if !t.Enabled() {
		return fmt.Errorf("telegram bot token belum diisi")
	}
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}
	return t.call("/sendMessage", payload, nil)
}

// GetChat looks up a chat by id. It is used to verify that the chat id the
// customer typed actually belongs to a user of the bot.
func (t *TelegramClient) GetChat(chatID string) (username string, firstName string, err error) {
	if !t.Enabled() {
		return "", "", fmt.Errorf("telegram bot token belum diisi")
	}
	var out telegramResponse
	if err := t.call("/getChat", map[string]any{"chat_id": chatID}, &out); err != nil {
		return "", "", err
	}
	return out.Result.Username, out.Result.FirstName, nil
}

// GetMe returns the bot username, used to build the "@bot" hint shown to the
// customer.
func (t *TelegramClient) GetMe() (username string, err error) {
	if !t.Enabled() {
		return "", fmt.Errorf("telegram bot token belum diisi")
	}
	var out telegramResponse
	if err := t.call("/getMe", map[string]any{}, &out); err != nil {
		return "", err
	}
	return out.Result.Username, nil
}

func (t *TelegramClient) call(method string, payload any, out *telegramResponse) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := t.BaseURL + "/bot" + t.BotToken + method
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram %s: HTTP %d: %s", method, resp.StatusCode, string(raw))
	}
	var envelope telegramResponse
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return fmt.Errorf("telegram %s: bad response: %s", method, string(raw))
	}
	if !envelope.OK {
		return fmt.Errorf("telegram %s: %s", method, envelope.Description)
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return err
		}
	}
	return nil
}

// ParseChatID extracts a numeric chat id from user input. Users often paste
// "@username" instead of the number; that form is resolved by Telegram itself
// when sending, so it is accepted here as well.
func ParseChatID(input string) string {
	value := strings.TrimSpace(input)
	value = strings.TrimPrefix(value, "@")
	if value == "" {
		return ""
	}
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return value
	}
	return ""
}
