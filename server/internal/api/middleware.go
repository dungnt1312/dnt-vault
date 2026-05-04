package api

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dnt/vault-server/internal/auth"
)

const maxRequestBodyBytes = 10 * 1024 * 1024 // 10 MB

type contextKey string

const usernameKey contextKey = "username"

type Middleware struct {
	auth         *auth.AuthService
	loginLimiter *loginRateLimiter
}

func NewMiddleware(auth *auth.AuthService) *Middleware {
	return &Middleware{
		auth:         auth,
		loginLimiter: newLoginRateLimiter(5, time.Minute),
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (m *Middleware) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, rw.status, time.Since(start).Round(time.Millisecond))
	})
}

func (m *Middleware) BodyLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			respondError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}

		token := parts[1]
		username, err := m.auth.ValidateToken(token)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), usernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *Middleware) LoginRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if !m.loginLimiter.allow(ip) {
			respondError(w, http.StatusTooManyRequests, "too many login attempts, try again later")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type loginAttempt struct {
	count     int
	resetTime time.Time
}

type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*loginAttempt
	maxTries int
	window   time.Duration
}

func newLoginRateLimiter(maxTries int, window time.Duration) *loginRateLimiter {
	rl := &loginRateLimiter{
		attempts: make(map[string]*loginAttempt),
		maxTries: maxTries,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *loginRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, a := range rl.attempts {
			if now.After(a.resetTime) {
				delete(rl.attempts, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *loginRateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	a, ok := rl.attempts[ip]
	if !ok || now.After(a.resetTime) {
		rl.attempts[ip] = &loginAttempt{count: 1, resetTime: now.Add(rl.window)}
		return true
	}
	if a.count >= rl.maxTries {
		return false
	}
	a.count++
	return true
}
