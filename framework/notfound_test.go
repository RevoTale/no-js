package framework

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNotFoundError struct{}

func (testNotFoundError) Error() string {
	return "missing"
}

func (testNotFoundError) NotFound() bool {
	return true
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	assert.True(t, IsNotFound(ErrNotFound))
	assert.True(t, IsNotFound(fmt.Errorf("wrapped: %w", ErrNotFound)))
	assert.True(t, IsNotFound(testNotFoundError{}))
	assert.True(t, IsNotFound(fmt.Errorf("wrapped: %w", testNotFoundError{})))
	assert.False(t, IsNotFound(errors.New("boom")))
	assert.False(t, IsNotFound(nil))
}
