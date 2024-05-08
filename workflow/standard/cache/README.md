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

