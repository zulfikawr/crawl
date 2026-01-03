package config

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

func validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if c.HTTPTimeout <= 0 {
		return fmt.Errorf("http timeout must be > 0")
	}
	if c.HTTPTimeout > 5*time.Minute {
		return fmt.Errorf("http timeout too large (max 5 minutes)")
	}

	if c.BrowserPoolSize <= 0 || c.BrowserPoolSize > DefaultMaxBrowserPoolSize {
		return fmt.Errorf("browser pool size must be between 1 and %d", DefaultMaxBrowserPoolSize)
	}

	if c.CacheMaxSizeBytes <= 0 {
		return fmt.Errorf("cache max size must be > 0")
	}
	if c.CacheMaxSizeBytes > 10*1024*1024*1024 { // 10GB
		return fmt.Errorf("cache max size too large (max 10GB)")
	}

	if c.StaticRateLimitRPS <= 0 || c.StaticRateLimitRPS > 1000 {
		return fmt.Errorf("static rate limit must be between 0 and 1000 requests/second")
	}

	if c.StaticRateLimitBurst <= 0 || c.StaticRateLimitBurst > 1000 {
		return fmt.Errorf("static rate limit burst must be between 0 and 1000")
	}

	if c.DynamicRateLimitRPS <= 0 || c.DynamicRateLimitRPS > 1000 {
		return fmt.Errorf("dynamic rate limit must be between 0 and 1000 requests/second")
	}

	if c.DynamicRateLimitBurst <= 0 || c.DynamicRateLimitBurst > 1000 {
		return fmt.Errorf("dynamic rate limit burst must be between 0 and 1000")
	}

	if c.UserAgent == "" {
		log.Warn().Msg("User agent is empty, using default")
		c.UserAgent = DefaultUserAgent
	}

	if c.ChromePath != "" {
		if _, err := os.Stat(c.ChromePath); err != nil {
			log.Warn().Str("path", c.ChromePath).Msg("Chrome path specified but not found")
		}
	}

	return nil
}
