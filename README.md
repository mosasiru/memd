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

```memd1.go
package main

import (
	"github.com/douban/libmc/golibmc"
	"github.com/mosasiru/memd"
	"fmt"
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
		fmt.Errorf("%s", err)
	}
	res := Result{}
	if err := c.FromItem(item, &res); err != nil {
		fmt.Errorf("%s", err)
	}
	fmt.Printf("%d", res.Hoge)
}

func getResultFromDB(key string) Result {
	return Result{1}
}
```

This is a simple example for `GetOrSet`.

`GetOrSet` is a useful method combining `Get` and `Set`.

`GetOrSet` works as `Get` from memcached first, and if no hit, `Set` value the callback returns (for example DB result), then return the value.

`ToItem` is a method for generating `*golibmc.Item`. It serialize a value with json as default.

`FromItem` is a method for deserializing `*golibmc.Item`. `FromItem` deserialize it with json as default.


## `GetOrSet` for multi layer memcached

```memd2.go
package main

import (
	"github.com/douban/libmc/golibmc"
	"github.com/mosasiru/memd"
	"fmt"
)

type Result struct {
	Hoge int    `json:"hoge"`
}

func main() {
	localMemd := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))
	remoteMemd := memd.New(golibmc.SimpleNew([]string{"localhost:11211"}))

	item, err := localMemd.GetOrSet("key", func(key string) (*golibmc.Item, error) {
		item, err := remoteMemd.GetOrSet(key, func(key string) (*golibmc.Item, error) {
			return remoteMemd.ToItem(key, getResultFromDB(key), 10)
		})
		if err != nil {
			fmt.Errorf("%s", err)
		}
		item.Expiration = 1
		return item, nil
	})
	if err != nil {
		fmt.Errorf("%s", err)
	}
	res := Result{}
	if err := localMemd.FromItem(item, &res); err != nil {
		fmt.Errorf("%s", err)
	}
	fmt.Printf("%d", res.Hoge)
}
```

This is a `GetOrSet` example for multi layer memcached, specifically local memcached and remote memcached.

`GetOrSet` can chain transparently. This example change local memcached's expiration time (ttl) between local and remote memcached.


## `GetOrSetMulti`

Consider that we want cache a SQL query such that `select * from hoge where id in (100,200,300...)` for each id. In other words, another query could be `select * from hoge where id in (100,400...)`. so we want to cache as per id, not as whole result. In such a case, we use `GetMulti` for ids first, and if no hit key exists,  ask SQL for remain ids, then `SetMulti` DB values.

`GetOrSetMulti` makes easier such a situation. Below is the example.



## `SetLogger`

## `SetSerializer`
