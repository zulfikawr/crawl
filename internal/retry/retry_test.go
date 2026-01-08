package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithRetry_Success(t *testing.T) {
	attempts := 0
	op := func() error {
		attempts++
		return nil
	}

	cfg := Config{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     1.0,
	}

	err := WithRetry(context.Background(), cfg, op)
	assert.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestWithRetry_FailAndRecover(t *testing.T) {
	attempts := 0
	op := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary fail")
		}
		return nil
	}

	cfg := Config{
		MaxAttempts:    5,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     1.0,
	}

	err := WithRetry(context.Background(), cfg, op)
	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestWithRetry_MaxRetriesExceeded(t *testing.T) {
	attempts := 0
	op := func() error {
		attempts++
		return errors.New("fail")
	}

	cfg := Config{
		MaxAttempts:    3,
		InitialBackoff: 1 * time.Millisecond,
		MaxBackoff:     10 * time.Millisecond,
		Multiplier:     1.0,
	}

	err := WithRetry(context.Background(), cfg, op)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation failed after 3 attempts")
	assert.Equal(t, 3, attempts)
}

func TestWithRetry_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	op := func() error {
		attempts++
		cancel() // Cancel context during first attempt
		return errors.New("fail")
	}

	cfg := Config{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond, // Long backoff to ensure we catch cancellation
		MaxBackoff:     100 * time.Millisecond,
		Multiplier:     1.0,
	}

	err := WithRetry(ctx, cfg, op)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, attempts)
}
