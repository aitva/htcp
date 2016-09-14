package htcp

import (
	"context"
	"io"
	"net/http"
)

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

type copyContextKey int

var myCopyContextKey = 0

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

	cli := &http.Client{}
	for i, d := range c.Servers {
		dst := d + r.URL.String()

		copy, err := http.NewRequest(r.Method, dst, r.Body)
		if err != nil {
			return err
		}
		copy.WithContext(r.Context())
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
	r.WithContext(ctx)
	code, err := c.handler.ServeHTTP(w, r)

	for _, resp := range c.Responses {
		resp.Body.Close()
	}
	return code, err
}
