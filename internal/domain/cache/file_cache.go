package cache

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/canpok1/ai-feed/internal/domain"
	"github.com/canpok1/ai-feed/internal/domain/entity"
)

// FileRecommendCache implements RecommendCache interface using JSON Lines file format
type FileRecommendCache struct {
	filePath string
	lockPath string
	urlSet   map[string]bool
	entries  []domain.RecommendEntry
	config   *entity.CacheConfig
	lockFile *os.File
}

// NewFileRecommendCache creates a new FileRecommendCache instance
func NewFileRecommendCache(config *entity.CacheConfig) *FileRecommendCache {
	lockPath := config.FilePath + ".lock"
	return &FileRecommendCache{
		filePath: config.FilePath,
		lockPath: lockPath,
		urlSet:   make(map[string]bool),
		entries:  make([]domain.RecommendEntry, 0),
		config:   config,
	}
}

// Initialize initializes the cache, loads existing data and acquires necessary locks
func (c *FileRecommendCache) Initialize() error {
	slog.Debug("Initializing file recommend cache", "file_path", c.filePath)

	// Create directory first (needed for lock file)
	if err := c.createDirectoryIfNotExists(); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Acquire lock after directory creation
	if err := c.acquireLock(); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	// Load existing cache data
	if err := c.loadFromFile(); err != nil {
		// Release lock on error (ignore release error as load already failed)
		_ = c.releaseLock()
		return fmt.Errorf("failed to load cache data: %w", err)
	}

	// Cleanup old entries
	c.cleanup()

	slog.Debug("File recommend cache initialized successfully",
		"entries_count", len(c.entries),
		"unique_urls", len(c.urlSet))

	return nil
}

// IsCached checks if the given URL is already cached (duplicate check)
func (c *FileRecommendCache) IsCached(url string) bool {
	normalizedURL := c.normalizeURL(url)
	return c.urlSet[normalizedURL]
}

// AddEntry adds a new entry to the cache with the given URL and title
func (c *FileRecommendCache) AddEntry(url, title string) error {
	normalizedURL := c.normalizeURL(url)

	// Check if already exists
	if c.urlSet[normalizedURL] {
		slog.Debug("URL already in cache, skipping", "url", normalizedURL)
		return nil
	}

	// Create new entry
	entry := domain.RecommendEntry{
		URL:      normalizedURL,
		Title:    title,
		PostedAt: time.Now(),
	}

	// Add to in-memory structures
	c.entries = append(c.entries, entry)
	c.urlSet[normalizedURL] = true

	// Cleanup if necessary (FIFO when max_entries exceeded)
	c.cleanupByMaxEntries()

	// Save to file
	if err := c.saveToFile(); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	slog.Debug("Added entry to cache", "url", normalizedURL, "title", title)
	return nil
}

// Close closes the cache, releases locks and performs cleanup
func (c *FileRecommendCache) Close() error {
	if c.lockFile != nil {
		if err := c.releaseLock(); err != nil {
			return fmt.Errorf("failed to release lock: %w", err)
		}
		slog.Debug("File recommend cache closed successfully")
	}
	return nil
}

// createDirectoryIfNotExists creates the cache directory if it doesn't exist
func (c *FileRecommendCache) createDirectoryIfNotExists() error {
	dir := filepath.Dir(c.filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		slog.Debug("Creating cache directory", "dir", dir)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return domain.ErrCacheDirectoryCreate
		}
	}
	return nil
}

// loadFromFile loads cache entries from the JSON Lines file
func (c *FileRecommendCache) loadFromFile() error {
	// Check if file exists
	if _, err := os.Stat(c.filePath); os.IsNotExist(err) {
		slog.Debug("Cache file does not exist, starting with empty cache", "file_path", c.filePath)
		return nil
	}

	file, err := os.Open(c.filePath)
	if err != nil {
		if os.IsPermission(err) {
			return domain.ErrCachePermission
		}
		return fmt.Errorf("failed to open cache file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	validEntries := make([]domain.RecommendEntry, 0)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry domain.RecommendEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Skip invalid lines (old format compatibility)
			slog.Warn("Skipping invalid cache entry line", "error", err.Error())
			continue
		}

		validEntries = append(validEntries, entry)
		normalizedURL := c.normalizeURL(entry.URL)
		c.urlSet[normalizedURL] = true
	}

	if err := scanner.Err(); err != nil {
		return domain.ErrCacheCorrupted
	}

	c.entries = validEntries
	slog.Debug("Loaded cache entries from file", "count", len(validEntries))
	return nil
}

