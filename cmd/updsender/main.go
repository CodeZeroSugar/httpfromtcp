package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	res, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatalf("failed to resolve address: %s", err)
	}
	conn, err := net.DialUDP("udp", nil, res)
	if err != nil {
		log.Fatalf("failed to get udp connection: %s", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("> ")

		str, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("failed to read string: %s", err)
		}

		_, err = conn.Write([]byte(str))
		if err != nil {
			log.Printf("failed to write string to connection: %s", err)
		}
	}
}
