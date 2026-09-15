package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"erp/internal/authorization/ports"
	"github.com/google/uuid"
)

type cacheEntry struct {
	permissions []string
	expiresAt   time.Time
}

type MemoryPolicyCache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func NewMemoryPolicyCache(ttl time.Duration) ports.PolicyCache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &MemoryPolicyCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *MemoryPolicyCache) cacheKey(userID uuid.UUID, legalEntityID *uuid.UUID) string {
	if legalEntityID == nil || *legalEntityID == uuid.Nil {
		return fmt.Sprintf("%s:global", userID.String())
	}
	return fmt.Sprintf("%s:%s", userID.String(), legalEntityID.String())
}

func (c *MemoryPolicyCache) Get(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.cacheKey(userID, legalEntityID)
	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	// Return a copy to avoid mutation
	result := make([]string, len(entry.permissions))
	copy(result, entry.permissions)
	return result, true
}

func (c *MemoryPolicyCache) Set(ctx context.Context, userID uuid.UUID, legalEntityID *uuid.UUID, permissions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.cacheKey(userID, legalEntityID)
	c.entries[key] = cacheEntry{
		permissions: permissions,
		expiresAt:   time.Now().Add(c.ttl),
	}
}

func (c *MemoryPolicyCache) InvalidateUser(ctx context.Context, userID uuid.UUID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := userID.String() + ":"
	for k := range c.entries {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.entries, k)
		}
	}
}

func (c *MemoryPolicyCache) InvalidateAll(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}
