package htcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

var fakeResponses = []*http.Response{
	{StatusCode: 400},
	{StatusCode: 202},
	{StatusCode: 404},
	{StatusCode: 201},
	{StatusCode: 500},
	{StatusCode: 200},
}

func loopOK(codes []int, responses []*http.Response) (rest []*http.Response) {
	for i, r := range responses {
		isOk := false
		for _, c := range codes {
			if r.StatusCode == c {
				isOk = true
			}
		}
		if !isOk {
			rest = responses[i:]
			break
		}
	}
	return
}

func loopKO(codes []int, responses []*http.Response) (rest []*http.Response) {
	for i, r := range responses {
		isKo := true
		for _, c := range codes {
			if r.StatusCode == c {
				isKo = false
			}
		}
		if !isKo {
			rest = responses[i:]
			break
		}
	}
	return
}

func TestParseOrder(t *testing.T) {
	table := []struct {
		Text  string
		Order Order
		Err   error
	}{
		{"paerpokz", OrderInvalid, ErrOrderInvalid},
		{"command", OrderCommand, nil},
		{"first-ko", OrderFirstKO, nil},
		{"first-ok", OrderFirstOK, nil},
	}

	for _, v := range table {
		o, err := ParseOrder(v.Text)
		if err != v.Err {
			t.Fatalf("expects %v got %v", v.Err, err)
		}
		if o != v.Order {
			t.Fatalf("expects %v got %v", v.Order, o)
		}
	}
}

func TestOrderCommand(t *testing.T) {
	codeOK := []int{200, 201, 202}

	cp := &CopyHandler{}
	cp.Responses = fakeResponses
	ctx := newCopyContext(context.Background(), cp)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	o := NewOrderHandler(EmptyNext, OrderCommand, codeOK)
	code, err := o.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}

	if len(fakeResponses) != len(cp.Responses) {
		t.Fatalf("expects %v got %v", len(fakeResponses), len(cp.Responses))
	}
	for i := 0; i < len(fakeResponses); i++ {
		fc := fakeResponses[i].StatusCode
		c := cp.Responses[i].StatusCode
		if fc != c {
			t.Fatalf("expects %v got %v", fc, c)
		}
	}
}

func TestOrderFirstKO(t *testing.T) {
	codeOK := []int{200, 201, 202}

	cp := &CopyHandler{}
	cp.Responses = fakeResponses
	ctx := newCopyContext(context.Background(), cp)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	o := NewOrderHandler(EmptyNext, OrderFirstKO, codeOK)
	code, err := o.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}

	respOK := loopKO(codeOK, cp.Responses)
	if len(respOK) == 0 {
		t.Fatalf("expects len(respOK) > 0 got %v", len(respOK))
	}
	rest := loopOK(codeOK, respOK)
	if len(rest) != 0 {
		t.Fatalf("expects len(rest) == 0 got %v", len(rest))
	}
}

func TestOrderFirstOK(t *testing.T) {
	codeOK := []int{200, 201, 202}

	cp := &CopyHandler{}
	cp.Responses = fakeResponses
	ctx := newCopyContext(context.Background(), cp)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	o := NewOrderHandler(EmptyNext, OrderFirstOK, codeOK)
	code, err := o.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}

	respKO := loopOK(codeOK, cp.Responses)
	if len(respKO) == 0 {
		t.Fatalf("expects len(respKO) > 0 got %v", len(respKO))
	}
	rest := loopKO(codeOK, respKO)
	if len(rest) != 0 {
		t.Fatalf("expects len(rest) == 0 got %v", len(rest))
	}
}
