package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type HTTPResponse struct {
	version     string
	status_code int
	message     string
	headers     map[string]string
	body        string
}

type HTTPRequest struct {
	method  string
	url     string
	version string
	headers map[string]string
	body    string
}

func (res HTTPResponse) to_string() string {
	var builder strings.Builder
	builder.WriteString(res.version)
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprint(res.status_code))
	builder.WriteString(" ")
	builder.WriteString(res.message)
	builder.WriteString("\r\n")

	for header_name, header_value := range res.headers {
		builder.WriteString(header_name)
		builder.WriteString(": ")
		builder.WriteString(header_value)
		builder.WriteString("\r\n")
	}
	builder.WriteString("\r\n")

	builder.WriteString(res.body)
	builder.WriteString("\r\n")

	return builder.String()
}

func parse_http_request(scanner *bufio.Scanner) HTTPRequest {
	scanner.Scan()
	status_line_splits := strings.Split(scanner.Text(), " ")
	method, url, version := status_line_splits[0], status_line_splits[1], status_line_splits[2]

	headers := make(map[string]string)
	for scanner.Scan() {
		header := strings.SplitN(scanner.Text(), ":", 2)
		if len(header) == 2 {
			headers[strings.TrimSpace(header[0])] = strings.TrimSpace(header[1])
		}
	}

	body := ""
	return HTTPRequest{method, url, version, headers, body}
}

func generate_http_response(request HTTPRequest) HTTPResponse {
	var status_code int
	var message string

	if request.url == "/" || strings.HasPrefix(request.url, "/echo/") {
		status_code = 200
		message = "OK"
	} else {
		status_code = 404
		message = "Not Found"
	}

	version := request.version

	body := ""
	if strings.HasPrefix(request.url, "/echo/") {
		body, _ = strings.CutPrefix(request.url, "/echo/")
	}

	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

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

	scanner := bufio.NewScanner(conn)
	request := parse_http_request(scanner)

	if err != nil {
		fmt.Println("Error reading HTTP request: ", err.Error())
		os.Exit(1)
	}

	response := generate_http_response(request)
	response_string := response.to_string()
	_, err = conn.Write([]byte(response_string))

	if err != nil {
		fmt.Println("Error sending HTTP status line: ", err.Error())
		os.Exit(1)
	}

	conn.Close()

}
