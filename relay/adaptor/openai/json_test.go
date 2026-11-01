package openai

import (
	"encoding/json"
	"testing"
)

func TestDecodeJSONBodyToleratesDoneMarker(t *testing.T) {
	body := []byte(`{"id":"x","choices":[],"usage":{"prompt_tokens":1}}data: [DONE]`)
	var out map[string]any
	if err := DecodeJSONBody(body, &out); err != nil {
		t.Fatalf("decode gagal: %v", err)
	}
	if out["id"] != "x" {
		t.Fatalf("isi salah: %+v", out)
	}
}

func TestStripStreamTerminator(t *testing.T) {
	body := []byte(`{"id":"x","choices":[],"usage":{"prompt_tokens":1}}data: [DONE]`)
	out := StripStreamTerminator(body)
	var probe map[string]any
	if err := json.Unmarshal(out, &probe); err != nil {
		t.Fatalf("hasil bukan JSON valid: %v (%s)", err, out)
	}
	if probe["id"] != "x" {
		t.Fatalf("isi salah: %+v", probe)
	}
}

func TestStripStreamTerminatorLeavesCleanBody(t *testing.T) {
	body := []byte(`{"id":"x"}`)
	out := StripStreamTerminator(body)
	if string(out) != string(body) {
		t.Fatalf("body bersih harus utuh, got %s", out)
	}
}

func TestStripStreamTerminatorLeavesStreamBody(t *testing.T) {
	// Body SSE asli tidak boleh dirusak: isinya bukan satu dokumen JSON.
	body := []byte("data: {\"choices\":[{\"delta\":{\"content\":\"hi\"}}]}\n\ndata: [DONE]\n\n")
	out := StripStreamTerminator(body)
	if string(out) != string(body) {
		t.Fatalf("body stream harus utuh, got %q", out)
	}
}