// saveToFile saves all cache entries to the JSON Lines file
func (c *FileRecommendCache) saveToFile() error {
	// Create temporary file for atomic write
	tempPath := c.filePath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		if os.IsPermission(err) {
			return domain.ErrCachePermission
		}
		return fmt.Errorf("failed to create temporary cache file: %w", err)
	}

	var success bool
	defer func() {
		// On error, close the file (if not already closed) and remove the temp file.
		// On success, the file is already closed and the temp file is renamed, so this is a no-op.
		file.Close()
		if !success {
			os.Remove(tempPath)
		}
	}()

	// Write entries as JSON Lines
	for _, entry := range c.entries {
		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal cache entry: %w", err)
		}
		if _, err := file.WriteString(string(data) + "\n"); err != nil {
			return fmt.Errorf("failed to write cache entry: %w", err)
		}
	}

	// Sync to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync cache file: %w", err)
	}

	// Must close before rename.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temporary cache file: %w", err)
	}

	// Atomic replace
	if err := os.Rename(tempPath, c.filePath); err != nil {
		return fmt.Errorf("failed to replace cache file: %w", err)
	}

	success = true
	slog.Debug("Saved cache entries to file", "count", len(c.entries))
	return nil
}

// acquireLock acquires the cache file lock
func (c *FileRecommendCache) acquireLock() error {
	// Ensure lock file directory exists
	lockDir := filepath.Dir(c.lockPath)
	if _, err := os.Stat(lockDir); os.IsNotExist(err) {
		slog.Debug("Creating lock file directory", "dir", lockDir)
		if err := os.MkdirAll(lockDir, 0755); err != nil {
			return fmt.Errorf("failed to create lock file directory: %w", err)
		}
	}

	lockFile, err := os.OpenFile(c.lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%w: lock file exists at %s. If no other ai-feed process is running, manually delete the lock file", domain.ErrCacheLocked, c.lockPath)
		}
		return fmt.Errorf("failed to create lock file: %w", err)
	}

	c.lockFile = lockFile
	slog.Debug("Acquired cache lock", "lock_path", c.lockPath)
	return nil
}

// releaseLock releases the cache file lock
func (c *FileRecommendCache) releaseLock() error {
	if c.lockFile != nil {
		c.lockFile.Close()
		if err := os.Remove(c.lockPath); err != nil && !os.IsNotExist(err) {
			slog.Warn("Failed to remove lock file", "lock_path", c.lockPath, "error", err.Error())
			return err
		}
		c.lockFile = nil
		slog.Debug("Released cache lock", "lock_path", c.lockPath)
	}
	return nil
}

// cleanup removes old entries based on retention_days
func (c *FileRecommendCache) cleanup() {
	if c.config.RetentionDays <= 0 {
		return
	}

	cutoffTime := time.Now().AddDate(0, 0, -c.config.RetentionDays)
	validEntries := make([]domain.RecommendEntry, 0)
	removedCount := 0

	for _, entry := range c.entries {
		if entry.PostedAt.After(cutoffTime) {
			validEntries = append(validEntries, entry)
		} else {
			// Remove expired entry from URL set
			normalizedURL := c.normalizeURL(entry.URL)
			delete(c.urlSet, normalizedURL)
			removedCount++
		}
	}

	if removedCount > 0 {
		c.entries = validEntries
		slog.Debug("Cleaned up old cache entries", "removed_count", removedCount, "remaining_count", len(validEntries))
	}
}

// cleanupByMaxEntries removes old entries when max_entries is exceeded (FIFO)
func (c *FileRecommendCache) cleanupByMaxEntries() {
	if c.config.MaxEntries <= 0 || len(c.entries) <= c.config.MaxEntries {
		return
	}

	// Calculate how many entries to remove
	excessCount := len(c.entries) - c.config.MaxEntries

	// Remove oldest entries (FIFO) and their corresponding URLs from urlSet
	removedEntries := c.entries[:excessCount]
	for _, entry := range removedEntries {
		normalizedURL := c.normalizeURL(entry.URL)
		delete(c.urlSet, normalizedURL)
	}
	c.entries = c.entries[excessCount:]

	slog.Debug("Cleaned up excess cache entries", "removed_count", len(removedEntries), "remaining_count", len(c.entries))
}

// normalizeURL normalizes URL by removing trailing slash
func (c *FileRecommendCache) normalizeURL(url string) string {
	return strings.TrimSuffix(url, "/")
}
