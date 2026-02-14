# PicoClaw Pod (Cloud-Ready, Env-Only)

This repository is a cloud-focused PicoClaw variant optimized for container platforms (for example Claw Cloud) where runtime configuration is provided through **Environment Variables** and persistent workspace storage.

## What Is Different In This Repo

- Supports **env-only configuration**: `config.json` is optional.
- Designed for pod/container deployment with:
  - Command + Arguments startup model
  - Environment Variables for providers/channels
  - Local Storage mount for persistent workspace/session state
- Keeps PicoClaw core architecture (agent loop, tools, channels, skills, cron, heartbeat).

## Quick Start

### 1. Build Image

```bash
docker build -t picoclaw-pod:latest .
```

### 2. Run Locally (Env-Only)

```bash
docker run --rm -it \
  -e PICOCLAW_AGENTS_DEFAULTS_WORKSPACE=/root/.picoclaw/workspace \
  -e PICOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter \
  -e PICOCLAW_AGENTS_DEFAULTS_MODEL=openai/gpt-4o-mini \
  -e PICOCLAW_PROVIDERS_OPENROUTER_API_KEY=YOUR_KEY \
  -v picoclaw-workspace:/root/.picoclaw/workspace \
  picoclaw-pod:latest gateway
```

## Cloud Deployment (Claw Cloud Style)

Use the platform "Advanced config" fields as follows.

### Command / Arguments

- `Command`: `picoclaw`
- `Arguments`: `gateway`

### Environment Variables (Minimum)

- `PICOCLAW_AGENTS_DEFAULTS_WORKSPACE=/root/.picoclaw/workspace`
- `PICOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter`
- `PICOCLAW_AGENTS_DEFAULTS_MODEL=openai/gpt-4o-mini`
- `PICOCLAW_PROVIDERS_OPENROUTER_API_KEY=YOUR_KEY`

### Local Storage

Mount persistent volume to:

- `/root/.picoclaw/workspace`

This preserves sessions, memory, cron jobs, and state across restarts.

## Channel Enablement Examples

### Telegram

```env
PICOCLAW_CHANNELS_TELEGRAM_ENABLED=true
PICOCLAW_CHANNELS_TELEGRAM_TOKEN=YOUR_BOT_TOKEN
PICOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM=123456789
```

### Discord

```env
PICOCLAW_CHANNELS_DISCORD_ENABLED=true
PICOCLAW_CHANNELS_DISCORD_TOKEN=YOUR_DISCORD_TOKEN
PICOCLAW_CHANNELS_DISCORD_ALLOW_FROM=123456789012345678
```

## Provider Variables

Common examples:

```env
PICOCLAW_PROVIDERS_OPENROUTER_API_KEY=
PICOCLAW_PROVIDERS_OPENAI_API_KEY=
PICOCLAW_PROVIDERS_ANTHROPIC_API_KEY=
PICOCLAW_PROVIDERS_GROQ_API_KEY=
PICOCLAW_PROVIDERS_ZHIPU_API_KEY=
PICOCLAW_PROVIDERS_GEMINI_API_KEY=
```

`PICOCLAW_PROVIDERS_OPENAI_API_KEY` is not only for OpenAI official endpoint.
You can use it for any OpenAI-compatible provider by setting a custom base URL:

```env
PICOCLAW_AGENTS_DEFAULTS_PROVIDER=openai
PICOCLAW_AGENTS_DEFAULTS_MODEL=your-model-name
PICOCLAW_PROVIDERS_OPENAI_API_KEY=YOUR_COMPATIBLE_KEY
PICOCLAW_PROVIDERS_OPENAI_API_BASE=https://your-compatible-endpoint/v1
```

Optional per provider:

```env
PICOCLAW_PROVIDERS_<PROVIDER>_API_BASE=
PICOCLAW_PROVIDERS_<PROVIDER>_PROXY=
PICOCLAW_PROVIDERS_<PROVIDER>_AUTH_METHOD=
```

## Recommended Full Env Config (Cloud)

Use this as a production baseline, then enable channels/services you need.

```env
# Core agent
PICOCLAW_AGENTS_DEFAULTS_WORKSPACE=/root/.picoclaw/workspace
PICOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE=true
PICOCLAW_AGENTS_DEFAULTS_PROVIDER=openai
PICOCLAW_AGENTS_DEFAULTS_MODEL=gpt-4o-mini
PICOCLAW_AGENTS_DEFAULTS_MAX_TOKENS=8192
PICOCLAW_AGENTS_DEFAULTS_TEMPERATURE=0.7
PICOCLAW_AGENTS_DEFAULTS_MAX_TOOL_ITERATIONS=20

# OpenAI-compatible provider (generic)
PICOCLAW_PROVIDERS_OPENAI_API_KEY=YOUR_KEY
PICOCLAW_PROVIDERS_OPENAI_API_BASE=https://your-openai-compatible-endpoint/v1

# Optional: switch to OpenRouter instead
# PICOCLAW_AGENTS_DEFAULTS_PROVIDER=openrouter
# PICOCLAW_PROVIDERS_OPENROUTER_API_KEY=YOUR_KEY
# PICOCLAW_PROVIDERS_OPENROUTER_API_BASE=https://openrouter.ai/api/v1

# Web tools (important for smarter agent behavior)
PICOCLAW_TOOLS_WEB_BRAVE_ENABLED=true
PICOCLAW_TOOLS_WEB_BRAVE_API_KEY=YOUR_BRAVE_SEARCH_API_KEY
PICOCLAW_TOOLS_WEB_BRAVE_MAX_RESULTS=5
PICOCLAW_TOOLS_WEB_DUCKDUCKGO_ENABLED=true
PICOCLAW_TOOLS_WEB_DUCKDUCKGO_MAX_RESULTS=5

# Gateway
PICOCLAW_GATEWAY_HOST=0.0.0.0
PICOCLAW_GATEWAY_PORT=18790

# Heartbeat / periodic tasks
PICOCLAW_HEARTBEAT_ENABLED=true
PICOCLAW_HEARTBEAT_INTERVAL=30

# Devices (usually false on cloud pods)
PICOCLAW_DEVICES_ENABLED=false
PICOCLAW_DEVICES_MONITOR_USB=false
```

If you use allow lists, pass comma-separated values, for example:

```env
PICOCLAW_CHANNELS_TELEGRAM_ALLOW_FROM=12345,67890
```

## Agent CLI

```bash
# One-shot
picoclaw agent -m "What is 2+2?"

# Interactive
picoclaw agent

# Gateway mode
picoclaw gateway

# Status
picoclaw status
```

## Build From Source

```bash
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw
make build
```

Binary output:

- `build/picoclaw` (symlink)
- `build/picoclaw-<os>-<arch>`

## Workspace Layout

Default workspace path:

- `~/.picoclaw/workspace`

Key directories:

- `sessions/` conversation history
- `memory/` long-term memory
- `state/` last active channel/chat state
- `cron/` scheduled jobs
- `skills/` installed/custom skills

## Notes

- If no config file is present, the app now starts with defaults + env overrides.
- For cloud production, prefer env vars and secrets manager over file-based config.
- Hardware tools (I2C/SPI/USB monitor) require device access and are typically disabled on standard cloud pods.

## License

MIT
