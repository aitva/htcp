# htcp
Htcp is a command line utility to duplicate HTTP traffic. It can duplicate an HTTP request
and send it to as many servers as you need. Filters can be apply to select which answer
must be returned to the client.

## Goal
I work on REST applications and sometime I need to duplicate the traffic from one
entry-point to two applications. Since I have faced this problem more than once,
I have built this small tool.

## Filters
Htcp can filter by :

- command : return response from the first server in the command line call
- error : return the first answer with an invalid status code
- ok : return the first answer with a valid status code

New filter can be added easily.

## Concurent
Htcp is not concurrent at this time, but the next version is going to
execute all the request concurrently to improve response time from the client.

## Apache
Explanation on how to connect Apache2 to htcp.
