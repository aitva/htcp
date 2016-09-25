package htcp

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"
)

type copyContextKey int

var myCopyContextKey = 0

func copyResponse(w http.ResponseWriter, r *http.Response) (int64, error) {
	h := w.Header()
	// Copy Headers.
	for k, vv := range r.Header {
		for _, v := range vv {
			h.Add(k, v)
		}
	}
	w.WriteHeader(r.StatusCode)
	return io.Copy(w, r.Body)
}

// cReader is a concurent copy reader, every byte read from the reader is
// duplicated to all the copies. This is a io.TeeReader with multiple writers
// and a sync.Mutex.
type cReader struct {
	sync.RWMutex
	r io.ReadCloser
	w []io.Writer
}

func (c *cReader) Read(p []byte) (n int, err error) {
	n, err = c.r.Read(p)
	if n <= 0 {
		return
	}

	c.Lock()
	defer c.Unlock()

	for _, w := range c.w {
		n, err = w.Write(p[:n])
		if err != nil {
			return n, err
		}
	}
	return
}

func (c *cReader) Close() error {
	return c.r.Close()
}

// Copy creates a new ReadCloser synchronized with cReader.
func (c *cReader) Copy() io.ReadCloser {
	buf := &cBufReader{
		Locker: c.RWMutex.RLocker(),
	}
	c.w = append(c.w, buf)
	return buf
}

// cBufReader is simply a buffer with a lock.
type cBufReader struct {
	sync.Locker
	bytes.Buffer
}

func (c *cBufReader) Read(p []byte) (n int, err error) {
	c.Lock()
	n, err = c.Buffer.Read(p)
	c.Unlock()
	return
}

func (c *cBufReader) Close() error {
	return nil
}

func makeReadCloserSlice(body io.ReadCloser, n int) []io.ReadCloser {
	readers := make([]io.ReadCloser, n)
	if body == nil {
		return readers
	}
	cr := &cReader{r: body}
	readers[0] = cr
	for i := 1; i < len(readers); i++ {
		readers[i] = cr.Copy()
	}
	return readers
}

// CopyHandler is an HTTP midleware handling request duplication.
type CopyHandler struct {
	handler   Handler
	Servers   []string
	Responses []*http.Response
}

// NewCopyHandler creates a new CopyHandler.
func NewCopyHandler(handler Handler, servers []string) *CopyHandler {
	return &CopyHandler{
		handler: handler,
		Servers: servers,
	}
}

// NewCopyContext returns a new Context that carries a CopyHandler.
func NewCopyContext(ctx context.Context, c *CopyHandler) context.Context {
	return context.WithValue(ctx, myCopyContextKey, c)
}

// FromCopyContext returns the CopyHandler stored in ctx, if any.
func FromCopyContext(ctx context.Context) (*CopyHandler, bool) {
	c, ok := ctx.Value(myCopyContextKey).(*CopyHandler)
	return c, ok
}

// SendCopy duplicate a request to the servers.
func (c *CopyHandler) SendCopy(r *http.Request) error {
	responses := make([]*http.Response, len(c.Servers))
	readers := makeReadCloserSlice(r.Body, len(c.Servers))
	cli := &http.Client{}
	// Remove Accept-Encoding from http.Request.
	// TODO: add timeout.
	cli.Transport = &http.Transport{
		DisableCompression: true,
	}
	for i, d := range c.Servers {
		dst := d + r.URL.String()
		copy, err := http.NewRequest(r.Method, dst, readers[i])
		if err != nil {
			return err
		}
		// Remove Go's 'User-Agent' from http.Request.
		r.Header.Set("User-Agent", "")
		for k, v := range r.Header {
			for i := 0; i < len(v); i++ {
				copy.Header.Set(k, v[i])
			}
		}

		copy = copy.WithContext(r.Context())
		resp, err := cli.Do(copy)
		if err != nil {
			return err
		}
		responses[i] = resp
	}

	c.Responses = responses
	return nil
}

func (c *CopyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	err := c.SendCopy(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ctx := NewCopyContext(r.Context(), c)
	code, err := c.handler.ServeHTTP(w, r.WithContext(ctx))

	for _, resp := range c.Responses {
		resp.Body.Close()
	}
	return code, err
}
