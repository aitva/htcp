package htcp

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
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

var clientHeaders = map[string][]string{
	"Content-Type":  {"text/plain"},
	"Cache-Control": {"no-cache"},
	"Cookie":        {"status=testing"},
}

func makeCopyHeaderHandler(t *testing.T, msg []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for n, recv := range r.Header {
			sent, ok := clientHeaders[n]
			if !ok {
				t.Fatal("unexpected header:", n)
				continue
			}
			if len(recv) != len(sent) {
				t.Fatalf("unexpected header content: %v != %v", sent, recv)
			}
			for i := range recv {
				if sent[i] != recv[i] {
					t.Fatalf("unexpected header content: %v != %v", sent, recv)
				}
			}
		}
		w.Write(msg)
	})
}

func TestCopy_Header(t *testing.T) {
	servers := make([]*httptest.Server, len(serverMessages))
	urls := make([]string, len(serverMessages))
	for i, msg := range serverMessages {
		servers[i] = httptest.NewServer(makeCopyHeaderHandler(t, []byte(msg)))
		defer servers[i].Close()
		urls[i] = servers[i].URL
	}

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	for n, vv := range clientHeaders {
		for _, v := range vv {
			req.Header.Set(n, v)
		}
	}
	rec := httptest.NewRecorder()

	c := NewCopyHandler(EmptyNext, urls)
	code, err := c.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}
}

const boundary = "boundaryboundaryboundary"

var clientData = []byte("--" + boundary + "\r\n" +
	"Content-Disposition: form-data; name=\"artiste\"\r\n" +
	"Content-Type: text/plain\r\n" +
	"\r\n" +
	"The B-52's" +
	"\r\n--" + boundary + "\r\n" +
	"Content-Disposition: form-data; name=\"friends\"\r\n" +
	"Content-Type: application/json\r\n" +
	"\r\n" +
	`{
    "artiste": "The B-52's",
    "albums": [
        {
            "name": "Cosmic Things",
            "date": 1989
        },
        {
            "name": "B-52's",
            "date": 1979
        },
    ]
}` +
	"\r\n--" + boundary + "--")

func makeCopyBodyHandler(t *testing.T, msg []byte) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Fatal("fail to read client request:", err)
		}
		if bytes.Compare(data, clientData) != 0 {
			t.Logf("got %q", data)
			t.Logf("expect %q", clientData)
			t.Fatal("unexpected data from client")
		}
		w.Write(msg)
	})
}

func TestCopy_Body(t *testing.T) {
	servers := make([]*httptest.Server, len(serverMessages))
	urls := make([]string, len(serverMessages))
	for i, msg := range serverMessages {
		servers[i] = httptest.NewServer(makeCopyBodyHandler(t, []byte(msg)))
		defer servers[i].Close()
		urls[i] = servers[i].URL
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(clientData))
	if err != nil {
		t.Fatalf("could not create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "mutlipart/form-data")
	req.Header.Set("Content-Type", "boundary="+boundary)
	req.Header.Set("Content-Length", strconv.Itoa(len(clientData)))
	rec := httptest.NewRecorder()

	c := NewCopyHandler(EmptyNext, urls)
	code, err := c.ServeHTTP(rec, req)
	if code != 0 || err != nil {
		t.Fatalf("ServeHTTP returned code %d and error: %v", code, err)
	}
}
