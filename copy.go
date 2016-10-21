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

// MultiReader is a concurent TeeReader, wich can duplicate
// reads to as many reader as needed.
type MultiReader struct {
	sync.RWMutex
	r io.ReadCloser
	w []io.Writer
}

// NewMultiReader instanciate a new MultiReader.
func NewMultiReader(r io.ReadCloser) *MultiReader {
	return &MultiReader{r: r}
}

func (m *MultiReader) Read(p []byte) (n int, err error) {
	n, err = m.r.Read(p)
	if n <= 0 {
		return
	}

	m.Lock()
	defer m.Unlock()

	for _, w := range m.w {
		n, err = w.Write(p[:n])
		if err != nil {
			return n, err
		}
	}
	return
}

// Close close the underlying ReadCloser.
func (m *MultiReader) Close() error {
	return m.r.Close()
}

// Copy creates a new ReadCloser synchronized with a MultiReader.
func (m *MultiReader) Copy() io.ReadCloser {
	buf := &bufReader{
		Locker: m.RWMutex.RLocker(),
	}
	m.w = append(m.w, buf)
	return buf
}

// cBufReader is simply a buffer with a lock.
type bufReader struct {
	sync.Locker
	bytes.Buffer
}

func (b *bufReader) Read(p []byte) (n int, err error) {
	b.Lock()
	n, err = b.Buffer.Read(p)
	b.Unlock()
	return
}

func (b *bufReader) Close() error {
	return nil
}

func makeReadCloserSlice(body io.ReadCloser, n int) []io.ReadCloser {
	readers := make([]io.ReadCloser, n)
	if body == nil {
		return readers
	}
	cr := NewMultiReader(body)
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
