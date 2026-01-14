
# httpfromtcp

Learning HTTP by implementing it from scratch over raw TCP sockets in Go.  
No `net/http` — just bytes, parsing, and protocol basics.

## Features

- TCP server that accepts HTTP/1.1 requests
- Parses method, path, version, and headers
- Sends simple status-line + headers + body responses

## Quick Start

```bash
git clone https://github.com/CodeZeroSugar/httpfromtcp.git
cd httpfromtcp
go mod tidy
go build -o httpfromtcp ./cmd/httpfromtcp
./httpfromtcp          # listens on :8080 by default
```

Test it:

```bash
curl http://localhost:42069/
# or raw:
(printf "GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n") | nc localhost 42069
```

Expected response (basic):

```
HTTP/1.1 200 OK
Content-Type: text/plain
Content-Length: 12

Hello world
```

## Why?

Understand sockets, request parsing, CRLF, Content-Length, and HTTP mechanics without framework magic.

## Status

Very early / educational project.  
Next: better error handling, POST support, simple routing.
