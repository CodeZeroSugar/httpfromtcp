package main

import (
	"fmt"
	"log"
	"net"

	"github.com/CodeZeroSugar/internal/request"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("failed to establish listener: %s", err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %s", err)
		}
		log.Println("Connection has been accepted!")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatalf("failed to get request from reader: %s", err)
		}

		method := req.RequestLine.Method
		target := req.RequestLine.RequestTarget
		version := req.RequestLine.HttpVersion

		headers := req.Headers

		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", method)
		fmt.Printf("- Target: %s\n", target)
		fmt.Printf("- Version: %s\n", version)
		fmt.Println("Headers:")
		for key, value := range headers {
			fmt.Printf("- %s: %s\n", key, value)
		}

	}
}
