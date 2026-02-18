---
name: gemini_vision_audio
description: Analyze images and audio files with Gemini multimodal API when media attachments are present or when the main model lacks vision/audio support. Use this skill to call the bundled script via exec, then use the script output as grounded context for the main LLM loop.
metadata: {"nanobot":{"emoji":"🎧","requires":{"bins":["go"]}}}
---

# Gemini Vision Audio

Use this skill when the user message includes media files (especially from Telegram) and your main model cannot process multimodal inputs directly.

## Environment

Require:

```bash
PICOCLAW_PROVIDERS_GEMINI_API_KEY=YOUR_GEMINI_KEY
```

Optional model override:

```bash
PICOCLAW_MULTIMODAL=gemini-2.5-flash-lite-latest
```

Default model if unset: `gemini-2.5-flash-lite-latest`.

## Tool-Calling Workflow

1. Detect media paths from the current message context (`Attached media files:` section).
2. Build a focused prompt for what to extract (description, OCR, summary, speaker intent, etc.).
3. Call `exec` with the script:

```bash
go run {baseDir}/scripts/gemini_multimodal.go \
  --prompt "Describe the image and extract visible text." \
  --media "/tmp/picoclaw_media/abc_photo.jpg"
```

Multiple files:

```bash
go run {baseDir}/scripts/gemini_multimodal.go \
  --prompt "Summarize each file, then provide one combined conclusion." \
  --media "/tmp/picoclaw_media/a.jpg" \
  --media "/tmp/picoclaw_media/b.ogg"
```

4. Use the returned text in your final answer or next tool step.

## Guardrails

- Do not switch the main provider/model. Keep main loop settings unchanged.
- Use Gemini only for multimodal interpretation tasks.
- If no media path is available, ask for media or continue with text-only handling.
- If API key is missing, return a clear setup error and continue gracefully.
