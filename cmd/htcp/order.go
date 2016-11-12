package main

import (
	"errors"
	"net/http"
)

const (
	// OrderInvalid represent an invalid or uninitialized OrderType
	OrderInvalid Order = iota
	// OrderCommand demands command line ordering
	OrderCommand
	// OrderFirstOK demands status code ordering starting with valid status
	OrderFirstOK
	// OrderFirstKO demands status code ordering starting with invalid status
	OrderFirstKO
)

// ErrOrderInvalid is return when trying to create an order type from an invalid string
var ErrOrderInvalid = errors.New("invalid Order string")

// Order describes all possible ordering for http answer.
type Order int

// ParseOrder create a new Order from a string.
func ParseOrder(str string) (Order, error) {
	valid := map[string]Order{
		"command":  OrderCommand,
		"first-ko": OrderFirstKO,
		"first-ok": OrderFirstOK,
	}
	t, ok := valid[str]
	if !ok {
		return OrderInvalid, ErrOrderInvalid
	}
	return t, nil
}

// Set is used by flag.Var to parse user defined type.
func (o *Order) Set(value string) error {
	tmp, err := ParseOrder(value)
	if err != nil {
		return err
	}
	*o = tmp
	return nil
}

func (o Order) String() string {
	str := "invalid"
	switch o {
	case OrderCommand:
		str = "command"
	case OrderFirstKO:
		str = "first-ok"
	case OrderFirstOK:
		str = "first-ko"
	}
	return str
}

// Sort orders http.Response using the Order type.
func (o Order) Sort(codeOK []int, responses []*http.Response) []*http.Response {
	switch o {
	case OrderFirstOK:
		valid, invalid := sortByCode(codeOK, responses)
		return append(valid, invalid...)
	case OrderFirstKO:
		valid, invalid := sortByCode(codeOK, responses)
		return append(invalid, valid...)
	default:
		break
	}
	return responses
}

// sortByCode order a set of http.Response using their StatusCode.
func sortByCode(codeOK []int, responses []*http.Response) (valid, invalid []*http.Response) {
	valid = make([]*http.Response, 0, len(responses))
	invalid = make([]*http.Response, 0, len(responses))

outter_loop:
	for _, r := range responses {
		for _, code := range codeOK {
			if r.StatusCode == code {
				valid = append(valid, r)
				continue outter_loop
			}
		}
		invalid = append(invalid, r)
	}
	return valid, invalid
}
