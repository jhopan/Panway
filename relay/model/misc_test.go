package model

import (
	"encoding/json"
	"testing"
)

func TestUsageParsesCachedTokens(t *testing.T) {
	raw := `{"prompt_tokens":3027,"completion_tokens":8,"total_tokens":3035,
		"prompt_tokens_details":{"cached_tokens":256}}`
	var usage Usage
	if err := json.Unmarshal([]byte(raw), &usage); err != nil {
		t.Fatalf("gagal parse usage: %v", err)
	}
	if usage.PromptTokensDetails == nil {
		t.Fatal("prompt_tokens_details tidak terbaca")
	}
	if usage.PromptTokensDetails.CachedTokens != 256 {
		t.Fatalf("cached tokens salah: %d", usage.PromptTokensDetails.CachedTokens)
	}
}

func TestUsageWithoutCachedDetails(t *testing.T) {
	var usage Usage
	if err := json.Unmarshal([]byte(`{"prompt_tokens":10,"completion_tokens":1}`), &usage); err != nil {
		t.Fatalf("gagal parse usage: %v", err)
	}
	if usage.PromptTokensDetails != nil {
		t.Fatalf("tanpa details harus nil: %+v", usage.PromptTokensDetails)
	}
}
