package cache

import "github.com/canpok1/ai-feed/internal/domain"

// NopCache is a no-operation implementation of RecommendCache interface.
// It implements the Null Object pattern, doing nothing for all operations.
type NopCache struct{}

// NewNopCache creates a new NopCache instance
func NewNopCache() *NopCache {
	return &NopCache{}
}

// Initialize does nothing for NopCache
func (n *NopCache) Initialize() error {
	return nil
}

// IsCached always returns false for NopCache
func (n *NopCache) IsCached(url string) bool {
	return false
}

// AddEntry does nothing for NopCache
func (n *NopCache) AddEntry(url, title string) error {
	return nil
}

// Close does nothing for NopCache
func (n *NopCache) Close() error {
	return nil
}

// Verify that NopCache implements RecommendCache interface
var _ domain.RecommendCache = (*NopCache)(nil)
