package cache

import "time"

type ItemInterface interface {
	Expired() bool
	Value() any
}

type Item struct {
	value     any
	expiredAt *time.Time
	ItemInterface
}

func (i *Item) Expired() bool {
	if i.expiredAt == nil {
		return false
	}
	return i.expiredAt.Before(time.Now())
}

func (i *Item) Value() any {
	return i.value
}

func NewItem(value any) *Item {
	return &Item{value: value}
}
