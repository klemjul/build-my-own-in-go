package main

import (
	"strconv"
	"strings"
)

type HttpRequest struct {
	method   string
	path     string
	protocol string
	headers  map[string]string
	body     []byte
}

type HttpResponse struct {
	status     int
	statusText string
	body       []byte
	headers    map[string]string
}

func handleRequest(req HttpRequest) HttpResponse {
	pathParts := strings.Split(req.path, "/")
	if pathParts[0] == "echo" {
		return HttpResponse{
			status:     200,
			statusText: "OK",
			body:       req.body,
			headers:    map[string]string{"Content-Type": "text/plain", "Content-Length": strconv.Itoa(len(req.body))},
		}
	}
	if pathParts[0] == "" {
		return HttpResponse{
			status:     200,
			statusText: "OK",
			body:       make([]byte, 0),
			headers:    make(map[string]string),
		}
	}
	return HttpResponse{
		status:     404,
		statusText: "Not Found",
		body:       make([]byte, 0),
		headers:    make(map[string]string),
	}
}
