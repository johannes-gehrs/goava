package goava

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCache(t *testing.T) {
	c := NewCache[string, int64]()

	c.Set("johannes", 16)

	res, found := c.Get("johannes")
	assert.True(t, found)
	assert.EqualValues(t, 16, res)
}
