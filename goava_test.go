package goava

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCachePutGetAndDelete(t *testing.T) {
	c := NewCache[string, int64]()

	c.Put("johannes", 16, time.Now().Add(time.Hour))

	res, found := c.Get("johannes")
	assert.True(t, found)
	assert.EqualValues(t, 16, res)

	assert.Equal(t, 1, c.itemsByExpiration.Len())

	found = c.Delete("johannes")
	assert.True(t, found)

	assert.Equal(t, 0, c.itemsByExpiration.Len())

	_, found = c.Get("johannes")
	assert.False(t, found)
}

func TestCacheEvictExpiredKey(t *testing.T) {
	c := NewCache[string, int64]()

	for i := 0; i < 10*1000; i++ {
		c.Put(fmt.Sprintf("%016d", i), int64(i), time.Now().Add(time.Hour).Add(time.Second*time.Duration(i)))
	}

	for i := 10 * 1000; i < 20*1000; i++ {
		c.Put(fmt.Sprintf("%016d", i), int64(i), time.Now().Add(-time.Hour).Add(-(time.Second * time.Duration(i))))
	}

	evicted := c.EvictExpiredKeys(context.TODO())

	assert.EqualValues(t, 10*1000, evicted)
	assert.Equal(t, 10*1000, c.itemsByExpiration.Len())
}

func TestCacheIdenticalExpirationTimes(t *testing.T) {
	c := NewCache[string, int64]()

	now := time.Now()

	for i := 0; i < 10*1000; i++ {
		c.Put(fmt.Sprintf("%016d", i), int64(i), now.Add(time.Hour))
	}

	for i := 10 * 1000; i < 20*1000; i++ {
		c.Put(fmt.Sprintf("%016d", i), int64(i), now.Add(-time.Hour))
	}

	assert.Equal(t, 2, c.itemsByExpiration.Len())

	evicted := c.EvictExpiredKeys(context.TODO())

	assert.EqualValues(t, 10000, evicted)
	assert.Equal(t, 1, c.itemsByExpiration.Len())
}
