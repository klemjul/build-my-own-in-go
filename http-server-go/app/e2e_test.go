package main

import (
	"io"
	"log"
	"net"
	"net/http"
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
	resp, err := http.Get("http://localhost:4221")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	log.Printf("Response: %s", body)

	if resp.Status != "200 OK" {
		t.Errorf("Expected %q, got %q", "200 OK", resp.Status)
	}
}

func TestServerResponseWithBody(t *testing.T) {
	resp, err := http.Get("http://localhost:4221/echo/abc")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read and process the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}
	log.Printf("Response: %s", body)

	expectedBody := "abc"
	if string(body) != expectedBody {
		t.Errorf("Expected body = %q, got %q", expectedBody, string(body))
	}

	hContentType := resp.Header.Get("Content-Type")
	expectedContentType := "text/plain"
	if hContentType != expectedContentType {
		t.Errorf("Expected Header.Content-Type = %q, got %q", expectedContentType, hContentType)
	}

	hContentLength := resp.Header.Get("Content-Length")
	expectedContentLength := "3"
	if hContentLength != expectedContentLength {
		t.Errorf("Expected Header.Content-Length = %q, got %q", expectedContentLength, hContentLength)
	}
}
