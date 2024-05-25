package main

import (
	"bufio"
	"flag"
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

func handle_echo_view(request HTTPRequest, args map[string]string) HTTPResponse {
	status_code := 200
	message := "OK"
	body, _ := strings.CutPrefix(request.url, "/echo/")
	version := request.version
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

func handle_files_view(request HTTPRequest, args map[string]string) HTTPResponse {
	status_code := 200
	message := "OK"
	body := ""
	version := request.version
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))
	directory := args["directory"]

	basename, _ := strings.CutPrefix(request.url, "/files/")
	filename := directory + basename

	fmt.Printf("test %s\n", filename)

	// Open the file
	file, err := os.Open(filename)
	if err != nil {
		status_code = 404
		message = "Not Found"
		return HTTPResponse{version, status_code, message, headers, body}
	}
	defer file.Close()

	// Create a new scanner for the file
	scanner := bufio.NewScanner(file)

	// Read and print the file line by line
	for scanner.Scan() {
		body += scanner.Text()
	}

	headers["Content-Type"] = "application/octet-stream"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

func handle_base_view(request HTTPRequest, args map[string]string) HTTPResponse {
	status_code := 200
	message := "OK"
	body := ""
	version := request.version
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

func handle_user_agent_view(request HTTPRequest, args map[string]string) HTTPResponse {
	status_code := 200
	message := "OK"
	body := request.headers["User-Agent"]
	version := request.version
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

func generate_http_response(request HTTPRequest, args map[string]string) HTTPResponse {
	handle_exact_request_map := map[string]func(HTTPRequest, map[string]string) HTTPResponse{
		"/":           handle_base_view,
		"/user-agent": handle_user_agent_view,
	}

	handle_prefix_request_map := map[string]func(HTTPRequest, map[string]string) HTTPResponse{
		"/echo":  handle_echo_view,
		"/files": handle_files_view,
	}

	for url, handle := range handle_exact_request_map {
		if request.url == url {
			return handle(request, args)
		}
	}

	for prefix_url, handle := range handle_prefix_request_map {
		if strings.HasPrefix(request.url, prefix_url) {
			return handle(request, args)
		}
	}

	status_code := 404
	message := "Not Found"
	version := request.version
	body := ""
	headers := make(map[string]string)
	headers["Content-Type"] = "text/plain"
	headers["Content-Length"] = fmt.Sprint(len(body))

	return HTTPResponse{version, status_code, message, headers, body}
}

func handle_connection(conn net.Conn, args map[string]string) {
	reader := bufio.NewReader(conn)
	request := parse_http_request(reader)
	response := generate_http_response(request, args)
	response_string := response.to_string()

	_, err := conn.Write([]byte(response_string))

	if err != nil {
		fmt.Println("Error sending response string: ", err.Error())
		os.Exit(1)
	}

	conn.Close()
}

func main() {
	var directory string
	args := make(map[string]string)

	// Define flags without setting default values
	flag.StringVar(&directory, "directory", "", "directory to check if the file exist")

	// Parse the flags
	flag.Parse()

	args["directory"] = directory

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		go handle_connection(conn, args)
	}

}
