package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	status_line := "HTTP/1.1 200 OK\r\n\r\n"
	_, err = conn.Write([]byte(status_line))

	if err != nil {
		fmt.Println("Error sending HTTP status line: ", err.Error())
		os.Exit(1)
	}

}
