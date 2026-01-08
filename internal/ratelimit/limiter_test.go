package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDomainLimiter_Wait(t *testing.T) {
	// Create a limiter allowing 10 requests per second with burst of 1
	// This means 1 req every 100ms
	limiter := NewDomainLimiter(10, 1)

	ctx := context.Background()
	domain := "example.com"

	// First request should be immediate (burst)
	start := time.Now()
	err := limiter.Wait(ctx, domain)
	assert.NoError(t, err)
	assert.WithinDuration(t, start, time.Now(), 10*time.Millisecond)

	// Second request should wait ~100ms
	start = time.Now()
	err = limiter.Wait(ctx, domain)
	assert.NoError(t, err)
	elapsed := time.Since(start)

	// Check if it waited at least 90ms (allowing some variance)
	assert.True(t, elapsed > 90*time.Millisecond, "Expected wait > 90ms, got %v", elapsed)
}

func TestDomainLimiter_DifferentDomains(t *testing.T) {
	// 1 req/sec -> 1s wait
	limiter := NewDomainLimiter(1, 1)
	ctx := context.Background()

	// Consume burst for domain A
	limiter.Wait(ctx, "a.com")

	// Domain B should still have burst available (immediate)
	start := time.Now()
	err := limiter.Wait(ctx, "b.com")
	assert.NoError(t, err)
	assert.WithinDuration(t, start, time.Now(), 10*time.Millisecond)
}

func TestDomainLimiter_ContextCancel(t *testing.T) {
	limiter := NewDomainLimiter(0.1, 1) // 1 req every 10s
	domain := "slow.com"

	// Consume burst
	limiter.Wait(context.Background(), domain)

	// Next request should wait, but we'll cancel context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx, domain)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline")
}
