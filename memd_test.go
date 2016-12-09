package memd

import (
	"log"
	"testing"
	"time"

	"github.com/douban/libmc/golibmc"
	"github.com/ugorji/go/codec"
)

type Result struct {
	Hoge int    `json:"hoge"`
	Fuga string `json:"fuga"`
}

func ResultGetOrSet(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	c.SetLogger(log.Printf)
	ck := "key1"
	item, err := c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Result{1, "aaa"}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	res := Result{}
	if err := c.FromItem(item, &res); err != nil {
		t.Error(err)
	}
	if res.Hoge != 1 || res.Fuga != "aaa" {
		t.Error("invalid origin")
	}

	item, err = c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Result{}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	res = Result{}
	if err := c.FromItem(item, &res); err != nil {
		t.Error(err)
	}
	if res.Hoge != 1 || res.Fuga != "aaa" {
		t.Error("invalid cache")
	}

	time.Sleep(1 * time.Second)
	if _, err = c.Get(ck); err != golibmc.ErrCacheMiss {
		t.Error("cache should be expired", err)
	}
}

func ResultGetOrSetMulti(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	c.SetLogger(log.Printf)
	keys := []string{"key1", "key2"}
	keyToResult := map[string]Result{
		"key1": {1, "aaa"},
		"key2": {2, "bbb"},
	}

	item, err := c.ToItem(keys[0], keyToResult[keys[0]], 1)
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
		return c.ToItemMap(map[string]interface{}{key: keyToResult[key]}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	if len(itemMap) != 2 {
		t.Error("result should be only two")
	}
	for key, item := range itemMap {
		var res Result
		if err = c.FromItem(item, &res); err != nil {
			t.Error(err)
		}
		exp := keyToResult[key]
		if res.Hoge != exp.Hoge || res.Fuga != exp.Fuga {
			t.Error("invalid cache")
		}
	}

	time.Sleep(1 * time.Second)
	if _, err = c.Get(keys[0]); err != golibmc.ErrCacheMiss {
		t.Error("cache should be expired", err)
	}
}

func ResultSerializer(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))
	var mh codec.MsgpackHandle
	c.SetSerializer(&mh)
	c.SetLogger(log.Printf)
	ck := "key2"
	item, err := c.ToItem(ck, Result{1, "aaa"}, 1)
	if err != nil {
		t.Error(err)
	}
	if err = c.Set(item); err != nil {
		t.Error(err)
	}
	res := Result{}
	item, err = c.GetOrSet(ck, func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, Result{}, 1)
	})
	if err != nil {
		t.Error(err)
	}
	res = Result{}
	if err := c.FromItem(item, &res); err != nil {
		t.Error(err)
	}
	if res.Hoge != 1 || res.Fuga != "aaa" {
		t.Error("invalid cache")
	}
}

func TestResultFromItem(t *testing.T) {
	c := New(golibmc.SimpleNew([]string{"localhost:11211"}))

	tests := []struct {
		b   []byte
		err error
	}{
		{[]byte(``), ErrEmptyValue},
		{[]byte(`{}`), nil},
		{[]byte(`{"foo":"bar"}`), nil},
	}

	for _, tt := range tests {
		item := &golibmc.Item{Value: tt.b}
		var v map[string]interface{}
		err := c.FromItem(item, &v)
		if err != tt.err {
			t.Errorf("FromItem(%+v, %s) = %#v", item, v, err)
		}
	}
}
