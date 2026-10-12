package http

import (
	"sync"
	"time"
)

type requestLimiter struct {
	entries   map[string]requestLimit
	limit     int
	mu        sync.Mutex
	window    time.Duration
	cleanupAt time.Time
}

type loginLimiter struct {
	byEmail, byIP *requestLimiter
}

func newLoginLimiter(limit int, window time.Duration) *loginLimiter {
	return &loginLimiter{byEmail: newRequestLimiter(limit, window), byIP: newRequestLimiter(limit, window)}
}

func (l *loginLimiter) allow(ip, email string) (bool, int) {
	allowedIP, retryIP := l.byIP.allow(ip)
	if !allowedIP {
		return false, retryIP
	}
	allowedEmail, retryEmail := l.byEmail.allow(email)
	if allowedIP && allowedEmail {
		return true, 0
	}
	if retryIP > retryEmail {
		return false, retryIP
	}
	return false, retryEmail
}

type requestLimit struct {
	count   int
	resetAt time.Time
}

func newRequestLimiter(limit int, window time.Duration) *requestLimiter {
	return &requestLimiter{entries: make(map[string]requestLimit), limit: limit, window: window}
}

func (l *requestLimiter) allow(key string) (bool, int) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	if !now.Before(l.cleanupAt) {
		for entryKey, entry := range l.entries {
			if !entry.resetAt.After(now) {
				delete(l.entries, entryKey)
			}
		}
		l.cleanupAt = now.Add(time.Second)
	}

	entry := l.entries[key]
	if !entry.resetAt.IsZero() && !entry.resetAt.After(now) {
		delete(l.entries, key)
		entry = requestLimit{}
	}
	if entry.resetAt.IsZero() && len(l.entries) >= 4096 {
		return false, 1
	}
	if entry.resetAt.IsZero() {
		entry.resetAt = now.Add(l.window)
	}
	if entry.count >= l.limit {
		retryAfter := int(time.Until(entry.resetAt).Seconds())
		if retryAfter < 1 {
			retryAfter = 1
		}
		return false, retryAfter
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}
