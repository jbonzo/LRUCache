package cache

import (
	"errors"
	"math"
	"time"

	"github.com/jonboulle/clockwork"
)

var (
	ErrDataNotInBackingStore = errors.New("the data requested is not in the backing backingStore")
)

// make a cache struct
// insert
// get
// size
// remaining space
// write through policy

// NewLRUCache creates a Cache pointer given a specified size. It uses a real clock to determine LRU
func NewLRUCache(size int) *Cache {
	return &Cache{
		size:  size,
		items: make(map[string]*CacheItem, size),
		backingStore: &hardStorage{
			store: make(map[string]interface{}),
		},
		clock: clockwork.NewRealClock(),
	}
}

// NewLRUCache creates a Cache pointer given a specified size. It uses the provided clock to determine LRU
func NewLRUCacheWithClock(size int, clock clockwork.Clock) *Cache {
	c := NewLRUCache(size)
	c.clock = clock
	return c
}

// Cache uses the LRU replacement model to implement an in memory cache
type Cache struct {
	size         int
	items        map[string]*CacheItem // maps item tag to item
	backingStore *hardStorage

	clock clockwork.Clock

	uses   int
	hits   int
	misses int
}

// CacheItem represents an item stored in the cache
type CacheItem struct {
	data     interface{}
	tag      string
	lastUsed time.Time
}

// AddItem adds an item to the cache. We assume the cache has been initialized
func (c *Cache) AddItem(tag string, data interface{}) {
	c.uses++

	// check if item is in cache
	item, hit := c.searchCache(tag)

	if hit {
		// update item and add to backing backingStore
		c.hits++
		item.data = data
		c.backingStore.addItem(tag, data)
		item.updateCacheItemLRU(c.clock)
		return
	}

	c.misses++

	// if not in cache check backing store
	_, hit = c.backingStore.getItem(tag)
	if !hit {
		// if not in backingStore then add it to backingStore
		c.backingStore.addItem(tag, data)
	}

	// no matter what update cache
	item.updateCacheItemLRU(c.clock)
	c.updateCache(tag, data)
	return
}

// GetItem returns the item with the request tag. If not present in any backingStore returns ErrDataNotInBackingStore
func (c *Cache) GetItem(tag string) (*CacheItem, error) {
	c.uses++

	item, hit := c.searchCache(tag)

	if hit {
		c.hits++
		item.updateCacheItemLRU(c.clock)
		return item, nil
	}

	c.misses++

	// if not check backing backingStore
	data, hit := c.backingStore.getItem(tag)
	if !hit {
		return nil, ErrDataNotInBackingStore
	}

	// if in backing backingStore and not cache then make new item and replace cache
	item = c.updateCache(tag, data)
	item.updateCacheItemLRU(c.clock)
	return item, nil
}

/* -------------------------- Helpers A.K.A. The Meat ----------------------------------------------------- */

// replace cache will create a new item given the tag and data. If the cache is full it will evict the lru.
// returns CacheItem used to update Cache
func (c *Cache) updateCache(tag string, data interface{}) *CacheItem {
	newItem := &CacheItem{
		tag:  tag,
		data: data,
	}

	if len(c.items) >= c.size {
		lru := c.findLRU()
		c.evict(newItem, lru)
	}

	return newItem
}

// evict removes the evictedResident and replaces it with the newResident
func (c *Cache) evict(newResident, evictedResident *CacheItem) {
	// TODO: if we change write policies then we need to account for this logic here to prevent data loss
	delete(c.items, evictedResident.tag)
	c.items[newResident.tag] = newResident
}

// searchCache returns the item and whether or not it was found. It also returns the LRU
func (c *Cache) searchCache(tag string) (item *CacheItem, hit bool) {
	item, hit = c.items[tag]
	return
}

// findLRU returns the LeastRecentlyUsed item in the cache
func (c *Cache) findLRU() *CacheItem {
	var lru *CacheItem

	curUnixTime := int64(math.MaxInt64)

	for _, item := range c.items {
		// if this item is after than curUnixTime then
		itemsUnixTime := item.lastUsed.UnixNano()
		if itemsUnixTime < curUnixTime {
			lru = item
			curUnixTime = itemsUnixTime
		}
	}

	return lru
}

func (c *CacheItem) updateCacheItemLRU(clock clockwork.Clock) {
	c.lastUsed = clock.Now().UTC()
}

// for testing mostly
func (c *Cache) resetCounters() {
	c.uses = 0
	c.misses = 0
	c.hits = 0
}

/* -------------------------------------------------------------------------------------------------------- */

type hardStorage struct {
	store map[string]interface{} // maps item tag to data
}

func (b *hardStorage) addItem(tag string, item interface{}) {
	b.store[tag] = item
}

func (b *hardStorage) getItem(tag string) (item interface{}, hit bool) {
	item, hit = b.store[tag]
	return
}
