package config

import (
	"path/filepath"
	"testing"
)

// TestDefaultConfig_HeartbeatEnabled verifies heartbeat is enabled by default
func TestDefaultConfig_HeartbeatEnabled(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Heartbeat.Enabled {
		t.Error("Heartbeat should be enabled by default")
	}
}

// TestDefaultConfig_WorkspacePath verifies workspace path is correctly set
func TestDefaultConfig_WorkspacePath(t *testing.T) {
	cfg := DefaultConfig()

	// Just verify the workspace is set, don't compare exact paths
	// since expandHome behavior may differ based on environment
	if cfg.Agents.Defaults.Workspace == "" {
		t.Error("Workspace should not be empty")
	}
}

// TestDefaultConfig_Model verifies model is set
func TestDefaultConfig_Model(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Agents.Defaults.Model == "" {
		t.Error("Model should not be empty")
	}
}

// TestDefaultConfig_MaxTokens verifies max tokens has default value
func TestDefaultConfig_MaxTokens(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Agents.Defaults.MaxTokens == 0 {
		t.Error("MaxTokens should not be zero")
	}
}

// TestDefaultConfig_MaxToolIterations verifies max tool iterations has default value
func TestDefaultConfig_MaxToolIterations(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Agents.Defaults.MaxToolIterations == 0 {
		t.Error("MaxToolIterations should not be zero")
	}
}

// TestDefaultConfig_Temperature verifies temperature has default value
func TestDefaultConfig_Temperature(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Agents.Defaults.Temperature == 0 {
		t.Error("Temperature should not be zero")
	}
}

// TestDefaultConfig_Gateway verifies gateway defaults
func TestDefaultConfig_Gateway(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Gateway.Host != "0.0.0.0" {
		t.Error("Gateway host should have default value")
	}
	if cfg.Gateway.Port == 0 {
		t.Error("Gateway port should have default value")
	}
}

// TestDefaultConfig_Providers verifies provider structure
func TestDefaultConfig_Providers(t *testing.T) {
	cfg := DefaultConfig()

	// Verify all providers are empty by default
	if cfg.Providers.Anthropic.APIKey != "" {
		t.Error("Anthropic API key should be empty by default")
	}
	if cfg.Providers.OpenAI.APIKey != "" {
		t.Error("OpenAI API key should be empty by default")
	}
	if cfg.Providers.OpenRouter.APIKey != "" {
		t.Error("OpenRouter API key should be empty by default")
	}
	if cfg.Providers.Groq.APIKey != "" {
		t.Error("Groq API key should be empty by default")
	}
	if cfg.Providers.Zhipu.APIKey != "" {
		t.Error("Zhipu API key should be empty by default")
	}
	if cfg.Providers.VLLM.APIKey != "" {
		t.Error("VLLM API key should be empty by default")
	}
	if cfg.Providers.Gemini.APIKey != "" {
		t.Error("Gemini API key should be empty by default")
	}
}

// TestDefaultConfig_Channels verifies channels are disabled by default
func TestDefaultConfig_Channels(t *testing.T) {
	cfg := DefaultConfig()

	// Verify all channels are disabled by default
	if cfg.Channels.WhatsApp.Enabled {
		t.Error("WhatsApp should be disabled by default")
	}
	if cfg.Channels.Telegram.Enabled {
		t.Error("Telegram should be disabled by default")
	}
	if cfg.Channels.Feishu.Enabled {
		t.Error("Feishu should be disabled by default")
	}
	if cfg.Channels.Discord.Enabled {
		t.Error("Discord should be disabled by default")
	}
	if cfg.Channels.MaixCam.Enabled {
		t.Error("MaixCam should be disabled by default")
	}
	if cfg.Channels.QQ.Enabled {
		t.Error("QQ should be disabled by default")
	}
	if cfg.Channels.DingTalk.Enabled {
		t.Error("DingTalk should be disabled by default")
	}
	if cfg.Channels.Slack.Enabled {
		t.Error("Slack should be disabled by default")
	}
}

