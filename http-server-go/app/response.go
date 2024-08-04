package main

import (
	"fmt"
	"strconv"
	"strings"
)

type HttpResponse struct {
	status     int
	statusText string
	body       []byte
	headers    map[string]string
}

func buildResponseText(response HttpResponse) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("HTTP/1.1 %s %s\r\n", strconv.Itoa(response.status), response.statusText))
	for key, value := range response.headers {
		builder.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	builder.WriteString("\r\n")
	if len(response.body) > 0 {
		builder.WriteString(string(response.body))
	}
	return builder.String()
}
