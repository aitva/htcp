package htcp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testCopyMsgA = "Hello Client From A!"
const testCopyMsgB = "Hello Client From B!"

func makeTestCopyServerHandler(msg []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(msg)
	})
}

func makeTestCopyHandler(t *testing.T) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
		c, ok := FromCopyContext(r.Context())
		if !ok {
			t.Fatalf("CopyHandler is not present in request.Context")
		}
		if len(c.Responses) != 2 {
			t.Fatalf("unexpected number of responses: %d", len(c.Responses))
		}

		for _, r := range c.Responses {
			if r.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: %d", r.StatusCode)
			}
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("fail to read server answer: %v", err)
			}
			if bytes.Equal(buf, []byte(testCopyMsgA)) {
				t.Fatalf("unexpected answer from server: %s", string(buf))
			}
		}
		return 0, nil
	})
}

func TestCopy(t *testing.T) {
	tsa := httptest.NewServer(makeTestCopyServerHandler([]byte(testCopyMsgA)))
	defer tsa.Close()
	tsb := httptest.NewServer(makeTestCopyServerHandler([]byte(testCopyMsgB)))
	defer tsb.Close()

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	rec := httptest.NewRecorder()

	c := NewCopyHandler(EmptyNext, []string{tsa.URL, tsb.URL})
	code, err := c.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}
}

// TODO: TestCopyHeader

// TODO: TestCopyQuery
