package proxy

import (
	"sync"
	"time"
)

// ProxyPool manages a list of proxies with rotation and health checking
type ProxyPool struct {
	proxies []string
	index   int
	mu      sync.Mutex
	failed  map[string]time.Time
}

// NewProxyPool creates a new ProxyPool
func NewProxyPool(proxies []string) *ProxyPool {
	return &ProxyPool{
		proxies: proxies,
		failed:  make(map[string]time.Time),
	}
}

// GetNext returns the next healthy proxy from the pool
func (p *ProxyPool) GetNext() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.proxies) == 0 {
		return ""
	}

	// Try to find a healthy proxy
	attempts := 0
	maxAttempts := len(p.proxies)

	for attempts < maxAttempts {
		attempts++
		proxy := p.proxies[p.index]
		p.index = (p.index + 1) % len(p.proxies)

		// Check if failed recently
		if failTime, ok := p.failed[proxy]; ok {
			if time.Since(failTime) < 5*time.Minute {
				// Still considered failed, try next
				continue
			}
			// Failure expired
			delete(p.failed, proxy)
		}

		return proxy
	}

	// All proxies failed recently, return least-recently-failed
	var oldestProxy string
	oldestTime := time.Now()
	for proxy, failTime := range p.failed {
		if failTime.Before(oldestTime) {
			oldestTime = failTime
			oldestProxy = proxy
		}
	}
	if oldestProxy != "" {
		return oldestProxy
	}

	// Fallback: return first proxy
	return p.proxies[0]
}

// MarkFailed marks a proxy as failed so it will be skipped for a while
func (p *ProxyPool) MarkFailed(proxy string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failed[proxy] = time.Now()
}

// MarkHealthy clears the failure status of a proxy
func (p *ProxyPool) MarkHealthy(proxy string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.failed, proxy)
}
