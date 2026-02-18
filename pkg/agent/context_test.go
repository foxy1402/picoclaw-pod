package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipeed/picoclaw/pkg/providers"
)

func TestBuildMediaContext(t *testing.T) {
	got := buildMediaContext([]string{
		"/tmp/image.jpg",
		"/tmp/voice.ogg",
		"/tmp/doc.pdf",
	})

	want := []string{
		"Attached media files:",
		"[image] /tmp/image.jpg",
		"[audio] /tmp/voice.ogg",
		"[file] /tmp/doc.pdf",
	}

	for _, s := range want {
		if !strings.Contains(got, s) {
			t.Fatalf("expected media context to contain %q, got: %s", s, got)
		}
	}
}

func TestBuildMessages_IncludesMediaInUserMessage(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(tmpDir, "skills"), 0755); err != nil {
		t.Fatalf("failed to create skills dir: %v", err)
	}
	cb := NewContextBuilder(tmpDir)
	msgs := cb.BuildMessages(
		[]providers.Message{},
		"",
		"analyze this",
		[]string{"/tmp/photo.png"},
		"telegram",
		"123",
	)

	if len(msgs) == 0 {
		t.Fatal("expected messages")
	}
	last := msgs[len(msgs)-1]
	if last.Role != "user" {
		t.Fatalf("expected last message role=user, got %s", last.Role)
	}
	if !strings.Contains(last.Content, "analyze this") {
		t.Fatalf("expected user text in content, got: %s", last.Content)
	}
	if !strings.Contains(last.Content, "[image] /tmp/photo.png") {
		t.Fatalf("expected media path in content, got: %s", last.Content)
	}
}
