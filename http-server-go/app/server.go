package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		log.Fatal(err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 4221")

	// Infinite loop to accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	requestLine, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Error reading request line:", err)
		return
	}

	requestLine = strings.TrimSpace(requestLine)
	requestParsed := strings.SplitN(requestLine, " ", 3)

	// Read headers
	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading header:", err)
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break // End of headers
		}
		headerParts := strings.SplitN(line, ":", 2)
		if len(headerParts) == 2 {
			headers[strings.TrimSpace(headerParts[0])] = strings.TrimSpace(headerParts[1])
		}
	}

	var body []byte
	// Optional: Read the request body if the request method is POST
	if headers["Content-Length"] != "" {
		contentLength := headers["Content-Length"]
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			log.Println("Error parsing Content-Length:", err)
			return
		}
		body := make([]byte, length)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			log.Println("Error reading body:", err)
			return
		}
		fmt.Println("Body:", string(body))
	}

	request := HttpRequest{
		method:   requestParsed[0],
		path:     requestParsed[1],
		protocol: requestParsed[2],
		body:     body,
		headers:  headers,
	}
	log.Println(request)
	response := handleRequest(request)
	responseText := buildResponseText((response))

	// Send a basic HTTP response
	_, err = conn.Write([]byte(responseText))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}
