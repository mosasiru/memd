package memd

import (
	"github.com/douban/libmc/golibmc"
	"github.com/ugorji/go/codec"
	"log"
)

// Client is wrapper of *golibmc.Client.
type Client struct {
	*golibmc.Client
	logf       func(format string, params ...interface{})
	serializer codec.Handle
}

// New create memd Client
func New(m *golibmc.Client) *Client {
	var jh codec.JsonHandle
	return &Client{m, log.Printf, &jh}
}

// SetLogger change its logger
func (c *Client) SetLogger(logf func(format string, params ...interface{})) *Client {
	c.logf = logf
	return c
}

// SetSerializer change its serialize method. default is json.
func (c *Client) SetSerializer(h codec.Handle) *Client {
	c.serializer = h
	return c
}

// GetOrSet ... Get from memcached, and if no hit, Set value gotten by callback, and return the value
func (c *Client) GetOrSet(key string, cb func(key string) (*golibmc.Item, error)) (*golibmc.Item, error) {
	item, err := c.Get(key)
	if err != nil {
		if err != golibmc.ErrCacheMiss {
			return nil, err
		}
	} else {
		c.logf("hit: %s", key)
		return item, nil
	}
	c.logf("no hit: %s", key)
	item, err = cb(key)
	if err != nil {
		return nil, err
	}
	c.Set(item)
	return item, nil
}

// GetOrSetMulti ... GetMulti from memcached, and if no hit key exists,SetMulti memcached values gotten by callback, and return all values
func (c *Client) GetOrSetMulti(keys []string, cb func(keys []string) (map[string]*golibmc.Item, error)) (map[string]*golibmc.Item, error) {
	itemMap, err := c.GetMulti(keys)
	if err != nil && err != golibmc.ErrCacheMiss {
		return nil, err
	}
	if itemMap == nil {
		itemMap = map[string]*golibmc.Item{}
	}
	// divide keys into hitKeys and remainKeys
	hitKeys := []string{}
	gotMap := map[string]bool{}
	for key := range itemMap {
		hitKeys = append(hitKeys, key)
		gotMap[key] = true
	}
	c.logf("hit keys: %s", hitKeys)
	remainKeys := []string{}
	for _, key := range keys {
		if _, ok := gotMap[key]; !ok {
			remainKeys = append(remainKeys, key)
		}
	}
	c.logf("remain keys: %s", remainKeys)
	if len(remainKeys) == 0 {
		return itemMap, nil
	}

	// get items respond to remain keys from callback
	cbItemMap, err := cb(remainKeys)
	if err != nil {
		return nil, err
	}
	cbItems := []*golibmc.Item{}
	for key, item := range cbItemMap {
		cbItems = append(cbItems, item)
		itemMap[key] = item
	}
	if len(cbItems) == 0 {
		return itemMap, nil
	}

	// cache items gotten by callback
	failedKeys, err := c.SetMulti(cbItems)
	if err != nil {
		return nil, err
	}
	if len(failedKeys) != 0 {
		c.logf("failed keys: %s", failedKeys)
	}

	return itemMap, nil
}

// ToItem ... serialize value and build *golibmc.Item
func (c *Client) ToItem(key string, _val interface{}, exp int64) (*golibmc.Item, error) {
	val := make([]byte, 0, 64)
	err := codec.NewEncoderBytes(&val, c.serializer).Encode(_val)
	if err != nil {
		return nil, err
	}
	return &golibmc.Item{
		Key:        key,
		Value:      val,
		Expiration: exp,
	}, nil
}

// FromItem ... deserialize item.Value
func (c *Client) FromItem(item *golibmc.Item, val interface{}) error {
	return codec.NewDecoderBytes(item.Value, c.serializer).Decode(val)
}

// ToItemMap ... serialize values and build *golibmc.Item map
func (c *Client) ToItemMap(keyToValue map[string]interface{}, exp int64) (map[string]*golibmc.Item, error) {
	itemMap := map[string]*golibmc.Item{}
	var err error
	for key, val := range keyToValue {
		itemMap[key], err = c.ToItem(key, val, exp)
		if err != nil {
			return itemMap, err
		}
	}
	return itemMap, nil
}
