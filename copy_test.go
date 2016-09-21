package htcp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var serverMessages = []string{
	"Hello Client From A!",
	"Hello Client From B!",
}

func makeCopyServerHandler(msg []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(msg)
	})
}

func makeCopyHandler(t *testing.T) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
		c, ok := FromCopyContext(r.Context())
		if !ok {
			t.Fatalf("CopyHandler is not present in request.Context")
		}
		if len(c.Responses) != 2 {
			t.Fatalf("unexpected number of responses: %d", len(c.Responses))
		}

		for i, r := range c.Responses {
			if r.StatusCode != http.StatusOK {
				t.Fatalf("unexpected status code: %d", r.StatusCode)
			}
			buf, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("fail to read server answer: %v", err)
			}
			if !bytes.Equal(buf, []byte(serverMessages[i])) {
				t.Fatalf("unexpected answer from server: %s", string(buf))
			}
		}
		return 0, nil
	})
}

func TestCopy(t *testing.T) {
	servers := make([]*httptest.Server, len(serverMessages))
	urls := make([]string, len(serverMessages))
	for i, msg := range serverMessages {
		servers[i] = httptest.NewServer(makeCopyServerHandler([]byte(msg)))
		defer servers[i].Close()
		urls[i] = servers[i].URL
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	rec := httptest.NewRecorder()

	c := NewCopyHandler(makeCopyHandler(t), urls)
	code, err := c.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}
}

func makeCopyHeaderHandler(msg []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: check received headers.
		w.Write(msg)
	})
}

func TestCopy_Header(t *testing.T) {
	servers := make([]*httptest.Server, len(serverMessages))
	urls := make([]string, len(serverMessages))
	for i, msg := range serverMessages {
		servers[i] = httptest.NewServer(makeCopyServerHandler([]byte(msg)))
		defer servers[i].Close()
		urls[i] = servers[i].URL
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	rec := httptest.NewRecorder()

	c := NewCopyHandler(EmptyNext, urls)
	code, err := c.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}
}

// TODO: TestCopyQuery
