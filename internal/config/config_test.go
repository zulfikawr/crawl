package config

import (
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	// Create a dummy command
	cmd := &cobra.Command{}

	cfg, err := Load(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify defaults
	assert.Equal(t, DefaultLogLevel, cfg.LogLevel)
	assert.Equal(t, DefaultHTTPTimeout, cfg.HTTPTimeout)
	assert.Equal(t, DefaultUserAgent, cfg.UserAgent)
	assert.Equal(t, DefaultBrowserPoolSize, cfg.BrowserPoolSize)
}

func TestLoad_EnvVars(t *testing.T) {
	// Set env vars
	os.Setenv("CRAWL_USER_AGENT", "EnvAgent/1.0")
	os.Setenv("CRAWL_PROXY", "http://env-proxy:8080")
	defer os.Unsetenv("CRAWL_USER_AGENT")
	defer os.Unsetenv("CRAWL_PROXY")

	cmd := &cobra.Command{}

	cfg, err := Load(cmd)
	assert.NoError(t, err)

	assert.Equal(t, "EnvAgent/1.0", cfg.UserAgent)
	assert.Equal(t, "http://env-proxy:8080", cfg.Proxy)
}

func TestLoad_Flags(t *testing.T) {
	// Setup command with flags
	cmd := &cobra.Command{}
	RegisterFlags(cmd)

	// Parse flags
	err := cmd.ParseFlags([]string{
		"--user-agent", "FlagAgent/1.0",
		"--timeout", "10s",
		"--json",
		"--verbose",
	})
	assert.NoError(t, err)

	cfg, err := Load(cmd)
	assert.NoError(t, err)

	assert.Equal(t, "FlagAgent/1.0", cfg.UserAgent)
	assert.Equal(t, 10*time.Second, cfg.HTTPTimeout)
	assert.True(t, cfg.JSONLog)
	assert.Equal(t, "debug", cfg.LogLevel) // verbose sets log level to debug
}

func TestValidate_Success(t *testing.T) {
	cfg := &Config{
		HTTPTimeout:           5 * time.Second,
		BrowserPoolSize:       1,
		StaticRateLimitRPS:    1,
		StaticRateLimitBurst:  1,
		DynamicRateLimitRPS:   1,
		DynamicRateLimitBurst: 1,
		CacheMaxSizeBytes:     1024,
	}
	err := validate(cfg)
	assert.NoError(t, err)
}

func TestValidate_Errors(t *testing.T) {
	tests := []struct {
		name        string
		cfg         *Config
		errContains string
	}{
		{
			name: "Negative Timeout",
			cfg: &Config{
				HTTPTimeout:       -1 * time.Second,
				CacheMaxSizeBytes: 1024,
			},
			errContains: "timeout",
		},
		{
			name: "Zero Browser Pool",
			cfg: &Config{
				HTTPTimeout:       1 * time.Second,
				BrowserPoolSize:   0,
				CacheMaxSizeBytes: 1024,
			},
			errContains: "pool size",
		},
		{
			name: "Negative RPS",
			cfg: &Config{
				HTTPTimeout:          1 * time.Second,
				BrowserPoolSize:      1,
				StaticRateLimitRPS:   -5,
				StaticRateLimitBurst: 1,
				CacheMaxSizeBytes:    1024,
			},
			errContains: "rate limit must be between",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validate(tc.cfg)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tc.errContains)
		})
	}
}
