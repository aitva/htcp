package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func init() {
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

func main() {
	flag.Parse()
	if flags.help {
		usage()
	}
	if flags.version {
		fmt.Fprintf(os.Stdout, "v%s\n", version)
		os.Exit(0)
	}
	order, err := NewOrderType(flags.order)
	if err != nil {
		log.Fatal("invalid -order value")
	}
	expects, err := parseStatusCodes(flags.expects)
	if err != nil {
		log.Fatal("invalid -expect value")
	}
	if !flags.verbose {
		log.SetOutput(ioutil.Discard)
	}
	p := &Proxy{
		Dest:        flag.Args(),
		Order:       order,
		StatusCodes: expects,
	}
	if len(p.Dest) < 2 {
		usage()
	}
	for i, srv := range p.Dest {
		if !strings.Contains(srv, "http://") {
			p.Dest[i] = "http://" + srv
		}
	}

	http.Handle("/", p)
	log.Printf("server is listening on %s", flags.listen)
	log.Fatal(http.ListenAndServe(flags.listen, nil))
}
