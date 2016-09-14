package htcp

import "net/http"

type (
	// Handler is like http.Handler except ServeHTTP may return a status
	// code and/or error.
	//
	// If ServeHTTP writes the response header, it should return a status
	// code of 0. This signals to other handlers before it that the response
	// is already handled, and that they should not write to it also. Keep
	// in mind that writing to the response body writes the header, too.
	//
	// If ServeHTTP encounters an error, it should return the error value
	// so it can be logged by designated error-handling middleware.
	//
	// If writing a response after calling the next ServeHTTP method, the
	// returned status code SHOULD be used when writing the response.
	//
	// If handling errors after calling the next ServeHTTP method, the
	// returned error value SHOULD be logged or handled accordingly.
	//
	// Otherwise, return values should be propagated down the middleware
	// chain by returning them unchanged.
	//
	// Original code from Caddy: https://godoc.org/github.com/mholt/caddy
	Handler interface {
		ServeHTTP(http.ResponseWriter, *http.Request) (int, error)
	}

	// HandlerFunc is a convenience type like http.HandlerFunc, except
	// ServeHTTP returns a status code and an error. See Handler
	// documentation for more information.
	//
	// Original code from Caddy: https://godoc.org/github.com/mholt/caddy
	HandlerFunc func(http.ResponseWriter, *http.Request) (int, error)
)

// ServeHTTP implements the Handler interface.
//
// Original code from Caddy: https://godoc.org/github.com/mholt/caddy
func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	return f(w, r)
}

// EmptyNext is a no-op function that can be passed into
// Middleware functions so that the assignment to the
// Next field of the Handler can be tested.
//
// Used primarily for testing but needs to be exported so
// plugins can use this as a convenience.
//
// Original code from Caddy: https://godoc.org/github.com/mholt/caddy
var EmptyNext = HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) { return 0, nil })
