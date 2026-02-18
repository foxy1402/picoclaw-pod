package agent

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	autoMultimodalHeader    = "Auto multimodal analysis (Gemini):"
	defaultMediaPrompt      = "Describe the attached media clearly. If it is an image, describe the scene and extract visible text. If it is audio/voice, transcribe and summarize key points."
	defaultMultimodalModel  = "gemini-2.5-flash-lite-latest"
	multimodalRequestTimout = 90 * time.Second
)

func (al *AgentLoop) enrichMessageWithMultimodal(ctx context.Context, userMessage string, media []string) string {
	if len(media) == 0 {
		return userMessage
	}
	apiKey := strings.TrimSpace(os.Getenv("PICOCLAW_PROVIDERS_GEMINI_API_KEY"))
	if apiKey == "" {
		return userMessage
	}

	prompt := buildMultimodalPrompt(userMessage)
	model := strings.TrimSpace(os.Getenv("PICOCLAW_MULTIMODAL"))
	if model == "" {
		model = defaultMultimodalModel
	}

	result, err := callGeminiMultimodal(ctx, apiKey, model, prompt, media)
	if err != nil {
		logger.WarnCF("agent", "Auto multimodal routing failed", map[string]interface{}{
			"error": err.Error(),
		})
		return userMessage
	}
	if strings.TrimSpace(result) == "" {
		return userMessage
	}

	base := strings.TrimSpace(userMessage)
	if base == "" {
		base = "[media message]"
	}
	return fmt.Sprintf("%s\n\n%s\n%s", base, autoMultimodalHeader, strings.TrimSpace(result))
}

func buildMultimodalPrompt(userMessage string) string {
	question := extractUserMediaQuestion(userMessage)
	if question == "" {
		return defaultMediaPrompt
	}
	return "Analyze the attached media and answer the user's request:\n" + question
}

func extractUserMediaQuestion(userMessage string) string {
	lines := strings.Split(userMessage, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" {
			continue
		}
		if isSyntheticMediaLine(trim) {
			continue
		}
		out = append(out, trim)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func isSyntheticMediaLine(line string) bool {
	lower := strings.ToLower(strings.TrimSpace(line))
	return strings.HasPrefix(lower, "[image:") ||
		lower == "[audio]" ||
		lower == "[file]" ||
		lower == "[voice]" ||
		strings.HasPrefix(lower, "[voice transcription:") ||
		lower == "[voice (transcription failed)]"
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
	Contents []geminiContent `json:"contents"`
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

func callGeminiMultimodal(ctx context.Context, apiKey, model, prompt string, media []string) (string, error) {
	parts := []geminiPart{{Text: prompt}}
	addedMedia := 0
	for _, mediaPath := range media {
		p := strings.TrimSpace(mediaPath)
		if p == "" {
			continue
		}
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MIMEType: detectMediaMIME(p),
				Data:     base64.StdEncoding.EncodeToString(data),
			},
		})
		addedMedia++
	}
	if addedMedia == 0 {
		return "", fmt.Errorf("no readable media files")
	}

	payload := geminiRequest{
		Contents: []geminiContent{{
			Role:  "user",
			Parts: parts,
		}},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal multimodal payload: %w", err)
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(model),
		url.QueryEscape(apiKey),
	)

	reqCtx, cancel := context.WithTimeout(ctx, multimodalRequestTimout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("create multimodal request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return "", fmt.Errorf("send multimodal request: %w", err)
	}
	defer resp.Body.Close()

	var parsed geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode multimodal response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	var out []string
	for _, c := range parsed.Candidates {
		for _, p := range c.Content.Parts {
			text := strings.TrimSpace(p.Text)
			if text != "" {
				out = append(out, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n\n")), nil
}

func detectMediaMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return "application/octet-stream"
	}
	if t := mime.TypeByExtension(ext); t != "" {
		return t
	}
	return "application/octet-stream"
}
