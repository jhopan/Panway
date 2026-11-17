package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/jhopan/Panway/common/render"

	"github.com/gin-gonic/gin"
	"github.com/jhopan/Panway/common"
	"github.com/jhopan/Panway/common/conv"
	"github.com/jhopan/Panway/common/logger"
	"github.com/jhopan/Panway/relay/model"
	"github.com/jhopan/Panway/relay/relaymode"
)

const (
	dataPrefix       = "data: "
	done             = "[DONE]"
	dataPrefixLength = len(dataPrefix)
)

func StreamHandler(c *gin.Context, resp *http.Response, relayMode int, promptTokens int, modelName string) (*model.ErrorWithStatusCode, string, *model.Usage) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	var usage *model.Usage

	common.SetEventStreamHeaders(c)

	doneRendered := false
	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < dataPrefixLength { // ignore blank line or wrong format
			continue
		}
		if data[:dataPrefixLength] != dataPrefix && data[:dataPrefixLength] != done {
			continue
		}
		if strings.HasPrefix(data[dataPrefixLength:], done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}
		switch relayMode {
		case relaymode.ChatCompletions:
			var streamResponse ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				render.StringData(c, data) // if error happened, pass the data to client
				continue                   // just ignore the error
			}
			if len(streamResponse.Choices) == 0 && streamResponse.Usage == nil {
				// but for empty choice and no usage, we should not pass it to client, this is for azure
				continue // just ignore empty choice
			}
			for _, choice := range streamResponse.Choices {
				responseText += conv.AsString(choice.Delta.Content)
			}
			if streamResponse.Usage != nil {
				estimated := streamResponse.Usage.Estimated
				if estimated {
					// Replace the upstream estimate before it reaches the
					// client, so the usage the customer sees matches the
					// usage Panway bills. The response text is already
					// complete by the time the final usage chunk arrives.
					normalized := NormalizeEstimatedUsage(streamResponse.Usage, responseText, modelName, promptTokens)
					if normalized != streamResponse.Usage {
						streamResponse.Usage = normalized
						data = rewriteStreamUsage(data, normalized)
					}
				}
				// Prefer measured usage over an estimate. An upstream may send
				// an estimate first and the measured value later, or keep
				// sending estimates after one exact record.
				if usage == nil || usage.Estimated || !estimated {
					usage = streamResponse.Usage
				}
			}
			render.StringData(c, data)
		case relaymode.Completions:
			render.StringData(c, data)
			var streamResponse CompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}
			for _, choice := range streamResponse.Choices {
				responseText += choice.Text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	if !doneRendered {
		render.Done(c)
	}

	err := resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	return nil, responseText, usage
}

// rewriteStreamUsage replaces the usage object of one SSE chunk while keeping
// every other field (choice deltas, provider extras) intact.
func rewriteStreamUsage(data string, usage *model.Usage) string {
	payload := strings.TrimPrefix(data, dataPrefix)
	var chunk map[string]any
	if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
		return data
	}
	usageMap, ok := chunk["usage"].(map[string]any)
	if !ok {
		usageMap = map[string]any{}
	}
	usageMap["prompt_tokens"] = usage.PromptTokens
	usageMap["completion_tokens"] = usage.CompletionTokens
	usageMap["total_tokens"] = usage.TotalTokens
	delete(usageMap, "estimated")
	chunk["usage"] = usageMap
	rewritten, err := json.Marshal(chunk)
	if err != nil {
		return data
	}
	return dataPrefix + string(rewritten)
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var textResponse SlimTextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = decodeJSONBody(responseBody, &textResponse)
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}
	// Reset response body
	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the HTTPClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)
	// Strip a trailing stream terminator ("data: [DONE]") that some upstreams
	// append to non-streaming responses. Forwarding it breaks every
	// OpenAI-compatible client, which parses the body as one JSON document.
	forwarded := StripStreamTerminator(responseBody)
	_, err = c.Writer.Write(forwarded)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	if textResponse.Usage.TotalTokens == 0 || (textResponse.Usage.PromptTokens == 0 && textResponse.Usage.CompletionTokens == 0) {
		completionTokens := 0
		for _, choice := range textResponse.Choices {
			completionTokens += CountTokenText(choice.Message.StringContent(), modelName)
		}
		textResponse.Usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	if textResponse.Usage.Estimated {
		var responseText strings.Builder
		for _, choice := range textResponse.Choices {
			responseText.WriteString(choice.Message.StringContent())
		}
		textResponse.Usage = *NormalizeEstimatedUsage(&textResponse.Usage, responseText.String(), modelName, promptTokens)
	} else {
		learnPromptOverhead(modelName, promptTokens, textResponse.Usage.PromptTokens)
	}
	return nil, &textResponse.Usage
}
