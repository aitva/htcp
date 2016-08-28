package main

import "errors"

const (
	// OrderInvalid represent an invalid or uninitialized OrderType
	OrderInvalid OrderType = iota
	// OrderCommand demands command line ordering
	OrderCommand
	// OrderFirstOK demands status code ordering starting with valid status
	OrderFirstOK
	// OrderFirstKO demands status code ordering starting with invalid status
	OrderFirstKO
)

// ErrOrderTypeInvalid is return when trying to create an order type from an invalid string
var ErrOrderTypeInvalid = errors.New("invalid OrderType")

// OrderType describes all possible ordering for http answer.
type OrderType int

// NewOrderType create a new OrderType from a string.
func NewOrderType(str string) (OrderType, error) {
	valid := map[string]OrderType{
		"command":  OrderCommand,
		"first-ko": OrderFirstOK,
		"first-ok": OrderFirstOK,
	}
	t, ok := valid[str]
	if !ok {
		return OrderInvalid, ErrOrderTypeInvalid
	}
	return t, nil
}
