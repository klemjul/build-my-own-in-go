package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
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
	req, _ := http.NewRequest("GET", baseUrl+"/echo/abc", nil)
	req.Header.Set("Accept-Encoding", "identity") // workaround to disable compression
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
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

	hContentEncoding := resp.Header.Get("Content-Encoding")
	if hContentEncoding != "" {
		t.Errorf("Expected Header.Content-Encoding = %q, got %q", "", hContentEncoding)
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
	tempFile, err := os.CreateTemp("", "TestFileFound-*")
	if err != nil {
		log.Fatalf("Failed to Create Temp File: %v", err)
	}
	defer os.Remove(tempFile.Name())
	_, _ = tempFile.WriteString("This is some example content.")
	tempFile.Sync()

	tempFile.Seek(0, io.SeekStart) // Go back to the start of the file
	expectedFile, _ := io.ReadAll(tempFile)

	resp, err := http.Get(baseUrl + "/files/" + filepath.Base((tempFile.Name())))
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

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

func TestGzipEncoding(t *testing.T) {
	customEncoding := "gzip"
	req, _ := http.NewRequest("GET", baseUrl+"/echo/abc", nil)
	req.Header.Set("Accept-Encoding", customEncoding)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var compressedData bytes.Buffer
	writer := gzip.NewWriter(&compressedData)
	writer.Write([]byte("abc"))
	writer.Close()

	if string(body) != compressedData.String() {
		t.Errorf("Expected %q, got %q", compressedData.String(), string(body))
	}

	hContentEncoding := resp.Header.Get("Content-Encoding")
	if hContentEncoding != "gzip" {
		t.Errorf("Expected Header.Content-Encoding = %q, got %q", "gzip", hContentEncoding)
	}

	hContentType := resp.Header.Get("Content-Type")
	expectedContentType := "text/plain"
	if hContentType != expectedContentType {
		t.Errorf("Expected Header.Content-Type = %q, got %q", expectedContentType, hContentType)
	}
}
