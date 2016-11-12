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

var args = struct {
	verbose bool
	version bool
	help    bool
	listen  string
	order   Order
	expects string
}{}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [flags] server1.org server2.com [...]\n\n"+
		"    A command to duplicate HTTP request.\n"+
		"    \n"+
		"    The response returned is from the first server in the command line,\n"+
		"    but responses can be ordered using the -order flag.\n"+
		"\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
	os.Exit(1)
}

func init() {
	flag.Usage = usage
	flag.BoolVar(&args.verbose, "verbose", false, "print verbose output on stdout")
	flag.BoolVar(&args.version, "version", false, "display command version")
	flag.BoolVar(&args.help, "help", false, "display the usage")
	flag.Var(&args.order, "order", "order responses using one of the following filter:\n"+
		"            command    first server in the command line call\n"+
		"            first-ko   first response with unexpect status code\n"+
		"            first-ko   first response with expected status code\n")
	flag.StringVar(&args.expects, "expects", "200 201 202 203 204", "valid http response code")
	flag.StringVar(&args.listen, "listen", "localhost:8080", "address to listen on")
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

func makeMainHandler(codeOK []int, o Order) htcp.HandlerFunc {
	return htcp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) (int, error) {
		copy, ok := htcp.FromCopyContext(r.Context())
		if !ok {
			return 500, errors.New("fail to retrieve CopyHandler from context")
		}
		resp := o.Sort(codeOK, copy.Responses)
		_, err := copyResponse(w, resp[0])
		if err != nil {
			return 0, err
		}
		return 0, nil
	})
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
	if args.help {
		usage()
	}
	if args.version {
		fmt.Fprintf(os.Stdout, "htcp version %s\n", version)
		os.Exit(0)
	}
	codeOK, err := parseStatusCodes(args.expects)
	if err != nil {
		log.Fatal("expect: invalid value")
	}
	if !args.verbose {
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

	mainHandler := makeMainHandler(codeOK, args.order)
	copyHandler := htcp.NewCopyHandler(mainHandler, servers)
	logHandler := makeLogHandler(copyHandler)
	http.Handle("/", logHandler)
	log.Printf("server is listening on %s", args.listen)
	log.Fatal(http.ListenAndServe(args.listen, nil))
}
