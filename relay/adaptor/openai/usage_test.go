package openai

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/jhopan/Panway/common/config"
	"github.com/jhopan/Panway/relay/model"
	"github.com/jhopan/Panway/relay/relaymode"
)

func useApproximateTokens(t *testing.T) {
	t.Helper()
	previous := config.ApproximateTokenEnabled
	config.ApproximateTokenEnabled = true
	t.Cleanup(func() { config.ApproximateTokenEnabled = previous })
}

func resetPromptOverheadForTest(t *testing.T) {
	t.Helper()
	promptOverhead.Lock()
	promptOverhead.byModel = map[string]int{}
	promptOverhead.Unlock()
}

func TestNormalizeEstimatedUsageUsesLocalCounts(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)

	upstream := &model.Usage{PromptTokens: 2500, CompletionTokens: 2, TotalTokens: 2502, Estimated: true}
	got := NormalizeEstimatedUsage(upstream, "hello world", "agnes/agnes-3.0-flash", 20)

	if got.Estimated {
		t.Fatal("usage hasil normalisasi tidak boleh ditandai estimated")
	}
	if got.PromptTokens != 20 {
		t.Fatalf("prompt harus dari hitungan lokal, dapat %d", got.PromptTokens)
	}
	wantCompletion := CountTokenText("hello world", "agnes/agnes-3.0-flash")
	if got.CompletionTokens != wantCompletion {
		t.Fatalf("completion harus dari teks stream, dapat %d ingin %d", got.CompletionTokens, wantCompletion)
	}
	if got.TotalTokens != got.PromptTokens+got.CompletionTokens {
		t.Fatalf("total tidak konsisten: %d", got.TotalTokens)
	}
	if !upstream.Estimated || upstream.PromptTokens != 2500 {
		t.Fatal("usage masukan tidak boleh diubah")
	}
}

func TestNormalizeEstimatedUsageKeepsExactUsage(t *testing.T) {
	upstream := &model.Usage{PromptTokens: 431, CompletionTokens: 2, TotalTokens: 433}
	got := NormalizeEstimatedUsage(upstream, "text", "m", 100)

	if got != upstream {
		t.Fatal("usage exact harus dikembalikan apa adanya")
	}
}

func TestNormalizeEstimatedUsageKeepsCompletionWithoutText(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)

	upstream := &model.Usage{PromptTokens: 2500, CompletionTokens: 7, TotalTokens: 2507, Estimated: true}
	got := NormalizeEstimatedUsage(upstream, "", "m", 12)

	if got.PromptTokens != 12 {
		t.Fatalf("prompt harus lokal, dapat %d", got.PromptTokens)
	}
	if got.CompletionTokens != 7 {
		t.Fatalf("completion tanpa teks harus memakai angka upstream, dapat %d", got.CompletionTokens)
	}
}

func TestNormalizeEstimatedUsageUsesLearnedPromptOverhead(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)

	learnPromptOverhead("agnes/agnes-3.0-flash", 10, 499)
	upstream := &model.Usage{PromptTokens: 2500, CompletionTokens: 2, TotalTokens: 2502, Estimated: true}
	got := NormalizeEstimatedUsage(upstream, "hi", "agnes/agnes-3.0-flash", 10)

	if got.PromptTokens != 499 {
		t.Fatalf("prompt harus ikut overhead upstream yang sudah dipelajari, dapat %d", got.PromptTokens)
	}
}

func TestNormalizeEstimatedUsageUpstreamModePassesThrough(t *testing.T) {
	previous := config.EstimatedUsageMode
	config.EstimatedUsageMode = "upstream"
	t.Cleanup(func() { config.EstimatedUsageMode = previous })

	upstream := &model.Usage{PromptTokens: 2500, CompletionTokens: 2, TotalTokens: 2502, Estimated: true}
	got := NormalizeEstimatedUsage(upstream, "hi", "m", 10)

	if got != upstream {
		t.Fatal("mode upstream harus memakai angka upstream apa adanya")
	}
}

func streamContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

func streamResponse(chunks ...string) *http.Response {
	return &http.Response{Body: io.NopCloser(strings.NewReader(strings.Join(chunks, "\n\n")))}
}

func TestStreamHandlerPrefersExactUsageAfterEstimate(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)
	c, _ := streamContext()
	resp := streamResponse(
		`data: {"choices":[{"delta":{"content":"hello world there"}}]}`,
		`data: {"choices":[],"usage":{"prompt_tokens":2500,"completion_tokens":1,"total_tokens":2501,"estimated":true}}`,
		`data: {"choices":[],"usage":{"prompt_tokens":431,"completion_tokens":2,"total_tokens":433}}`,
		`data: [DONE]`,
	)

	err, text, usage := StreamHandler(c, resp, relaymode.ChatCompletions, 10, "test-model")
	if err != nil {
		t.Fatalf("stream error: %v", err)
	}
	if text != "hello world there" {
		t.Fatalf("teks stream salah: %q", text)
	}
	if usage == nil || usage.Estimated {
		t.Fatal("usage exact harus dipilih")
	}
	if usage.PromptTokens != 431 || usage.CompletionTokens != 2 {
		t.Fatalf("usage salah: %+v", usage)
	}
}

func TestStreamHandlerDoesNotLetEstimateOverwriteExactUsage(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)
	c, _ := streamContext()
	resp := streamResponse(
		`data: {"choices":[],"usage":{"prompt_tokens":431,"completion_tokens":2,"total_tokens":433}}`,
		`data: {"choices":[],"usage":{"prompt_tokens":4513,"completion_tokens":2,"total_tokens":4515,"estimated":true}}`,
		`data: [DONE]`,
	)

	_, _, usage := StreamHandler(c, resp, relaymode.ChatCompletions, 10, "test-model")
	if usage == nil || usage.Estimated || usage.PromptTokens != 431 {
		t.Fatalf("estimasi tidak boleh menimpa usage exact: %+v", usage)
	}
}

func TestStreamHandlerNormalizesEstimatedUsageBeforeForwarding(t *testing.T) {
	useApproximateTokens(t)
	resetPromptOverheadForTest(t)
	c, w := streamContext()
	resp := streamResponse(
		`data: {"choices":[{"delta":{"content":"hello world there"}}]}`,
		`data: {"choices":[],"usage":{"prompt_tokens":2500,"completion_tokens":2,"total_tokens":2502,"estimated":true,"neurons":1.5}}`,
		`data: [DONE]`,
	)

	_, _, usage := StreamHandler(c, resp, relaymode.ChatCompletions, 10, "test-model")
	if usage == nil || usage.Estimated {
		t.Fatalf("estimasi harus dinormalisasi: %+v", usage)
	}
	if usage.PromptTokens != 10 {
		t.Fatalf("prompt harus lokal, dapat %d", usage.PromptTokens)
	}

	forwarded := w.Body.String()
	if strings.Contains(forwarded, `"estimated":true`) {
		t.Fatalf("chunk yang dikirim ke klien masih estimated: %s", forwarded)
	}
	if !strings.Contains(forwarded, `"prompt_tokens":10`) {
		t.Fatalf("chunk yang dikirim ke klien harus memakai prompt hasil normalisasi: %s", forwarded)
	}
	if !strings.Contains(forwarded, `"neurons":1.5`) {
		t.Fatalf("field lain di chunk harus dipertahankan: %s", forwarded)
	}
}
