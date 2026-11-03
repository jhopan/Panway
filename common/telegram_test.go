package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseChatID(t *testing.T) {
	cases := map[string]string{
		"123456789": "123456789",
		" 123456789 ": "123456789",
		"-1001234567890": "-1001234567890",
		"@username": "",
		"":           "",
		"abc":        "",
	}
	for input, want := range cases {
		if got := ParseChatID(input); got != want {
			t.Errorf("ParseChatID(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestTelegramClientDisabled(t *testing.T) {
	c := NewTelegramClient("", "")
	if c.Enabled() {
		t.Fatal("klien tanpa token tidak boleh enabled")
	}
	if err := c.SendText("1", "halo"); err == nil {
		t.Fatal("kirim tanpa token harus error")
	}
}

func TestTelegramSendText(t *testing.T) {
	var gotChatID string
	var gotText string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		json.NewDecoder(r.Body).Decode(&payload)
		gotChatID, _ = payload["chat_id"].(string)
		gotText, _ = payload["text"].(string)
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	c := NewTelegramClient("token", srv.URL)
	if err := c.SendText("42", "kode: 123456"); err != nil {
		t.Fatalf("kirim gagal: %v", err)
	}
	if gotChatID != "42" {
		t.Fatalf("chat id salah: %q", gotChatID)
	}
	if gotText != "kode: 123456" {
		t.Fatalf("teks salah: %q", gotText)
	}
}

func TestTelegramGetChat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ok": true,
			"result": map[string]any{
				"id": 42, "username": "pelanggan", "first_name": "Budi",
			},
		})
	}))
	defer srv.Close()

	c := NewTelegramClient("token", srv.URL)
	username, firstName, err := c.GetChat("42")
	if err != nil {
		t.Fatalf("get chat gagal: %v", err)
	}
	if username != "pelanggan" || firstName != "Budi" {
		t.Fatalf("data chat salah: %q %q", username, firstName)
	}
}

func TestTelegramErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"ok": false, "description": "chat not found",
		})
	}))
	defer srv.Close()

	c := NewTelegramClient("token", srv.URL)
	if _, _, err := c.GetChat("99"); err == nil {
		t.Fatal("respons ok=false harus jadi error")
	}
}
