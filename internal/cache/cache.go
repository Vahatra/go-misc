package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/allegro/bigcache/v3"
)

type Cache struct {
	cache *bigcache.BigCache
}

// newBigCache returns a new BigCache struct
func NewCache(ctx context.Context) (*Cache, error) {
	// When cache load can be predicted in advance then it is better to use custom initialization
	// because additional memory allocation can be avoided in that way.
	config := bigcache.Config{
		// number of shards (must be a power of 2)
		Shards: 128,

		// time after which entry can be evicted
		LifeWindow: 5 * time.Minute,

		// Interval between removing expired entries (clean up).
		// If set to <= 0 then no action is performed.
		// Setting to < 1 second is counterproductive — bigcache has a one second resolution.
		CleanWindow: 1 * time.Minute,

		// rps * lifeWindow, used only in initial memory allocation
		MaxEntriesInWindow: 1000 * 10 * 60,

		// max entry size in bytes, used only in initial memory allocation
		MaxEntrySize: 500,

		// prints information about additional memory allocation
		Verbose: false,

		// cache will not allocate more memory than this limit, value in MB
		// if value is reached then the oldest entries can be overridden for the new ones
		// 0 value means no size limit
		HardMaxCacheSize: 128,

		// callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A bitmask representing the reason will be returned.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		OnRemove: nil,

		// OnRemoveWithReason is a callback fired when the oldest entry is removed because of its expiration time or no space left
		// for the new entry, or because delete was called. A constant representing the reason will be passed through.
		// Default value is nil which means no callback and it prevents from unwrapping the oldest entry.
		// Ignored if OnRemove is specified.
		OnRemoveWithReason: nil,
	}

	c, err := bigcache.New(ctx, config)
	if err != nil {
		return &Cache{}, err
	}
	return &Cache{cache: c}, nil
}

// Set inserts the key/value pair into the cache.
// Only the exported fields of the given struct will be
// serialized and stored
func (c *Cache) Set(key string, value interface{}) error {
	valueBytes, err := encode(value)
	if err != nil {
		return err
	}

	return c.cache.Set(key, valueBytes)
}

// Get returns the value correlating to the key in the cache
func (c *Cache) Get(key string) (interface{}, error) {
	// Get the value in the byte format it is stored in
	valueBytes, err := c.cache.Get(key)
	if err != nil {
		return nil, err
	}

	// Deserialize the bytes of the value
	value, err := decode(valueBytes)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func encode(value interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	gob.Register(value)

	err := enc.Encode(&value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decode(valueBytes []byte) (interface{}, error) {
	var value interface{}
	buf := bytes.NewBuffer(valueBytes)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&value)
	if err != nil {
		return nil, err
	}

	return value, nil
}
