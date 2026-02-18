package agent

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sipeed/picoclaw/pkg/logger"
)

const (
	autoMultimodalHeader = "Auto multimodal analysis (Gemini):"
	defaultMediaPrompt   = "Describe the attached media clearly. If it is an image, describe the scene and extract visible text. If it is audio/voice, transcribe and summarize key points."
)

func (al *AgentLoop) enrichMessageWithMultimodal(ctx context.Context, userMessage string, media []string) string {
	if len(media) == 0 {
		return userMessage
	}
	if strings.TrimSpace(os.Getenv("PICOCLAW_PROVIDERS_GEMINI_API_KEY")) == "" {
		return userMessage
	}

	scriptPath := al.resolveGeminiMultimodalScriptPath()
	if scriptPath == "" {
		logger.WarnCF("agent", "Skipping auto multimodal routing: Gemini script not found", nil)
		return userMessage
	}

	prompt := buildMultimodalPrompt(userMessage)
	result, err := runGeminiMultimodalScript(ctx, scriptPath, prompt, media)
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

func (al *AgentLoop) resolveGeminiMultimodalScriptPath() string {
	candidates := []string{
		filepath.Join(al.workspace, "skills", "gemini_vision_audio", "scripts", "gemini_multimodal.go"),
		filepath.Join("skills", "gemini_vision_audio", "scripts", "gemini_multimodal.go"),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			return p
		}
	}
	return ""
}

func runGeminiMultimodalScript(ctx context.Context, scriptPath, prompt string, media []string) (string, error) {
	args := []string{"run", scriptPath, "--prompt", prompt}
	for _, m := range media {
		path := strings.TrimSpace(m)
		if path == "" {
			continue
		}
		args = append(args, "--media", path)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Env = os.Environ()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("go multimodal helper failed: %s", errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
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
