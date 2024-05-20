package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	version = "HTTP/1.1"
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
	builder.WriteString(version)
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

	return builder.String()
}

func parse_http_request(reader *bufio.Reader) HTTPRequest {
	data := make([]byte, 4096)
	_, err := reader.Read(data)

	if err != nil {
		fmt.Println("Error reading: ", err.Error())
		os.Exit(1)
	}
	data_string := string(data)
	data_split := strings.Split(data_string, "\r\n")
	status_line_splits := strings.Split(data_split[0], " ")
	method, url := status_line_splits[0], status_line_splits[1]

	headers := make(map[string]string)

	for i := 1; i < len(data_split)-2; i++ {
		header := strings.SplitN(data_split[i], ":", 2)
		if len(header) == 2 {
			headers[strings.TrimSpace(header[0])] = strings.TrimSpace(header[1])
		}
	}
	body := data_split[len(data_split)-1]
	return HTTPRequest{method, url, version, headers, body}
}

func generate_http_response(request HTTPRequest) HTTPResponse {
	var status_code int
	var message string

	if request.url == "/" || request.url == "/user-agent" || strings.HasPrefix(request.url, "/echo/") {
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
	} else if request.url == "/user-agent" {
		body = request.headers["User-Agent"]
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

	reader := bufio.NewReader(conn)
	request := parse_http_request(reader)

	if err != nil {
		fmt.Println("Error reading HTTP request: ", err.Error())
		os.Exit(1)
	}

	response := generate_http_response(request)
	response_string := response.to_string()
	_, err = conn.Write([]byte(response_string))
	fmt.Printf("test: " + response_string)

	if err != nil {
		fmt.Println("Error sending response string: ", err.Error())
		os.Exit(1)
	}

	conn.Close()

}
