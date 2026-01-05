package main

import (
	"fmt"
	"log"
	"net"
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

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}
		_, ok := <-lines
		if !ok {
			log.Println("Connection has been closed!")
		}

	}
}
