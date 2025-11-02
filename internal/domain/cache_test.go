package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrCacheLocked(t *testing.T) {
	t.Run("エラーメッセージが正しい", func(t *testing.T) {
		err := ErrCacheLocked

		assert.NotNil(t, err)
		assert.Equal(t, "cache file is locked by another process", err.Error())
	})
}

func TestErrCacheCorrupted(t *testing.T) {
	t.Run("エラーメッセージが正しい", func(t *testing.T) {
		err := ErrCacheCorrupted

		assert.NotNil(t, err)
		assert.Equal(t, "cache file is corrupted", err.Error())
	})
}

func TestErrCachePermission(t *testing.T) {
	t.Run("エラーメッセージが正しい", func(t *testing.T) {
		err := ErrCachePermission

		assert.NotNil(t, err)
		assert.Equal(t, "permission denied for cache file or directory", err.Error())
	})
}

func TestErrCacheDirectoryCreate(t *testing.T) {
	t.Run("エラーメッセージが正しい", func(t *testing.T) {
		err := ErrCacheDirectoryCreate

		assert.NotNil(t, err)
		assert.Equal(t, "failed to create cache directory", err.Error())
	})
}
