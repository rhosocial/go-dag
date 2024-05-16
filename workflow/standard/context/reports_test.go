package context

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReportsSetGlobal(t *testing.T) {
	reports, err := NewReports()
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, reports)
	err = reports.AddGlobal("abc", "123")
	assert.NoError(t, err)
}

func TestReports(t *testing.T) {
	// Test NewReports with no options
	reports, err := NewReports()
	assert.NoError(t, err, "NewReports should not return an error without options")
	assert.NotNil(t, reports, "NewReports should return a non-nil Reports instance")

	// Test AddGlobal and GetGlobal
	err = reports.AddGlobal("globalKey", "globalValue")
	assert.NoError(t, err, "AddGlobal should not return an error")

	value, err := reports.GetGlobal("globalKey")
	assert.NoError(t, err, "GetGlobal should not return an error")
	assert.Equal(t, "globalValue", value, "GetGlobal should return the correct value")

	// Test GetGlobal with a non-existent key
	_, err = reports.GetGlobal("nonExistentKey")
	assert.Error(t, err, "GetGlobal should return an error for a non-existent key")

	// Test AddTransit and GetTransit
	err = reports.AddTransit("transit1", "transitKey1", "transitValue1")
	assert.NoError(t, err, "AddTransit should not return an error")

	value, err = reports.GetTransit("transit1", "transitKey1")
	assert.NoError(t, err, "GetTransit should not return an error")
	assert.Equal(t, "transitValue1", value, "GetTransit should return the correct value")

	// Test GetTransit with a non-existent transit key
	_, err = reports.GetTransit("transit1", "nonExistentTransitKey")
	assert.Error(t, err, "GetTransit should return an error for a non-existent key")

	// Test GetTransit with a non-existent transit
	_, err = reports.GetTransit("nonExistentTransit", "someKey")
	assert.Error(t, err, "GetTransit should return an error for a non-existent transit")
}

func TestNewReportsWithOptions(t *testing.T) {
	optionErr := errors.New("option error")
	option := func(o *Reports) error {
		return optionErr
	}

	// Test NewReports with an option that returns an error
	r, err := NewReports(option)
	assert.Error(t, err, "NewReports should return an error when an option returns an error")
	assert.Equal(t, optionErr, err, "NewReports should return the correct error")
	assert.Nil(t, r, "NewReports should return a nil Reports instance when an option returns an error")
}

func TestNewReportsGetTransitEmpty(t *testing.T) {
	reports, err := NewReports()
	if err != nil {
		t.Fail()
		return
	}
	assert.NotNil(t, reports)
	result, err := reports.GetTransit("transit", "abc")
	assert.Nil(t, result)
	assert.Error(t, err)
}
