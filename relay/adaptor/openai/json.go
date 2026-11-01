package openai

import (
	"bytes"
	"encoding/json"
)

// DecodeJSONBody parses the upstream response body as JSON.
//
// Some upstreams (PanRouter among them) append a stream terminator such as
// "data: [DONE]" to a non-streaming response. A plain json.Unmarshal fails on
// the trailing text, so the decoder stops after the first complete JSON value
// instead.
//
// The returned body is the untouched original, so the response can still be
// forwarded verbatim to the client.
func DecodeJSONBody(body []byte, out any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	return decoder.Decode(out)
}

// decodeJSONBody is the internal alias used inside this package.
func decodeJSONBody(body []byte, out any) error {
	return DecodeJSONBody(body, out)
}

// StripStreamTerminator removes a trailing stream terminator from a
// non-streaming response body.
//
// PanRouter appends "data: [DONE]" to its non-streaming responses. Passing
// that through breaks every OpenAI-compatible client: the SDK tries to parse
// the whole body as one JSON document and fails with "Extra data".
//
// The body is only modified when the terminator is genuinely there and what
// remains is a complete JSON document, so a normal response is untouched.
func StripStreamTerminator(body []byte) []byte {
	trimmed := bytes.TrimRight(body, "\r\n\t ")
	if !bytes.HasSuffix(trimmed, []byte("[DONE]")) {
		return body
	}
	// Cut the terminator and everything before it on the same "data:" line.
	idx := bytes.LastIndex(trimmed, []byte("data:"))
	if idx < 0 {
		return body
	}
	head := bytes.TrimRight(trimmed[:idx], "\r\n\t ")
	if len(head) == 0 {
		return body
	}
	// What remains must still be a complete JSON document.
	var probe map[string]any
	if err := json.Unmarshal(head, &probe); err != nil {
		return body
	}
	return head
}
