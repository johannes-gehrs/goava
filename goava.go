package goava

type (
	Cache[KeyT, ValT comparable] struct {
		cm map[KeyT]ValT
	}
)

func (c *Cache[KeyT, ValT]) Get(key KeyT) (ValT, bool) {
	val, found := c.cm[key]
	return val, found
}

func (c *Cache[KeyT, ValT]) Set(key KeyT, val ValT) {
	c.cm[key] = val
}

func NewCache[KeyT, ValT comparable]() *Cache[KeyT, ValT] {
	c:= &Cache[KeyT, ValT]{
		cm: make(map[KeyT]ValT),
	}

	return c
}
