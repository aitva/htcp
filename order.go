package htcp

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

// OrderHandler is an HTTP middleware handling response ordering.
type OrderHandler struct {
	next   Handler
	Order  Order
	CodeOK []int
}

// NewOrderHandler instantiate an OrderHandler.
func NewOrderHandler(next Handler, o Order, codeOK []int) *OrderHandler {
	return &OrderHandler{
		next:   next,
		Order:  o,
		CodeOK: codeOK,
	}
}

func (o *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	copy, ok := FromCopyContext(r.Context())
	if !ok {
		return 500, errors.New("fail to retrieve CopyHandler from context")
	}
	switch o.Order {
	case OrderCommand:
		// Nothing to do.
	case OrderFirstOK:
		copy.Responses = o.OrderFirstOK(copy.Responses)
	case OrderFirstKO:
		copy.Responses = o.OrderFirstKO(copy.Responses)
	default:
		return 500, errors.New("invalid order request")
	}
	return o.next.ServeHTTP(w, r)
}

func (o *OrderHandler) sortByCode(responses []*http.Response) (valid, invalid []*http.Response) {
	valid = make([]*http.Response, 0, len(responses))
	invalid = make([]*http.Response, 0, len(responses))

outter_loop:
	for _, r := range responses {
		for _, code := range o.CodeOK {
			if r.StatusCode == code {
				valid = append(valid, r)
				continue outter_loop
			}
		}
		invalid = append(invalid, r)
	}
	return valid, invalid
}

// OrderFirstKO order http.Response by status code. It puts the requests
// with a code outside OrderHandler.CodeOK first in the returned slice.
func (o *OrderHandler) OrderFirstKO(responses []*http.Response) []*http.Response {
	valid, invalid := o.sortByCode(responses)
	return append(invalid, valid...)
}

// OrderFirstOK order http.Response by status code. It puts the requests
// with a code inside OrderHandler.CodeOK first in the returned slice.
func (o *OrderHandler) OrderFirstOK(responses []*http.Response) []*http.Response {
	valid, invalid := o.sortByCode(responses)
	return append(valid, invalid...)
}
