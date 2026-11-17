// Package connect implements `panway connect`, the one-command setup helper
// for OpenAI-compatible clients.
package connect

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const usageText = `Usage: panway connect --key sk-... [--base-url http://host:3000/v1] [--client all]

Clients: all, env, curl, python, node, cursor, continue, aider

The command prints copy-paste setup for the selected client and can write the
environment exports to a file with --env-file.
`

type options struct {
	baseURL string
	key     string
	client  string
	envFile string
}

func Run(args []string, out io.Writer) int {
	fs := flag.NewFlagSet("connect", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", envOr("PANWAY_BASE_URL", "http://localhost:3000/v1"), "OpenAI-compatible base URL")
	key := fs.String("key", envOr("PANWAY_API_KEY", os.Getenv("OPENAI_API_KEY")), "API key")
	client := fs.String("client", "all", "client to configure")
	envFile := fs.String("env-file", "", "write the environment exports to this file")
	if err := fs.Parse(args); err != nil {
		fmt.Fprint(out, usageText)
		return 2
	}
	opts := options{
		baseURL: normalizeBaseURL(*baseURL),
		key:     strings.TrimSpace(*key),
		client:  strings.ToLower(strings.TrimSpace(*client)),
		envFile: strings.TrimSpace(*envFile),
	}
	if opts.key == "" {
		fmt.Fprintln(out, "error: an API key is required (--key, PANWAY_API_KEY or OPENAI_API_KEY)")
		fmt.Fprint(out, usageText)
		return 2
	}
	if !validClient(opts.client) {
		fmt.Fprintf(out, "error: unknown client %q\n%s", opts.client, usageText)
		return 2
	}

	if opts.envFile != "" {
		if err := writeEnvFile(opts); err != nil {
			fmt.Fprintf(out, "error: cannot write env file: %v\n", err)
			return 1
		}
		fmt.Fprintf(out, "Environment written to %s\n\n", opts.envFile)
	}

	fmt.Fprintf(out, "Panway base URL: %s\n", opts.baseURL)
	switch opts.client {
	case "env":
		printEnv(out, opts)
	case "curl":
		printCurl(out, opts)
	case "python":
		printPython(out, opts)
	case "node":
		printNode(out, opts)
	case "cursor":
		printCursor(out, opts)
	case "continue":
		printContinue(out, opts)
	case "aider":
		printAider(out, opts)
	case "all":
		printEnv(out, opts)
		printCurl(out, opts)
		printPython(out, opts)
		printNode(out, opts)
		printCursor(out, opts)
		printContinue(out, opts)
		printAider(out, opts)
	}
	return 0
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func normalizeBaseURL(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimSuffix(value, "/")
	if !strings.HasSuffix(value, "/v1") {
		value += "/v1"
	}
	return value
}

func validClient(client string) bool {
	switch client {
	case "env", "curl", "python", "node", "cursor", "continue", "aider", "all":
		return true
	}
	return false
}

func writeEnvFile(opts options) error {
	content := fmt.Sprintf("export OPENAI_BASE_URL=%q\nexport OPENAI_API_KEY=%q\n", opts.baseURL, opts.key)
	return os.WriteFile(opts.envFile, []byte(content), 0o600)
}

func printEnv(out io.Writer, opts options) {
	fmt.Fprintf(out, "export OPENAI_BASE_URL=%q\nexport OPENAI_API_KEY=%q\n\n", opts.baseURL, opts.key)
}

func printCurl(out io.Writer, opts options) {
	fmt.Fprintf(out, `curl %s/chat/completions \
  -H "Authorization: Bearer %s" \
  -H "Content-Type: application/json" \
  -d '{"model":"MODEL_ID","messages":[{"role":"user","content":"Hello"}]}'

`, opts.baseURL, opts.key)
}

func printPython(out io.Writer, opts options) {
	fmt.Fprintf(out, `from openai import OpenAI
client = OpenAI(base_url=%q, api_key=%q)
print(client.chat.completions.create(
    model="MODEL_ID",
    messages=[{"role": "user", "content": "Hello"}],
).choices[0].message.content)

`, opts.baseURL, opts.key)
}

func printNode(out io.Writer, opts options) {
	fmt.Fprintf(out, `import OpenAI from "openai";
const client = new OpenAI({ baseURL: %q, apiKey: %q });
const reply = await client.chat.completions.create({
  model: "MODEL_ID",
  messages: [{ role: "user", content: "Hello" }],
});
console.log(reply.choices[0].message.content);

`, opts.baseURL, opts.key)
}

func printCursor(out io.Writer, opts options) {
	fmt.Fprintf(out, `Cursor:
  Settings -> Models -> OpenAI API Key: %s
  Override OpenAI Base URL: %s
  Add the model id from the Panway catalogue, then verify in Cursor.

`, opts.key, opts.baseURL)
}

func printContinue(out io.Writer, opts options) {
	fmt.Fprintf(out, `Continue (~/.continue/config.yaml):
  models:
    - name: panway
      provider: openai
      model: MODEL_ID
      apiBase: %s
      apiKey: %s

`, opts.baseURL, opts.key)
}

func printAider(out io.Writer, opts options) {
	fmt.Fprintf(out, `aider --openai-api-base %s --openai-api-key %s --model MODEL_ID

`, opts.baseURL, opts.key)
}
