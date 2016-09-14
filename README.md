# htcp
Htcp is a command line utility to duplicate HTTP traffic. It can duplicate an HTTP request
and send it to as many servers as you need. Filters can be apply to select which answer
must be returned to the client.

This application is cross platform and works on Windows, OSX and Linux.
The usage is quite simple:

```bash
# htcp start listening on port 8080 and duplicate
# all the incoming traffic to golang.org and example.com
$ htcp --listen localhost:8080 golang.org example.com
```

## Status
Htcp is under heavy development and must not be use in production.

## Goal
Htcp is meant to be simple and production ready. This is not a proxy,
all it does is send an exact copy of the incoming traffic to multiple servers,
and copy back a response.

I use it on rest endpoint to duplicate the traffic to a test servers or other applications.

## Filters
Htcp can filter by :

- `command` : return response from the first server in the command line call
- `first-ko` : return the first answer with an invalid status code
- `first-ok` : return the first answer with a valid status code

New filter can be added easily.

## Improvement
- make the requests concurrent
- add timeout flag
- add unit test
- add command test
- add benchmark
