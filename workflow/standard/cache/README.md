# cache

This code package defines the cache interface and the simplest cache implementation.

## [interface.go](interface.go)

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

## [item.go](item.go) & [item_option.go](item_option.go)

`item.go` defines the cache item. The cache item interface needs to implement the `Value()` and `Expired()` methods.

`item_option.go` defines the options for declaring cache items. Currently, the options are designed only for expiration time.

## [memory.go](memory.go)

`memory.go` defines the simplest in-memory cache. 

This implementation is only used for demonstration and testing purposes of the cache interface and cache item interface.