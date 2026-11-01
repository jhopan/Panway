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
