# cache

This code package defines the cache interface and the most basic cache implementation.

## Features

### [interface.go](interface.go)

This file defines the methods that the cache should implement.

It contains two definitions:
`Interface` and `KeyGetter`.

`Interface` requires that the cache should implement four methods:

- `Get(KeyGetter) error`
- `Set(KeyGetter, any, ...ItemOption) error`
- `Delete(KeyGetter) error`
- `Clear() error`

The `KeyGetter` interface involved in the above method needs to be implemented:

- `GetKey() string`

`KeyGetter` is suitable for complex key scenarios.
Therefore, you can decide the specific implementation of the `GetKey()` method based on actual needs.

For example, consider a Key structure containing multiple attributes.
If the absence of a cache item can be determined solely based on the first attribute,
there is no need to consider the remaining attributes.

On the other hand, the attributes and their order within the Key structure can also vary.
Therefore, we abstract the KeyGetter interface and Cache interface here,
leaving the implementation details of how to compose keys up to you.

### [item.go](item.go) & [item_option.go](item_option.go)

`item.go` defines the cache item. The cache item interface needs to implement the `Value()` and `Expired()` methods.

`item_option.go` defines the options for declaring cache items. Currently, the options are designed only for expiration time.

### [memory.go](memory.go)

`memory.go` defines the simplest in-memory cache. 


This functionality solely provides implementation samples for cache definition and is not recommended for actual production use.
This is because the functionality employs pointers to cache items, which would necessitate periodic scanning by the garbage collector,
thus potentially lowering performance.

## Installation

Ensure you have installed the `go-dag` package:

```shell
go get github.com/rhosocial/go-dag/workflow/standard/cache
```

## Usage

### Example

Here's a basic example of how to use the cache package:

```go
package main

import (
	"fmt"
	"time"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
)

// MyKey defines a struct that implements the KeyGetter interface
type MyKey struct {
	id string
	cache.KeyGetter
}

func (k MyKey) GetKey() string {
	return k.id
}

func main() {
	// Create an in-memory cache
	memoryCache := cache.NewMemoryCache()

	// Create a key
	key := MyKey{id: "exampleKey"}

	// Set a cache item
	err := memoryCache.Set(key, "exampleValue", cache.WithTTL(time.Second))
	if err != nil {
		fmt.Println("Error setting cache:", err)
		return
	}

	// Get a cache item
	value, err := memoryCache.Get(key)
	if err != nil {
		fmt.Println("Error getting cache:", err)
		return
	}
	fmt.Println("Cached value:", value)

	// Delete a cache item
	err = memoryCache.Delete(key)
	if err != nil {
		fmt.Println("Error deleting cache:", err)
		return
	}

	// Clear the cache
	err = memoryCache.Clear()
	if err != nil {
		fmt.Println("Error clearing cache:", err)
		return
	}
}
```

In this example, we define a simple `MyKey` struct to implement the KeyGetter interface.
We then create an in-memory cache and perform cache set, get, delete, and clear operations.