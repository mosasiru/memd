package memd

import (
	"github.com/douban/libmc/golibmc"
	"github.com/ugorji/go/codec"
	"log"
)

type Client struct {
	*golibmc.Client
	logf       func(format string, params ...interface{})
	serializer codec.Handle
}

func New(m *golibmc.Client) *Client {
	var mh codec.MsgpackHandle
	return &Client{m, log.Printf, &mh}
}

func (c *Client) SetLogger(logf func(format string, params ...interface{})) *Client {
	c.logf = logf
	return c
}

func (c *Client) SetSerializer(h codec.Handle) *Client {
	c.serializer = h
	return c
}

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

func (c *Client) GetOrSetMulti(keys []string, cb func(keys []string) (map[string]*golibmc.Item, error)) (map[string]*golibmc.Item, error) {
	item_map, err := c.GetMulti(keys)
	if err != nil && err != golibmc.ErrCacheMiss {
		return nil, err
	}
	if item_map == nil {
		item_map = map[string]*golibmc.Item{}
	}
	hit_keys := []string{}
	gotmap := map[string]bool{}
	for key, _ := range item_map {
		hit_keys = append(hit_keys, key)
		gotmap[key] = true
	}
	c.logf("hit keys: %s", hit_keys)
	remain_keys := []string{}
	for _, key := range keys {
		if _, ok := gotmap[key]; !ok {
			remain_keys = append(remain_keys, key)
		}
	}
	c.logf("remain keys: %s", remain_keys)
	if len(remain_keys) == 0 {
		return item_map, nil
	}

	cb_item_map, err := cb(remain_keys)
	if err != nil {
		return nil, err
	}
	cb_items := []*golibmc.Item{}
	for key, item := range cb_item_map {
		cb_items = append(cb_items, item)
		item_map[key] = item
	}
	failed_keys, err := c.SetMulti(cb_items)
	if err != nil {
		return nil, err
	}
	if len(failed_keys) != 0 {
		c.logf("failed keys: %s", failed_keys)
	}

	return item_map, nil
}

func (c *Client) ToItem(key string, cb func() (interface{}, error), exp int64) (*golibmc.Item, error) {
	_val, err := cb()
	if err != nil {
		return nil, err
	}

	val := make([]byte, 0, 64)
	codec.NewEncoderBytes(&val, c.serializer).Encode(_val)
	if err != nil {
		return nil, err
	}
	return &golibmc.Item{
		Key:        key,
		Value:      val,
		Expiration: exp,
	}, nil
}

func (c *Client) FromItem(item *golibmc.Item, val interface{}) error {
	return codec.NewDecoderBytes(item.Value, c.serializer).Decode(val)
}
