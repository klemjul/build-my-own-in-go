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

func handleRequest(req HttpRequest) HttpResponse {
	pathParts := strings.Split(req.path, "/")
	if pathParts[1] == "echo" {
		return HttpResponse{
			status:     200,
			statusText: "OK",
			body:       []byte(pathParts[2]),
			headers:    map[string]string{"Content-Type": "text/plain", "Content-Length": strconv.Itoa(len(pathParts[2]))},
		}
	}
	if pathParts[1] == "" {
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
