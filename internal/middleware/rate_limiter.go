package middleware

import (
	"net/http"
)

// RateLimiter defines the interface for rate limiting middleware
type RateLimiter interface {
	Limit(next http.Handler) http.Handler
}
