package goava

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/armon/go-radix"
)

type (
	item[ValT any] struct {
		expiresAt time.Time
		val       ValT
	}

	Cache[KeyT comparable, ValT any] struct {
		cm                                 map[KeyT]item[ValT]
		itemsByExpiration                  *radix.Tree
		mutex                              *sync.Mutex
		onlyOneEvictionProcessAtAGivenTime *sync.Once
	}

)

func isExpired(now, expiryTime time.Time) bool {
	return !(expiryTime.After(now))
}

func (it *item[ValT]) expiryTimeAsString() string {
	return it.expiresAt.UTC().Format(time.RFC3339Nano)
}

func expiryTimeFromString(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(fmt.Sprintf("unable to parse time in tree-datastructure as time.Time - this is unexpected. String is '%s'",
			s))
	}

	return t
}

func (c *Cache[KeyT, ValT]) Get(key KeyT) (ValT, bool) {
	var zeroVal ValT

	item, found := c.cm[key]
	if !found || isExpired(time.Now(), item.expiresAt) {
		return zeroVal, false
	}

	return item.val, found
}

func (c *Cache[KeyT, ValT]) Put(key KeyT, val ValT, expiresAt time.Time) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	itm := item[ValT]{
		expiresAt: expiresAt,
		val:       val,
	}

	c.cm[key] = itm

	treeValues, found := c.itemsByExpiration.Get(itm.expiryTimeAsString())
	if !found {
		treeValues = make(map[KeyT]struct{})
	}

	treeValues.(map[KeyT]struct{})[key] = struct {}{}

	c.itemsByExpiration.Insert(itm.expiryTimeAsString(), treeValues)
}

func (c *Cache[KeyT, ValT]) Delete(key KeyT) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	itm, found := c.cm[key]
	if !found {
		return false
	}

	delete(c.cm, key)

	treeValues, found := c.itemsByExpiration.Get(itm.expiryTimeAsString())
	if !found {
		treeValues = make(map[KeyT]struct{})
	}

	treeValuesTypeAsserted := treeValues.(map[KeyT]struct{})

	delete(treeValuesTypeAsserted, key)

	if len(treeValuesTypeAsserted) == 0 {
		c.itemsByExpiration.Delete(itm.expiryTimeAsString())
	} else {
		c.itemsByExpiration.Insert(itm.expiryTimeAsString(), treeValuesTypeAsserted)
	}

	return true
}

func (c *Cache[KeyT, ValT]) EvictExpiredKeys(ctx context.Context) (evictedCount int64) {
	c.onlyOneEvictionProcessAtAGivenTime.Do(func() {
		c.mutex.Lock()
		toBeEvicted := make([]KeyT, 0)
		c.itemsByExpiration.Walk(func(s string, cacheKeyMap interface{}) bool {
			select {
			case <-ctx.Done():
				// timeout reached - we'll stop walking the tree but still remove the items we collected
				return true
			default:
				if isExpired(time.Now(), expiryTimeFromString(s)) {
					for cacheKey := range cacheKeyMap.(map[KeyT]struct{}) {
						toBeEvicted = append(toBeEvicted, cacheKey)
					}
					return false
				}
				// if the key isn't expired all the next ones also won't be expired as they are sorted by eviction time in the tree.
				// So we can stop iteration.
				return true
			}
		})
		c.mutex.Unlock()

		for _, cacheKey := range toBeEvicted {
			c.Delete(cacheKey)
			evictedCount++
		}
	})

	c.onlyOneEvictionProcessAtAGivenTime = &sync.Once{}

	return evictedCount
}

func NewCache[KeyT comparable, ValT any]() *Cache[KeyT, ValT] {
	c := &Cache[KeyT, ValT]{
		cm:                                 make(map[KeyT]item[ValT]),
		mutex:                              &sync.Mutex{},
		itemsByExpiration:                  radix.New(),
		onlyOneEvictionProcessAtAGivenTime: &sync.Once{},
	}

	return c
}
