package goava

import (
	"sync"
	"time"

	"github.com/armon/go-radix"
)

type (
	item[ValT comparable] struct {
		expiresAt time.Time
		val       ValT
	}

	Cache[KeyT, ValT comparable] struct {
		cm                map[KeyT]item[ValT]
		itemsByExpiration *radix.Tree
		mutex             *sync.Mutex
	}
)

func (it *item[ValT]) isExpired(now time.Time) bool {
	return !(it.expiresAt.After(now))
}

func (c *Cache[KeyT, ValT]) Get(key KeyT) (ValT, bool) {
	var zeroVal ValT

	item, found := c.cm[key]
	if !found || item.isExpired(time.Now()) {
		return zeroVal, false
	}

	return item.val, found
}

func (c *Cache[KeyT, ValT]) Set(key KeyT, val ValT, expiresAt time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cm[key] = item[ValT]{
		expiresAt: expiresAt,
		val:       val,
	}
	c.itemsByExpiration.Insert(expiresAt.UTC().Format(time.RFC3339Nano), key)
}

func NewCache[KeyT, ValT comparable](maxItems, maxEvictionsPerSecond int64) *Cache[KeyT, ValT] {
	c := &Cache[KeyT, ValT]{
		cm:                make(map[KeyT]item[ValT]),
		mutex:             &sync.Mutex{},
		itemsByExpiration: radix.New(),
	}

	return c
}
