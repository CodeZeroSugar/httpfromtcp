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
		defer f.Close()
		defer close(ch)
		for {
			n, err := f.Read(b)
			if n > 0 {
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
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Println(err)
				return
			}

		}
		if len(currentLine) > 0 {
			ch <- currentLine
		}
	}()
	return ch
}
