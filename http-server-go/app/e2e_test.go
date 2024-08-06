package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

// Global variable to hold the server listener
var listener net.Listener
var baseUrl = "http://localhost:4221"

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

func TestResponse(t *testing.T) {
	resp, err := http.Get(baseUrl)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.Status != "200 OK" {
		t.Errorf("Expected %q, got %q", "200 OK", resp.Status)
	}
}

func TestResponseWithBody(t *testing.T) {
	resp, err := http.Get(baseUrl + "/echo/abc")
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

func TestReadHeader(t *testing.T) {
	customUserAgent := "MyCustomUserAgent/1.0"
	req, _ := http.NewRequest("GET", baseUrl+"/user-agent", nil)
	req.Header.Set("User-Agent", customUserAgent)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if string(body) != customUserAgent {
		t.Errorf("Expected body = %q, got %q", customUserAgent, string(body))
	}

	hContentType := resp.Header.Get("Content-Type")
	expectedContentType := "text/plain"
	if hContentType != expectedContentType {
		t.Errorf("Expected Header.Content-Type = %q, got %q", expectedContentType, hContentType)
	}

	hContentLength := resp.Header.Get("Content-Length")
	expectedContentLength := strconv.Itoa((len(customUserAgent)))
	if hContentLength != expectedContentLength {
		t.Errorf("Expected Header.Content-Length = %q, got %q", expectedContentLength, hContentLength)
	}
}

func TestConcurrentConnection(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(baseUrl)
			if err != nil {
				log.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			if resp.Status != "200 OK" {
				t.Errorf("Expected %q, got %q", "200 OK", resp.Status)
			}
		}()
	}
	wg.Wait()
}

func TestFileFound(t *testing.T) {
	resp, err := http.Get(baseUrl + "/files/README.md")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	expectedFile, _ := os.ReadFile("../../README.md")
	if err != nil {
		log.Fatalf("Failed to READ FILE: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)

	hContentType := resp.Header.Get("Content-Type")
	expectedContentType := "application/octet-stream"
	if hContentType != expectedContentType {
		t.Errorf("Expected Header.Content-Type = %q, got %q", expectedContentType, hContentType)
	}

	hContentLength := resp.Header.Get("Content-Length")
	expectedContentLength := strconv.Itoa((len(expectedFile)))
	if hContentLength != expectedContentLength {
		t.Errorf("Expected Header.Content-Length = %q, got %q", expectedContentLength, hContentLength)
	}

	if string(body) != string(expectedFile) {
		t.Errorf("Expected body = %q, got %q", string(expectedFile), string(body))
	}
}

func TestFileNotFound(t *testing.T) {
	resp, err := http.Get(baseUrl + "/files/non_existant_file")
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.Status != "404 Not Found" {
		t.Errorf("Expected %q, got %q", "404 Not Found", resp.Status)
	}
}
