package main

import (
	"io"
	"log"
	"net/http"
)

// Proxy represents a ProxyServer.
type Proxy struct {
	Order       OrderType
	StatusCodes []int
	Dest        []string
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (p *Proxy) send(r *http.Request) ([]*http.Response, error) {
	responses := make([]*http.Response, len(p.Dest))

	cli := &http.Client{}
	for i, d := range p.Dest {
		dst := d + r.URL.String()

		log.Printf("sending request to: %s", dst)
		copy, err := http.NewRequest(r.Method, dst, r.Body)
		if err != nil {
			return nil, err
		}
		copy.WithContext(r.Context())
		resp, err := cli.Do(copy)
		if err != nil {
			return nil, err
		}
		log.Printf("got answer with code: %d", resp.StatusCode)
		responses[i] = resp
	}
	return responses, nil
}

func (p *Proxy) orderKO(responses []*http.Response) []*http.Response {
	valid := make([]*http.Response, 0, len(responses))
	invalid := make([]*http.Response, 0, len(responses))

outter_loop:
	for _, r := range responses {
		for _, code := range p.StatusCodes {
			if r.StatusCode == code {
				valid = append(valid, r)
				continue outter_loop
			}
		}
		invalid = append(invalid, r)
	}
	return append(invalid, valid...)
}

func (p *Proxy) orderOK(responses []*http.Response) []*http.Response {
	valid := make([]*http.Response, 0, len(responses))
	invalid := make([]*http.Response, 0, len(responses))

outter_loop:
	for _, r := range responses {
		for _, code := range p.StatusCodes {
			if r.StatusCode == code {
				valid = append(valid, r)
				continue outter_loop
			}
		}
		invalid = append(invalid, r)
	}
	return append(valid, invalid...)
}

func (p *Proxy) order(responses []*http.Response) []*http.Response {
	switch p.Order {
	case OrderCommand:
		return responses
	case OrderFirstKO:
		return p.orderKO(responses)
	case OrderFirstOK:
		return p.orderOK(responses)
	default:
		return responses
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request from %s", r.RemoteAddr)

	responses, err := p.send(r)
	if err != nil {
		answer := "sending fail: " + err.Error()
		log.Print(answer)
		w.Header().Add("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(answer + "\n"))
		return
	}

	responses = p.order(responses)
	resp := responses[0]

	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	n, err := io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("fail to answer client: %s", err)
	}
	log.Printf("anwser client: %db", n)

	for _, resp = range responses {
		resp.Body.Close()
	}
}
