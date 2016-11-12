package main

import (
	"net/http"
	"testing"
)

var testResp = []*http.Response{
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

	o := OrderCommand
	ordered := o.Sort(codeOK, testResp)

	if len(testResp) != len(ordered) {
		t.Fatalf("expects %v got %v", len(testResp), len(ordered))
	}
	for i := 0; i < len(testResp); i++ {
		fc := testResp[i].StatusCode
		c := ordered[i].StatusCode
		if fc != c {
			t.Fatalf("expects %v got %v", fc, c)
		}
	}
}

func TestOrderFirstOK(t *testing.T) {
	codeOK := []int{200, 201, 202}

	o := OrderFirstOK
	ordered := o.Sort(codeOK, testResp)

	if len(testResp) != len(ordered) {
		t.Fatalf("expects %v got %v", len(testResp), len(ordered))
	}

	respKO := loopOK(codeOK, ordered)
	if len(respKO) == 0 {
		t.Fatalf("expects len(respKO) > 0 got %v", len(respKO))
	}
	rest := loopKO(codeOK, respKO)
	if len(rest) != 0 {
		t.Fatalf("expects len(rest) == 0 got %v", len(rest))
	}
}

func TestOrderFirstKO(t *testing.T) {
	codeOK := []int{200, 201, 202}

	o := OrderFirstKO
	ordered := o.Sort(codeOK, testResp)

	if len(testResp) != len(ordered) {
		t.Fatalf("expects %v got %v", len(testResp), len(ordered))
	}

	respOK := loopKO(codeOK, ordered)
	if len(respOK) == 0 {
		t.Fatalf("expects len(respOK) > 0 got %v", len(respOK))
	}
	rest := loopOK(codeOK, respOK)
	if len(rest) != 0 {
		t.Fatalf("expects len(rest) == 0 got %v", len(rest))
	}
}
