package framework

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

type requestCacheContextKey struct{}

type requestCache struct {
	mu      sync.Mutex
	entries map[string]*requestCacheEntry
}

type requestCacheEntry struct {
	done chan struct{}
	val  any
	err  error
}

// WithRequestCache injects a request-scoped cache into the context.
func WithRequestCache(ctx context.Context) context.Context {
	if ctx == nil {
		return ctx
	}
	if _, ok := ctx.Value(requestCacheContextKey{}).(*requestCache); ok {
		return ctx
	}
	return context.WithValue(ctx, requestCacheContextKey{}, &requestCache{
		entries: make(map[string]*requestCacheEntry),
	})
}

// CachedCall executes fn once per key in the request-scoped cache and shares
// the result with concurrent callers using the same context.
func CachedCall[T any](
	ctx context.Context,
	key string,
	fn func(context.Context) (T, error),
) (T, error) {
	var zero T
	if fn == nil {
		return zero, fmt.Errorf("cached call requires a function")
	}
	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" || ctx == nil {
		return fn(ctx)
	}

	cache, ok := ctx.Value(requestCacheContextKey{}).(*requestCache)
	if !ok || cache == nil {
		return fn(ctx)
	}

	cache.mu.Lock()
	entry, exists := cache.entries[trimmedKey]
	if !exists {
		entry = &requestCacheEntry{done: make(chan struct{})}
		cache.entries[trimmedKey] = entry
		cache.mu.Unlock()

		entry.val, entry.err = fn(ctx)
		close(entry.done)

		if entry.err != nil {
			return zero, entry.err
		}

		typed, castOK := entry.val.(T)
		if !castOK {
			return zero, fmt.Errorf("cached call type mismatch for key %q", trimmedKey)
		}
		return typed, nil
	}
	cache.mu.Unlock()

	select {
	case <-entry.done:
	case <-ctx.Done():
		return zero, ctx.Err()
	}

	if entry.err != nil {
		return zero, entry.err
	}
	typed, castOK := entry.val.(T)
	if !castOK {
		return zero, fmt.Errorf("cached call type mismatch for key %q", trimmedKey)
	}
	return typed, nil
}
