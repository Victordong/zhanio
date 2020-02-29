package main

import (
	"fmt"
	"net"
)

func Echo(c net.Conn) {
	defer c.Close()
	line := make([]byte, 1024)
	length, _ := c.Read(line)
	_, _ = c.Write(line[:length])
}

func main() {
	fmt.Printf("Server is ready...\n")
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Printf("Failure to listen: %s\n", err.Error())
	}

	for {
		if c, err := l.Accept(); err == nil {
			go Echo(c) //new thread
		}
	}
}
