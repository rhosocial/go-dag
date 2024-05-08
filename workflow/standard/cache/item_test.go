package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewItem(t *testing.T) {
	item := NewItem(int64(1))
	assert.NotNil(t, item)

	assert.Equal(t, item.Value(), int64(1))
	assert.Equal(t, item.Expired(), false)
}

func TestItem_Expired(t *testing.T) {
	expiredAt := time.Now().Add(-time.Second)
	item := NewItem(int64(1))
	item.expiredAt = &expiredAt

	assert.Equal(t, item.Expired(), true)
	assert.Equal(t, item.Value(), int64(1))

	expiredAt = time.Now().Add(time.Second)
	assert.Equal(t, item.Expired(), false)
}
