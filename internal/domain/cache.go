package domain

import (
	"errors"
	"time"
)

// RecommendEntry represents a single cache entry for recommended articles
type RecommendEntry struct {
	URL      string    `json:"url"`
	Title    string    `json:"title"`
	PostedAt time.Time `json:"posted_at"`
}

// RecommendCache provides an interface for managing recommend article cache
type RecommendCache interface {
	// Initialize initializes the cache, loads existing data and acquires necessary locks
	Initialize() error

	// IsCached checks if the given URL is already cached (duplicate check)
	IsCached(url string) bool

	// AddEntry adds a new entry to the cache with the given URL and title
	AddEntry(url, title string) error

	// Close closes the cache, releases locks and performs cleanup
	Close() error
}

// Cache-related errors
var (
	// ErrCacheLocked is returned when cache file is already locked by another process
	ErrCacheLocked = errors.New("cache file is locked by another process")

	// ErrCacheCorrupted is returned when cache file is corrupted and cannot be read
	ErrCacheCorrupted = errors.New("cache file is corrupted")

	// ErrCachePermission is returned when there are permission issues with cache file or directory
	ErrCachePermission = errors.New("permission denied for cache file or directory")

	// ErrCacheDirectoryCreate is returned when cache directory cannot be created
	ErrCacheDirectoryCreate = errors.New("failed to create cache directory")
)
