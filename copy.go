package htcp

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type copyContextKey int

var myCopyContextKey = 0

// newCopyContext returns a new Context that carries a CopyHandler.
func newCopyContext(ctx context.Context, c *CopyHandler) context.Context {
	return context.WithValue(ctx, myCopyContextKey, c)
}

// FromCopyContext returns the CopyHandler stored in ctx, if any.
func FromCopyContext(ctx context.Context) (*CopyHandler, bool) {
	c, ok := ctx.Value(myCopyContextKey).(*CopyHandler)
	return c, ok
}

func makeBufferSlices(n int) (readers []*bytes.Buffer, writers []io.Writer) {
	buffers := make([]bytes.Buffer, n)
	readers = make([]*bytes.Buffer, n)
	writers = make([]io.Writer, n)
	for i := range buffers {
		readers[i] = &buffers[i]
		writers[i] = &buffers[i]
	}
	return readers, writers
}

// CopyHandler is an HTTP midleware handling request duplication.
type CopyHandler struct {
	client    *http.Client
	handler   Handler
	Servers   []string
	Responses []*http.Response
}

// NewCopyHandler creates a new CopyHandler.
func NewCopyHandler(handler Handler, servers []string) *CopyHandler {
	// Declaring a client here enable connection reuse.
	cli := &http.Client{}
	// Remove Accept-Encoding from http.Request.
	// TODO: add timeout.
	cli.Transport = &http.Transport{
		DisableCompression: true,
	}
	return &CopyHandler{
		client:  cli,
		handler: handler,
		Servers: servers,
	}
}

// SendCopies duplicate a request to the servers.
func (c *CopyHandler) SendCopies(r *http.Request) error {
	var g errgroup.Group
	responses := make([]*http.Response, len(c.Servers))
	readers, writers := makeBufferSlices(len(c.Servers))

	if r.Body != nil {
		io.Copy(io.MultiWriter(writers...), r.Body)
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

		rq := copy.WithContext(r.Context())
		i := i
		g.Go(func() error {
			resp, err := c.client.Do(rq)
			if err == nil {
				responses[i] = resp
			}
			return err
		})
	}

	err := g.Wait()
	c.Responses = responses
	return err
}

func (c *CopyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	err := c.SendCopies(r)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	ctx := newCopyContext(r.Context(), c)
	code, err := c.handler.ServeHTTP(w, r.WithContext(ctx))

	for _, resp := range c.Responses {
		if resp != nil {
			resp.Body.Close()
		}
	}
	return code, err
}
