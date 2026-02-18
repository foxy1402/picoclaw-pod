package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultModel              = "gemini-2.5-flash-lite-latest"
	mediaCleanupMaxAge        = time.Hour
	defaultMediaDir           = "/tmp/picoclaw_media"
	defaultRequestTimeout     = 120 * time.Second
	defaultGenerationTemp     = 0.2
	fallbackContentType       = "application/octet-stream"
	errCodeInvalidUsage int   = 2
	errCodeRequestFail  int   = 1
)

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type geminiInlineData struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inline_data,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents         []geminiContent       `json:"contents"`
	GenerationConfig map[string]float64    `json:"generationConfig,omitempty"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func main() {
	if code := run(); code != 0 {
		os.Exit(code)
	}
}

func run() int {
	var prompt string
	var modelOverride string
	var mediaArgs multiFlag

	flag.StringVar(&prompt, "prompt", "", "Instruction for Gemini.")
	flag.StringVar(&modelOverride, "model", "", "Override model (else use PICOCLAW_MULTIMODAL).")
	flag.Var(&mediaArgs, "media", "Media file path (repeatable).")
	flag.Parse()

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		fmt.Fprintln(os.Stderr, "ERROR: --prompt is required")
		return errCodeInvalidUsage
	}

	apiKey := strings.TrimSpace(os.Getenv("PICOCLAW_PROVIDERS_GEMINI_API_KEY"))
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "ERROR: missing PICOCLAW_PROVIDERS_GEMINI_API_KEY")
		return errCodeInvalidUsage
	}

	model := strings.TrimSpace(modelOverride)
	if model == "" {
		model = strings.TrimSpace(os.Getenv("PICOCLAW_MULTIMODAL"))
	}
	if model == "" {
		model = defaultModel
	}

	if len(mediaArgs) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: at least one --media path is required")
		return errCodeInvalidUsage
	}

	mediaPaths := make([]string, 0, len(mediaArgs))
	for _, raw := range mediaArgs {
		path := strings.TrimSpace(raw)
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			fmt.Fprintf(os.Stderr, "ERROR: media file not found: %s\n", path)
			return errCodeInvalidUsage
		}
		mediaPaths = append(mediaPaths, path)
	}
	if len(mediaPaths) == 0 {
		fmt.Fprintln(os.Stderr, "ERROR: at least one valid --media path is required")
		return errCodeInvalidUsage
	}

	cleanupOldMedia(defaultMediaDir, mediaCleanupMaxAge)

	respText, rawResp, err := callGemini(apiKey, model, prompt, mediaPaths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Gemini API request failed: %v\n", err)
		if strings.TrimSpace(rawResp) != "" {
			fmt.Fprintln(os.Stderr, rawResp)
		}
		return errCodeRequestFail
	}

	if strings.TrimSpace(respText) == "" {
		fmt.Fprintln(os.Stderr, "ERROR: Gemini returned no text output")
		if strings.TrimSpace(rawResp) != "" {
			fmt.Fprintln(os.Stderr, rawResp)
		}
		return errCodeRequestFail
	}

	fmt.Println(respText)
	return 0
}

func callGemini(apiKey, model, prompt string, mediaPaths []string) (string, string, error) {
	parts := []geminiPart{{Text: prompt}}
	for _, path := range mediaPaths {
		part, err := toInlinePart(path)
		if err != nil {
			return "", "", err
		}
		parts = append(parts, part)
	}

	reqPayload := geminiRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: parts,
		}},
		GenerationConfig: map[string]float64{"temperature": defaultGenerationTemp},
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		return "", "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(model),
		url.QueryEscape(apiKey),
	)

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: defaultRequestTimeout}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return "", "", err
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return "", "", fmt.Errorf("read response: %w", err)
	}

	rawResp := string(respBody)
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return "", rawResp, fmt.Errorf("Gemini API HTTP %d", httpResp.StatusCode)
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", rawResp, fmt.Errorf("parse response: %w", err)
	}

	text := extractText(parsed)
	return text, rawResp, nil
}

func toInlinePart(path string) (geminiPart, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return geminiPart{}, fmt.Errorf("read media file %q: %w", path, err)
	}

	mimeType := detectMIME(path)
	encoded := base64.StdEncoding.EncodeToString(data)

	return geminiPart{
		InlineData: &geminiInlineData{
			MIMEType: mimeType,
			Data:     encoded,
		},
	}, nil
}

func detectMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return fallbackContentType
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return fallbackContentType
}

func extractText(resp geminiResponse) string {
	var out []string
	for _, c := range resp.Candidates {
		for _, part := range c.Content.Parts {
			if strings.TrimSpace(part.Text) != "" {
				out = append(out, part.Text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n\n"))
}

func cleanupOldMedia(mediaDir string, maxAge time.Duration) {
	entries, err := os.ReadDir(mediaDir)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(mediaDir, entry.Name()))
		}
	}
}
