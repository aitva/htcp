package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var opt = struct {
	addr string
	code int
}{}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %v [ARGS] ADDR\n"+
		"    Listen for http request on ADDR and return\n"+
		"    an http status code.\n"+
		"\n"+
		"ARGS:\n"+
		"    -code    http status code to return (default: 200)\n"+
		"\n"+
		"ADDR:\n"+
		"    ADDR is the address to listen on in the form: localhost:8080\n", os.Args[0])
}

func init() {
	flag.Usage = usage
	flag.IntVar(&opt.code, "code", 200, "error code to return to clients")
}

func makeStatusHandler(code int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
	})
}

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	opt.addr = flag.Args()[0]

	err := http.ListenAndServe(opt.addr, makeStatusHandler(opt.code))
	fmt.Fprint(os.Stderr, "fail with error: %v", err)
	os.Exit(1)
}
