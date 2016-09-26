package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aitva/htcp"
)

var version = "0.01.00"

var flags = struct {
	verbose bool
	version bool
	help    bool
	listen  string
	order   string
	expects string
}{}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] server1.org server2.com [...]\n\n"+
		"    A command to duplicate HTTP request to multiple server.\n"+
		"    \n"+
		"    The response returned is from the first server in the command line.\n"+
		"    But another response can be selected using the -order flag.\n"+
		"\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.Usage = usage
	flag.BoolVar(&flags.verbose, "verbose", false, "print verbose output on stdout")
	flag.BoolVar(&flags.version, "version", false, "display command version")
	flag.BoolVar(&flags.help, "help", false, "display the usage")
	flag.StringVar(&flags.order, "order", "command", "order server response and return the first one.\n"+
		"        Valid values are:\n"+
		"            command    first server in the command\n"+
		"            first-ko   first response with unexpect status code\n"+
		"            first-ko   first response with expected status code\n"+
		"       ")
	flag.StringVar(&flags.expects, "expect", "200 201 202 203 204", "valid http response code")
	flag.StringVar(&flags.listen, "listen", "localhost:8080", "address to listen on")
}

func parseStatusCodes(str string) ([]int, error) {
	var codes []int
	words := strings.Split(str, " ")
	for _, w := range words {
		i, err := strconv.ParseUint(w, 10, 32)
		if err != nil {
			return nil, err
		}
		codes = append(codes, int(i))
	}
	return codes, nil
}

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

func mainHandler(w http.ResponseWriter, r *http.Request) (int, error) {
	copyHandler, ok := htcp.FromCopyContext(r.Context())
	if !ok {
		return 500, errors.New("fail to retrieve CopyHandler from context")
	}
	resp := copyHandler.Responses[0]
	_, err := copyResponse(w, resp)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func makeLogHandler(h htcp.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		code, err := h.ServeHTTP(w, r)
		took := time.Since(t0)
		if code == 0 && err == nil {
			log.Printf("handle in %v", took)
			return
		}
		if code == 0 && err != nil {
			log.Printf("fail in %v with error: %v", took, err)
			return
		}
		log.Printf("fail in %v with code %d and error: %v", took, code, err)
		w.WriteHeader(code)
		w.Write([]byte(err.Error()))
		w.Write([]byte("\n"))
	})
}

func main() {
	flag.Parse()
	if flags.help {
		usage()
	}
	if flags.version {
		fmt.Fprintf(os.Stdout, "v%s\n", version)
		os.Exit(0)
	}
	_, err := parseStatusCodes(flags.expects)
	if err != nil {
		log.Fatal("expect: invalid value")
	}
	if !flags.verbose {
		log.SetOutput(ioutil.Discard)
	}
	if flag.NArg() < 1 {
		usage()
	}

	servers := flag.Args()
	for i := range servers {
		if !strings.Contains(servers[i], "http://") {
			servers[i] = "http://" + servers[i]
		}
	}

	copyHandler := htcp.NewCopyHandler(htcp.HandlerFunc(mainHandler), servers)
	logHandler := makeLogHandler(copyHandler)
	http.Handle("/", logHandler)
	log.Printf("server is listening on %s", flags.listen)
	log.Fatal(http.ListenAndServe(flags.listen, nil))
}
