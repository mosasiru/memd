package memd

import (
	"log"
	"github.com/douban/libmc/golibmc"
)

type Client struct {
	*golibmc.Client
	logf func(format string, params ...interface{})
}

func New(m *golibmc.Client) *Client {
	return &Client{m, log.Printf}
}

func (c *Client) SetLogger(logf func(format string, params ...interface{})) {
	c.logf = logf
}

func (c *Client) GetOrSet(key string, cb func(key string)(*golibmc.Item, error)) (*golibmc.Item, error) {
	item, err := c.Get(key)
	if err != nil {
		if err != golibmc.ErrCacheMiss {
			return nil, err
		}
	} else {
		return item, nil
	}

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
	for key, _ := range(item_map){
		hit_keys = append(hit_keys, key)
		gotmap[key] = true
	}
	c.logf("hit keys: %s", hit_keys)
	remain_keys := []string{}
	for _, key := range(keys) {
		if _, ok := gotmap[key]; ! ok {
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
	for key, item := range(cb_item_map){
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
