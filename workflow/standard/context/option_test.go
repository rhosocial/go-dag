package context

import (
	"errors"
	"testing"

	"github.com/rhosocial/go-dag/workflow/standard/cache"
	"github.com/stretchr/testify/assert"
)

func WithGlobalOptions() OptionsItem {
	return func(o *Options) error {
		return o.setGlobal("abc", "123")
	}
}

func WithGlobalOptionsError() OptionsItem {
	return func(o *Options) error {
		return errors.New("abc")
	}
}

func TestNewOptions(t *testing.T) {
	options, err := NewOptions(WithGlobalOptions())
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, options)
	global, err := options.GetGlobal("abc")
	if err != nil {
		t.Fail()
		return
	}
	result := global.(string)
	assert.Equal(t, "123", result)

	options, err = NewOptions(WithGlobalOptionsError())
	assert.Equal(t, "abc", err.Error())
}

func TestNewOptionsSetGlobalEmpty(t *testing.T) {
	options, err := NewOptions()
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, options)
	err = options.setGlobal("abc", "123")
	assert.NoError(t, err)
}

func WithTransitsOptions() OptionsItem {
	return func(o *Options) error {
		return o.setTransit("t1", "abc", "123")
	}
}

func WithTransitsOptionsError() OptionsItem {
	return func(o *Options) error {
		return errors.New("abc")
	}
}

func TestNewOptionsWithTransits(t *testing.T) {
	options, err := NewOptions(WithTransitsOptions())
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, options)

	transit, err := options.GetTransit("t1", "abc")
	if err != nil {
		t.Fail()
		return
	}
	result := transit.(string)
	assert.Equal(t, "123", result)

	transit, err = options.GetTransit("t2", "abc")
	assert.ErrorAs(t, err, &cache.ErrKeyNotFound{})

	options, err = NewOptions(WithTransitsOptionsError())
	assert.Equal(t, "abc", err.Error())
}

func TestNewOptionsWithEmpty(t *testing.T) {
	options, err := NewOptions()
	assert.NoError(t, err)
	assert.NotNil(t, options)
	transit, err := options.GetTransit("t1", "abc")
	assert.ErrorAs(t, err, &cache.ErrKeyNotFound{})
	assert.Nil(t, transit)
}
