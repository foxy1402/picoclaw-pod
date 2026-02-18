package agent

import (
	"strings"
	"testing"
)

func TestBuildMultimodalPrompt_WithCaption(t *testing.T) {
	msg := "what is the serial number?\n[image: photo]"
	got := buildMultimodalPrompt(msg)
	if !strings.Contains(got, "answer the user's request") {
		t.Fatalf("expected request-driven prompt, got: %s", got)
	}
	if !strings.Contains(got, "what is the serial number?") {
		t.Fatalf("expected caption in prompt, got: %s", got)
	}
}

func TestBuildMultimodalPrompt_WithoutCaption(t *testing.T) {
	msg := "[image: photo]"
	got := buildMultimodalPrompt(msg)
	if got != defaultMediaPrompt {
		t.Fatalf("expected default prompt, got: %s", got)
	}
}

func TestExtractUserMediaQuestion_RemovesSyntheticLines(t *testing.T) {
	msg := "please summarize\n[image: photo]\n[voice transcription: hello world]\n[file]"
	got := extractUserMediaQuestion(msg)
	if got != "please summarize" {
		t.Fatalf("unexpected extracted question: %q", got)
	}
}
