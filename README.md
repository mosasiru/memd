# memd

`memd` is a memcached useful client wrapping https://github.com/douban/libmc

`memd` focuses on transparent managing chain for multi layer memcached.
The main methods are `GetOrSet` and `GetOrSetMulti`.

## Installing

### Using *go get*

    $ go get github.com/mosasiru/memd

After this command *mutex* is ready to use. Its source will be in:

    $GOPATH/src/github.com/mosasiru/memd

# Example

## `GetOrSet`

```go
package main

import (
	"github.com/douban/libmc/golibmc"
	"github.com/mosasiru/memd"
	"log"
)

type Result struct {
	Hoge int    `json:"hoge"`
}

func main() {
	c := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))

	item, err := c.GetOrSet("key", func(key string) (*golibmc.Item, error) {
		return c.ToItem(key, getResultFromDB(key), 10)
	})
	if err != nil {
		log.Fatalf("%s", err)
	}
	res := Result{}
	if err := c.FromItem(item, &res); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%d", res.Hoge)
}

func getResultFromDB(key string) Result {
	return Result{1} // dummy
}

```

This is a simple example for `GetOrSet`.

`GetOrSet` is a useful method combining `Get` and `Set`.

`GetOrSet` works as `Get` from memcached first, and if no hit, `Set` value the callback returns (for example DB result), then return the value.

`ToItem` is a method for generating `*golibmc.Item`. It serialize a value with json as default.

`FromItem` is a method for deserializing `*golibmc.Item`. `FromItem` deserialize it with json as default.


## `GetOrSet` for multi layer memcached

```go
package main

import (
	"github.com/douban/libmc/golibmc"
	"github.com/mosasiru/memd"
	"log"
)

type Result struct {
	Hoge int    `json:"hoge"`
}

func main() {
	localMemd := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))
	remoteMemd := memd.New(golibmc.SimpleNew([]string{"remotehost:11211"}))

    item, err := localMemd.GetOrSet("key", func(key string) (*golibmc.Item, error) {
		return remoteMemd.GetOrSet(key, func(key string) (*golibmc.Item, error) {
			return remoteMemd.ToItem(key, getResultFromDB(key), 1)
		})
	})
	if err != nil {
		log.Fatalf("%s", err)
	}
	res := Result{}
	if err := localMemd.FromItem(item, &res); err != nil {
		log.Fatalf("%s", err)
	}
	log.Printf("%d", res.Hoge)
}

func getResultFromDB(key string) Result {
	return Result{1} // dummy
}

```

This is a `GetOrSet` example for multi layer memcached, specifically local memcached and remote memcached.
`GetOrSet` can chain transparently.


## `GetOrSetMulti`

Consider that we want cache a SQL query such that `select * from hoge where id in (100,200,300...)` for each id. In other words, another query could be `select * from hoge where id in (100,400...)`. so we want to cache as per id, not as whole result. In such a case, we use `GetMulti` for ids first, and if no hit key exists,  ask SQL for remain ids, then `SetMulti` DB values.

`GetOrSetMulti` makes easier such a situation. Below is the example.

```go
package main

import (
	"github.com/douban/libmc/golibmc"
	"github.com/mosasiru/memd"
	"log"
)

type Result struct {
	Hoge int   `json:"hoge"`
}

func main() {
	c := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))
	keys := []string{"key1", "key2", "key3"}
	itemMap, err := c.GetOrSetMulti(keys, func(keys []string) (map[string]*golibmc.Item, error) {
		return c.ToItemMap(getResultMapFromDB(keys), 1)
	})
	if err != nil {
		log.Fatalf("%s", err)
	}
	for key, item := range itemMap {
		var res Result
		if err = c.FromItem(item, &res); err != nil {
			log.Fatalf("%s", err)
		}
		log.Printf("%s: %d", key, res.Hoge)
	}

}

func getResultMapFromDB(keys []string) map[string]interface{} {
	// dummy
	res := map[string]interface{}{}
	for _, key := range keys {
		res[key] = Result{1}
	}
	return res
}
```

`GetOrSetMulti` can chain as same as `GetOrSet`.
Below code is an example.

```go
itemMap, err := localMemd.GetOrSetMulti(keys, func(keys []string) (map[string]*golibmc.Item, error) {
    return remoteMemd.GetOrSetMulti(keys, func(keys []string) (map[string]*golibmc.Item, error) {
        return remoteMemd.ToItemMap(getResultMapFromDB(keys), 1)
    })
})
```


## `SetLogger`

```go
c := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))
c.SetLogger(log.Printf)
```

`GetOrSet` and `GetOrSetMulti` hit state is logged. the default logger is `log.Printf`. you can change that logger.

## `SetSerializer`

`ToItem`, `ToItemMap`, and `FromItem` use json as a serializer. you can use other serializer with https://github.com/ugorji/go  `Handle`. json, msgpack, binc, and so on.
