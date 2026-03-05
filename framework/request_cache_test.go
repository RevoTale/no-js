package framework

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedCall_DeduplicatesConcurrentCalls(t *testing.T) {
	t.Parallel()

	ctx := WithRequestCache(context.Background())
	release := make(chan struct{})
	started := make(chan struct{})
	var startedOnce sync.Once

	callCount := 0
	var mu sync.Mutex
	fn := func(context.Context) (string, error) {
		mu.Lock()
		callCount++
		mu.Unlock()
		startedOnce.Do(func() { close(started) })
		<-release
		return "ok", nil
	}

	resultCh := make(chan string, 2)
	errCh := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			value, err := CachedCall(ctx, "dedupe-key", fn)
			if err != nil {
				errCh <- err
				return
			}
			resultCh <- value
		}()
	}

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("cached call did not start")
	}

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	assert.Equal(t, 1, callCount)
	mu.Unlock()

	close(release)

	for i := 0; i < 2; i++ {
		select {
		case err := <-errCh:
			require.NoError(t, err)
		case value := <-resultCh:
			assert.Equal(t, "ok", value)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for cached call result")
		}
	}
}

func TestCachedCall_WithoutRequestCacheExecutesEveryTime(t *testing.T) {
	t.Parallel()

	callCount := 0
	fn := func(context.Context) (int, error) {
		callCount++
		return callCount, nil
	}

	first, err := CachedCall(context.Background(), "same-key", fn)
	require.NoError(t, err)
	second, err := CachedCall(context.Background(), "same-key", fn)
	require.NoError(t, err)

	assert.Equal(t, 1, first)
	assert.Equal(t, 2, second)
}
