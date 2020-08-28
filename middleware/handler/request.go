package handler

import (
	"encoding/json"
	"errors"
)

var (
	errMissingFilterKey  = errors.New("missing filter key")
	errInvalidFilterType = errors.New("invalid filter type")
)

// DisplayAble used to determine whether a request contains display sort
type DisplayAble interface {
	GetOrder() string
	SetOrder(string)
}

// Display represent a request supports display filtering and sorting
type Display struct {
	Filter map[string]interface{} `json:"filter"`
	Sorter Sorter                 `json:"sort"`
}

// Sorter represent a request supports display sorting
type Sorter struct {
	By    string `json:"by"`
	Order string `json:"order"`
}

// GetOrder return sort's order
func (d *Display) GetOrder() string {
	return d.Sorter.Order
}

// SetOrder set sort's order
func (d *Display) SetOrder(order string) {
	d.Sorter.Order = order
}

// GetFilterString give the filter keyword return the string value
func (d *Display) GetFilterString(filterKey string) (string, error) {
	if _, ok := d.Filter[filterKey]; !ok {
		return "", errMissingFilterKey
	}
	if val, ok := d.Filter[filterKey].(string); ok {
		return val, nil
	}
	return "", errInvalidFilterType
}

// GetFilterNum give the filter keyword return the numeric value
func (d *Display) GetFilterNum(filterKey string) (interface{}, error) {
	if _, ok := d.Filter[filterKey]; !ok {
		return 0, errMissingFilterKey
	}
	switch val := d.Filter[filterKey].(type) {
	case int, int16, int32, int64, int8, uint, uint16, uint32, uint64, uint8, float32, float64:
		return val, nil
	}

	return 0, errInvalidFilterType
}

// GetFilterBoolean give the filter keyword return the boolean value
func (d *Display) GetFilterBoolean(filterKey string) (bool, error) {
	if _, ok := d.Filter[filterKey]; !ok {
		return false, errMissingFilterKey
	}
	if val, ok := d.Filter[filterKey].(bool); ok {
		return val, nil
	}
	return false, errInvalidFilterType
}

// GetFilterObject give the filter keyword return the object value
func (d *Display) GetFilterObject(filterKey string, obj interface{}) error {
	if _, ok := d.Filter[filterKey]; !ok {
		return errMissingFilterKey
	}

	bytes, err := json.Marshal(d.Filter[filterKey])
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, obj)
}
