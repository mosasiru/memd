package memd

import (
	"github.com/douban/libmc/golibmc"
	"github.com/ugorji/go/codec"
	"log"
	"testing"
	"time"
)

type Hoge struct {
	Piyo int    `json:"hoge"`
	Fuga string `json:"fuga"`
}

func HogeGetOrSet(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	c.SetLogger(log.Printf)
	ck := "key1"
	item, err := c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Hoge{1, "aaa"}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	val := Hoge{}
	if err := c.FromItem(item, &val); err != nil {
		t.Error(err)
	}
	if val.Piyo != 1 || val.Fuga != "aaa" {
		t.Error("invalid origin")
	}

	item, err = c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Hoge{}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	val = Hoge{}
	if err := c.FromItem(item, &val); err != nil {
		t.Error(err)
	}
	if val.Piyo != 1 || val.Fuga != "aaa" {
		t.Error("invalid cache")
	}

	time.Sleep(1 * time.Second)
	if _, err = c.Get(ck); err != golibmc.ErrCacheMiss {
		t.Error("cache should be expired: %s", err)
	}
}

func HogeGetOrSetMulti(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	c.SetLogger(log.Printf)
	keys := []string{"key1", "key2"}
	keyToHoge := map[string]Hoge{
		"key1": Hoge{1, "aaa"},
		"key2": Hoge{2, "bbb"},
	}

	item, err := c.ToItem(keys[0], keyToHoge[keys[0]], 1)
	if err != nil {
		t.Error(err)
	}
	if err = c.Set(item); err != nil {
		t.Error(err)
	}
	itemMap, err := c.GetOrSetMulti(keys, func(keys []string) (map[string]*golibmc.Item, error) {
		if len(keys) != 1 {
			t.Error("cache should be only one")
		}
		key := keys[0]
		return c.ToItemMap(map[string]interface{}{key: keyToHoge[key]}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	if len(itemMap) != 2 {
		t.Error("result should be only two")
	}
	for key, item := range itemMap {
		var val Hoge
		if err = c.FromItem(item, &val); err != nil {
			t.Error(err)
		}
		exp := keyToHoge[key]
		if val.Piyo != exp.Piyo || val.Fuga != exp.Fuga {
			t.Error("invalid cache")
		}
	}

	time.Sleep(1 * time.Second)
	if _, err = c.Get(keys[0]); err != golibmc.ErrCacheMiss {
		t.Error("cache should be expired: %s", err)
	}
}

func HogeSerializer(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	var mh codec.MsgpackHandle
	c.SetSerializer(&mh)
	c.SetLogger(log.Printf)
	ck := "key2"
	item, err := c.ToItem(ck, Hoge{1, "aaa"}, 1)
	if err != nil {
		t.Error(err)
	}
	if err = c.Set(item); err != nil {
		t.Error(err)
	}
	val := Hoge{}
	item, err = c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Hoge{}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	val = Hoge{}
	if err := c.FromItem(item, &val); err != nil {
		t.Error(err)
	}
	if val.Piyo != 1 || val.Fuga != "aaa" {
		t.Error("invalid cache")
	}
}
