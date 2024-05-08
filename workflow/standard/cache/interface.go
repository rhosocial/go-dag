package cache

type KeyGetter interface {
	GetKey() string
}

type Interface interface {
	Get(key KeyGetter) (any, error)
	Set(key KeyGetter, value any, options ...ItemOption) error
	Delete(key KeyGetter) error
	Clear() error
}
