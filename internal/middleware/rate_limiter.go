package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// implements a simple per-IP token bucket rate limiter.
type tokenBucket struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	rate     int           // tokens added per window
	window   time.Duration // refill window
	capacity int           // max burst
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

func newTokenBucket(rate int, window time.Duration) *tokenBucket {
	tb := &tokenBucket{
		buckets:  make(map[string]*bucket),
		rate:     rate,
		window:   window,
		capacity: rate,
	}
	// Periodic cleanup goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			tb.cleanup()
		}
	}()
	return tb
}

func (tb *tokenBucket) allow(key string) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	b, ok := tb.buckets[key]
	if !ok {
		b = &bucket{tokens: tb.capacity, lastRefill: time.Now()}
		tb.buckets[key] = b
	}

	// Refill based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed/tb.window) * tb.rate
	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > tb.capacity {
			b.tokens = tb.capacity
		}
		b.lastRefill = now
	}

	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}

func (tb *tokenBucket) cleanup() {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	threshold := time.Now().Add(-10 * time.Minute)
	for k, b := range tb.buckets {
		if b.lastRefill.Before(threshold) {
			delete(tb.buckets, k)
		}
	}
}

// limiters for different endpoint classes.
var (
	// 10 requests per minute per IP (brute-force protection)
	authLimiter = newTokenBucket(10, time.Minute)
	// 20 uploads per minute per IP
	uploadLimiter = newTokenBucket(20, time.Minute)
	// 300 requests per minute per IP
	apiLimiter = newTokenBucket(300, time.Minute)
)

func limitWith(tb *tokenBucket) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !tb.allow(ip) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "too many requests",
				"retry_after": "60s",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// applies strict rate limiting suitable for login/register endpoints.
func RateLimitAuth() gin.HandlerFunc { return limitWith(authLimiter) }

func RateLimitUpload() gin.HandlerFunc { return limitWith(uploadLimiter) }

func RateLimitAPI() gin.HandlerFunc { return limitWith(apiLimiter) }
