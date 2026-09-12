package middleware

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lp/campus-market/internal/constants"
	"github.com/lp/campus-market/internal/util"
)

type bucket struct {
	count   int
	resetAt time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
}

// NewRateLimiter builds a limiter allowing `limit` requests per window.
func NewRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, buckets: map[string]*bucket{}}
}

func (r *rateLimiter) allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	b, ok := r.buckets[key]
	if !ok || now.After(b.resetAt) {
		r.buckets[key] = &bucket{count: 1, resetAt: now.Add(r.window)}
		return true
	}
	b.count++
	return b.count <= r.limit
}

// RateLimit returns a Gin middleware that limits per client IP.
func RateLimit(limiter *rateLimiter, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.allow(key) {
			logger.Warn(fmt.Sprintf(constants.LogRateLimitReached, key, c.FullPath()))
			util.Fail(c, 429, constants.CodeRateLimited, constants.MsgRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
