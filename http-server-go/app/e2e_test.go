package main

import (
	"bufio"
	"net"
	"os"
	"testing"
	"time"
)

// Global variable to hold the server listener
var listener net.Listener

// TestMain sets up the server before tests and tears it down afterward
func TestMain(m *testing.M) {
	// Start the server for e2e tests
	go func() {
		main()
	}()

	// Give the server some time to start
	time.Sleep(1 * time.Second)

	// Run tests
	code := m.Run()

	// Shutdown the server (if applicable) and exit
	if listener != nil {
		listener.Close()
	}
	os.Exit(code)
}

func TestServerResponse(t *testing.T) {
	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:4221")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	request := "GET / HTTP/1.1\r\n\r\n"
	conn.Write([]byte(request))

	reader := bufio.NewReader(conn)
	responseLine, _ := reader.ReadString('\n')

	expectedStatus := "HTTP/1.1 200 OK\r\n"
	if responseLine != expectedStatus {
		t.Errorf("Expected %q, got %q", expectedStatus, responseLine)
	}
}
