// internal/engine/dynamic/browser_pool.go
package dynamic

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/rs/zerolog/log"
	"github.com/zulfikawr/crawl/internal/config"
)

// BrowserPool manages a pool of reusable Chrome browser contexts
// This dramatically reduces startup overhead from ~1500ms to ~50ms per request
type BrowserPool struct {
	size        int
	contexts    chan *BrowserContext
	allocCtx    context.Context
	allocCancel context.CancelFunc
	mu          sync.Mutex
	closed      bool
}

// BrowserContext wraps a chromedp context with its cancel function
type BrowserContext struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

// BrowserPoolOptions configures the browser pool
type BrowserPoolOptions struct {
	Size      int
	Headless  bool
	UserAgent string
	Proxy     string
	ExtraArgs []chromedp.ExecAllocatorOption
}

// NewBrowserPool creates a new pool of browser contexts
func NewBrowserPool(opts BrowserPoolOptions) (*BrowserPool, error) {
	if opts.Size <= 0 {
		opts.Size = 3 // Default to 3 contexts
	}
	if opts.Size > 10 {
		opts.Size = 10 // Max 10 contexts to avoid resource exhaustion
	}
	if opts.UserAgent == "" {
		opts.UserAgent = config.DefaultUserAgent
	}

	log.Debug().Int("size", opts.Size).Msg("Creating browser pool")

	// Auto-detect Chrome path
	chromePath := FindChrome()

	// Create allocator options - shared by all contexts
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-breakpad", true),
		chromedp.Flag("disable-client-side-phishing-detection", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-hang-monitor", true),
		chromedp.Flag("disable-ipc-flooding-protection", true),
		chromedp.Flag("disable-prompt-on-repost", true),
		chromedp.Flag("disable-renderer-backgrounding", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("log-level", "3"),
		chromedp.Flag("metrics-recording-only", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("safebrowsing-disable-auto-update", true),
		chromedp.UserAgent(opts.UserAgent),
		// Point 10: Additional optimization flags
		chromedp.Flag("disable-features", "site-per-process,TranslateUI,BlinkGenPropertyTrees"),
		chromedp.Flag("enable-features", "NetworkService,NetworkServiceInProcess"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("window-size", "1920,1080"),
		chromedp.Flag("disk-cache-size", "0"),
		chromedp.Flag("media-cache-size", "0"),
	}

	// Set Chrome path if found
	if chromePath != "" {
		allocOpts = append([]chromedp.ExecAllocatorOption{chromedp.ExecPath(chromePath)}, allocOpts...)
	}

	// Configure headless mode
	if opts.Headless {
		allocOpts = append(allocOpts, chromedp.Flag("headless", "new"))
	} else {
		allocOpts = append(allocOpts, chromedp.Flag("headless", false))
	}

	// Add proxy if specified
	if opts.Proxy != "" {
		allocOpts = append(allocOpts, chromedp.ProxyServer(opts.Proxy))
	}

	// Add extra args
	allocOpts = append(allocOpts, opts.ExtraArgs...)

	// Create parent allocator context
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), allocOpts...)

	pool := &BrowserPool{
		size:        opts.Size,
		contexts:    make(chan *BrowserContext, opts.Size),
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		closed:      false,
	}

	// Pre-create browser contexts concurrently
	var wg sync.WaitGroup
	var errMu sync.Mutex
	var firstErr error

	for i := 0; i < opts.Size; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			browserCtx, browserCancel := chromedp.NewContext(allocCtx)

			// Warm up the context by loading a blank page
			// Use a goroutine to handle timeout without using a child context that might close the target
			done := make(chan error, 1)
			go func() {
				done <- chromedp.Run(browserCtx, chromedp.Navigate("about:blank"))
			}()

			select {
			case err := <-done:
				if err != nil {
					browserCancel()
					errMu.Lock()
					if firstErr == nil {
						firstErr = fmt.Errorf("failed to warm up browser context %d: %w", id, err)
					}
					errMu.Unlock()
					return
				}
			case <-time.After(10 * time.Second):
				browserCancel()
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("timeout warming up browser context %d", id)
				}
				errMu.Unlock()
				return
			}

			// Add to pool - safe because channel is buffered to opts.Size
			pool.contexts <- &BrowserContext{
				Ctx:    browserCtx,
				Cancel: browserCancel,
			}

			log.Debug().Int("context_id", id).Msg("Browser context initialized")
		}(i)
	}

	wg.Wait()

	if firstErr != nil {
		pool.Close()
		return nil, firstErr
	}

	log.Info().Int("pool_size", opts.Size).Msg("Browser pool ready")

	return pool, nil
}

