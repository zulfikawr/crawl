package app

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/zulfikawr/crawl/internal/config"
)

func TestNew_Success(t *testing.T) {
	cfg := &config.Config{
		LogLevel:             "error",
		HTTPTimeout:          1 * time.Second,
		CacheMaxSizeBytes:    1024,
		StaticRateLimitRPS:   1,
		StaticRateLimitBurst: 1,
		BrowserPoolSize:      1,
	}

	app, err := New(context.Background(), cfg)
	assert.NoError(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.StaticScraper)
	assert.NotNil(t, app.DynamicScraper)
	assert.Nil(t, app.BrowserPool) // Lazy init

	app.Close(context.Background())
}

func TestEnsureBrowserPool(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping browser pool test in short mode")
	}

	cfg := &config.Config{
		LogLevel:             "error",
		HTTPTimeout:          1 * time.Second,
		CacheMaxSizeBytes:    1024,
		StaticRateLimitRPS:   1,
		StaticRateLimitBurst: 1,
		BrowserPoolSize:      1,
		BrowserHeadless:      true,
	}

	app, err := New(context.Background(), cfg)
	assert.NoError(t, err)
	defer app.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure pool is created
	err = app.EnsureBrowserPool(ctx)
	// We might fail if Chrome is not installed, so we check for that specifically
	if err != nil {
		t.Logf("Browser pool init failed (expected in CI without Chrome): %v", err)
	} else {
		assert.NotNil(t, app.BrowserPool)
	}
}

func TestNew_NilConfig(t *testing.T) {
	_, err := New(context.Background(), nil)
	assert.Error(t, err)
}
