package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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

	request := []byte{}
	_, err = conn.Read(request)

	if err != nil {
		fmt.Println("Error reading HTTP request: ", err.Error())
		os.Exit(1)
	}

	req_list := strings.Split(string(request), " ")
	request_target := req_list[1]

	var status_code string
	var message string
	if request_target == "/" {
		status_code = "200"
		message = "OK"
	} else {
		status_code = "404"
		message = "NOT FOUND"
	}

	protocol := "HTTP/1.1"
	status_line := protocol + " " + status_code + " " + message + "\r\n\r\n"
	_, err = conn.Write([]byte(status_line))

	if err != nil {
		fmt.Println("Error sending HTTP status line: ", err.Error())
		os.Exit(1)
	}

}