// Acquire gets a browser context from the pool (blocks if none available)
func (bp *BrowserPool) Acquire(timeout time.Duration) (*BrowserContext, error) {
	// Check closed status BEFORE blocking on channel
	bp.mu.Lock()
	closed := bp.closed
	bp.mu.Unlock()

	if closed {
		return nil, fmt.Errorf("browser pool is closed")
	}

	if timeout > 0 {
		select {
		case ctx := <-bp.contexts:
			// Double-check after acquisition
			bp.mu.Lock()
			stillClosed := bp.closed
			bp.mu.Unlock()

			if stillClosed {
				ctx.Cancel()
				return nil, fmt.Errorf("browser pool is closed")
			}
			log.Debug().Msg("Browser context acquired from pool")
			return ctx, nil
		case <-time.After(timeout):
			return nil, fmt.Errorf("timeout waiting for available browser context")
		}
	}

	// No timeout, block until available
	ctx := <-bp.contexts
	// Double-check after acquisition
	bp.mu.Lock()
	stillClosed := bp.closed
	bp.mu.Unlock()

	if stillClosed {
		ctx.Cancel()
		return nil, fmt.Errorf("browser pool is closed")
	}
	log.Debug().Msg("Browser context acquired from pool")
	return ctx, nil
}

// Release returns a browser context to the pool
func (bp *BrowserPool) Release(ctx *BrowserContext) {
	bp.mu.Lock()
	if bp.closed {
		// Pool is closed, cancel the context
		ctx.Cancel()
		bp.mu.Unlock()
		return
	}
	bp.mu.Unlock()

	// Perform cleanup asynchronously to allow the caller to proceed immediately
	go func() {
		// Clean up the context with timeout and proper error handling
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cleanupCancel()

		// Navigate to about:blank to clear state
		// We use the context's own allocator but a fresh timeout
		err := chromedp.Run(ctx.Ctx,
			chromedp.ActionFunc(func(c context.Context) error {
				select {
				case <-cleanupCtx.Done():
					return cleanupCtx.Err()
				default:
					return chromedp.Navigate("about:blank").Do(c)
				}
			}),
		)

		if err != nil {
			// If context was canceled (likely due to shutdown), silence the warning
			if err == context.Canceled || err.Error() == "context canceled" {
				ctx.Cancel()
				return
			}
			log.Warn().Err(err).Msg("Failed to cleanup browser context, will cancel and discard")
			ctx.Cancel()
			return // Don't return to pool
		}

		// Return to pool safely
		bp.mu.Lock()
		defer bp.mu.Unlock()

		if bp.closed {
			ctx.Cancel()
			return
		}

		select {
		case bp.contexts <- ctx:
			log.Debug().Msg("Browser context released to pool")
		default:
			// Pool is full (shouldn't happen), cancel the context
			ctx.Cancel()
			log.Warn().Msg("Browser pool full, discarding context")
		}
	}()
}

// Close shuts down all browser contexts and the allocator
func (bp *BrowserPool) Close() error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.closed {
		return nil
	}
	bp.closed = true

	log.Debug().Msg("Closing browser pool")

	// Cancel allocator FIRST to prevent new context creation
	bp.allocCancel()

	// Close the channel
	close(bp.contexts)

	// Cancel all remaining contexts (drain the channel)
	for ctx := range bp.contexts {
		if ctx != nil && ctx.Cancel != nil {
			ctx.Cancel()
		}
	}

	log.Info().Msg("Browser pool closed")

	return nil
}

// Size returns the pool size
func (bp *BrowserPool) Size() int {
	return bp.size
}

// Available returns the number of available contexts in the pool
func (bp *BrowserPool) Available() int {
	return len(bp.contexts)
}