// TestDefaultConfig_WebTools verifies web tools config
func TestDefaultConfig_WebTools(t *testing.T) {
	cfg := DefaultConfig()

	// Verify web tools defaults
	if cfg.Tools.Web.Brave.MaxResults != 5 {
		t.Error("Expected Brave MaxResults 5, got ", cfg.Tools.Web.Brave.MaxResults)
	}
	if cfg.Tools.Web.Brave.APIKey != "" {
		t.Error("Brave API key should be empty by default")
	}
	if cfg.Tools.Web.DuckDuckGo.MaxResults != 5 {
		t.Error("Expected DuckDuckGo MaxResults 5, got ", cfg.Tools.Web.DuckDuckGo.MaxResults)
	}
}

// TestConfig_Complete verifies all config fields are set
func TestConfig_Complete(t *testing.T) {
	cfg := DefaultConfig()

	// Verify complete config structure
	if cfg.Agents.Defaults.Workspace == "" {
		t.Error("Workspace should not be empty")
	}
	if cfg.Agents.Defaults.Model == "" {
		t.Error("Model should not be empty")
	}
	if cfg.Agents.Defaults.Temperature == 0 {
		t.Error("Temperature should have default value")
	}
	if cfg.Agents.Defaults.MaxTokens == 0 {
		t.Error("MaxTokens should not be zero")
	}
	if cfg.Agents.Defaults.MaxToolIterations == 0 {
		t.Error("MaxToolIterations should not be zero")
	}
	if cfg.Gateway.Host != "0.0.0.0" {
		t.Error("Gateway host should have default value")
	}
	if cfg.Gateway.Port == 0 {
		t.Error("Gateway port should have default value")
	}
	if !cfg.Heartbeat.Enabled {
		t.Error("Heartbeat should be enabled by default")
	}
}

func TestLoadConfig_EnvOnlyWithoutFile(t *testing.T) {
	t.Setenv("PICOCLAW_AGENTS_DEFAULTS_MODEL", "openai/gpt-4o-mini")
	t.Setenv("PICOCLAW_AGENTS_DEFAULTS_WORKSPACE", "/data/workspace")
	t.Setenv("PICOCLAW_GATEWAY_PORT", "19999")
	t.Setenv("PICOCLAW_HEARTBEAT_ENABLED", "false")

	cfgPath := filepath.Join(t.TempDir(), "does-not-exist.json")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Agents.Defaults.Model != "openai/gpt-4o-mini" {
		t.Fatalf("expected model from env, got %q", cfg.Agents.Defaults.Model)
	}
	if cfg.Agents.Defaults.Workspace != "/data/workspace" {
		t.Fatalf("expected workspace from env, got %q", cfg.Agents.Defaults.Workspace)
	}
	if cfg.Gateway.Port != 19999 {
		t.Fatalf("expected gateway port from env, got %d", cfg.Gateway.Port)
	}
	if cfg.Heartbeat.Enabled {
		t.Fatal("expected heartbeat disabled from env")
	}
}

func TestLoadConfig_ProviderEnvOverrides(t *testing.T) {
	t.Setenv("PICOCLAW_PROVIDERS_OPENROUTER_API_KEY", "or-key")
	t.Setenv("PICOCLAW_PROVIDERS_OPENROUTER_API_BASE", "https://openrouter.example/v1")
	t.Setenv("PICOCLAW_PROVIDERS_OPENROUTER_PROXY", "http://proxy.internal:8080")
	t.Setenv("PICOCLAW_PROVIDERS_OPENROUTER_AUTH_METHOD", "token")

	cfgPath := filepath.Join(t.TempDir(), "does-not-exist.json")
	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.Providers.OpenRouter.APIKey != "or-key" {
		t.Fatalf("expected provider api key from env, got %q", cfg.Providers.OpenRouter.APIKey)
	}
	if cfg.Providers.OpenRouter.APIBase != "https://openrouter.example/v1" {
		t.Fatalf("expected provider api base from env, got %q", cfg.Providers.OpenRouter.APIBase)
	}
	if cfg.Providers.OpenRouter.Proxy != "http://proxy.internal:8080" {
		t.Fatalf("expected provider proxy from env, got %q", cfg.Providers.OpenRouter.Proxy)
	}
	if cfg.Providers.OpenRouter.AuthMethod != "token" {
		t.Fatalf("expected provider auth method from env, got %q", cfg.Providers.OpenRouter.AuthMethod)
	}
}
