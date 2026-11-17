package connect

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRunPrintsOpenAICompatibleSetup(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{"--key", "sk-test", "--base-url", "https://api.example.com", "--client", "python"}, &out)
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	text := out.String()
	if !strings.Contains(text, "https://api.example.com/v1") {
		t.Fatalf("base URL tidak dinormalisasi: %s", text)
	}
	if !strings.Contains(text, "base_url=\"https://api.example.com/v1\"") {
		t.Fatalf("snippet python salah: %s", text)
	}
	if !strings.Contains(text, "sk-test") {
		t.Fatalf("key tidak muncul: %s", text)
	}
}

func TestRunRequiresKey(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{"--client", "env"}, &out)
	if code == 0 {
		t.Fatal("tanpa key harus gagal")
	}
	if !strings.Contains(out.String(), "API key is required") {
		t.Fatalf("pesan error tidak jelas: %s", out.String())
	}
}

func TestRunWritesEnvFile(t *testing.T) {
	var out bytes.Buffer
	path := t.TempDir() + "/panway.env"
	code := Run([]string{"--key", "sk-env", "--base-url", "http://localhost:3000/v1", "--client", "env", "--env-file", path}, &out)
	if code != 0 {
		t.Fatalf("exit code %d", code)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("env file tidak dibuat: %v", err)
	}
	if !strings.Contains(string(data), "OPENAI_API_KEY") || !strings.Contains(string(data), "sk-env") {
		t.Fatalf("isi env file salah: %s", data)
	}
}

func TestNormalizeBaseURL(t *testing.T) {
	cases := map[string]string{
		"http://localhost:3000":       "http://localhost:3000/v1",
		"http://localhost:3000/":      "http://localhost:3000/v1",
		"https://api.example.com/v1/": "https://api.example.com/v1",
	}
	for input, want := range cases {
		if got := normalizeBaseURL(input); got != want {
			t.Fatalf("%q -> %q, ingin %q", input, got, want)
		}
	}
}
