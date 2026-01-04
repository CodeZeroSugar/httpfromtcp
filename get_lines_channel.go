package main

import (
	"io"
	"log"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	var currentLine string
	b := make([]byte, 8)
	ch := make(chan string)

	go func() {
		for {
			n, err := f.Read(b)
			if err == io.EOF {
				f.Close()
				break
			}
			if err != nil {
				log.Fatal(err)
			}

			str := string(b[:n])
			parts := strings.Split(str, "\n")
			for i, part := range parts {
				if i == len(parts)-1 {
					currentLine += part
				} else {
					currentLine += part

					ch <- currentLine

					currentLine = ""
				}
			}
		}
		if len(currentLine) > 0 {
			ch <- currentLine
		}
		close(ch)
	}()
	return ch
}
