package htcp

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

// func (c *Copy) orderKO(responses []*http.Response) []*http.Response {
// 	valid := make([]*http.Response, 0, len(responses))
// 	invalid := make([]*http.Response, 0, len(responses))
//
// outter_loop:
// 	for _, r := range responses {
// 		for _, code := range c.StatusCodes {
// 			if r.StatusCode == code {
// 				valid = append(valid, r)
// 				continue outter_loop
// 			}
// 		}
// 		invalid = append(invalid, r)
// 	}
// 	return append(invalid, valid...)
// }
//
// func (c *Copy) orderOK(responses []*http.Response) []*http.Response {
// 	valid := make([]*http.Response, 0, len(responses))
// 	invalid := make([]*http.Response, 0, len(responses))
//
// outter_loop:
// 	for _, r := range responses {
// 		for _, code := range c.StatusCodes {
// 			if r.StatusCode == code {
// 				valid = append(valid, r)
// 				continue outter_loop
// 			}
// 		}
// 		invalid = append(invalid, r)
// 	}
// 	return append(valid, invalid...)
// }
//
// func (c *Copy) order(responses []*http.Response) []*http.Response {
// 	switch c.Order {
// 	case OrderCommand:
// 		return responses
// 	case OrderFirstKO:
// 		return c.orderKO(responses)
// 	case OrderFirstOK:
// 		return c.orderOK(responses)
// 	default:
// 		return responses
// 	}
// }
