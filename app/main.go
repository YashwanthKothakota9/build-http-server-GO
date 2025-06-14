package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Request struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    string
}

func responseWithBody(body string, file ...bool) []byte {
	contentType := "text/plain"
	if len(file) > 0 && file[0] {
		contentType = "application/octet-stream"
	}
	return fmt.Appendf(nil, "HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", contentType, len(body), body)
}

func responseWithEncoding(compressedData []byte) []byte {
	return fmt.Appendf(nil, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(compressedData), compressedData)
}

func compressData(data string) []byte {
	compressedData := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(compressedData)
	gz.Write([]byte(data))
	gz.Close()
	return compressedData.Bytes()
}

func parseBuffer(buffer []byte) map[string]any {
	parts := strings.SplitN(string(buffer), "\r\n\r\n", 2)
	if len(parts) != 2 {
		return map[string]any{}
	}
	requestLineHeaders := strings.Split(parts[0], "\r\n")
	requestLine := requestLineHeaders[0]
	requestParts := strings.Split(requestLine, " ")
	requestMethod := requestParts[0]
	path := requestParts[1]
	requestLineHeaders = requestLineHeaders[1:]
	requestHeaders := make(map[string]string)
	for _, header := range requestLineHeaders {
		parts := strings.SplitN(header, ": ", 2)
		if len(parts) == 2 {
			requestHeaders[parts[0]] = parts[1]
		}
	}
	requestBody := parts[1]
	return map[string]any{
		"method":  requestMethod,
		"path":    path,
		"headers": requestHeaders,
		"body":    requestBody,
	}
}

func handleRequest(request map[string]any, connection net.Conn, dir_path ...string) {
	requestMethod := request["method"].(string)
	path := request["path"].(string)
	requestHeaders := request["headers"].(map[string]string)
	requestBody := request["body"].(string)

	var response []byte
	var err error

	// Log the incoming request
	log.Printf("Received %s request for path: %s", requestMethod, path)

	switch {
	case path == "/":
		response = []byte("HTTP/1.1 200 OK\r\n\r\n")
	case path == "/user-agent":
		userAgentValue := requestHeaders["User-Agent"]
		response = responseWithBody(userAgentValue)
	case strings.HasPrefix(path, "/files"):
		fileName := path[7:]
		filePath := filepath.Join(dir_path[0], fileName)
		if requestMethod == "GET" {
			content, err := os.ReadFile(filePath)
			if err != nil {
				log.Printf("Error reading file %s: %v", filePath, err)
				response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
			} else {
				response = responseWithBody(string(content), true)
			}
		} else if requestMethod == "POST" {
			contentLengthInt, err := strconv.Atoi(requestHeaders["Content-Length"])
			if err != nil || len(requestBody) != contentLengthInt {
				log.Printf("Invalid content length for POST request: %v", err)
				response = []byte("HTTP/1.1 400 Bad Request\r\n\r\n")
			} else {
				err = os.WriteFile(filePath, []byte(requestBody), 0644)
				if err != nil {
					log.Printf("Error writing file %s: %v", filePath, err)
					response = []byte("HTTP/1.1 500 Internal Server Error\r\n\r\n")
				} else {
					response = []byte("HTTP/1.1 201 Created\r\n\r\n")
				}
			}
		}
	case strings.HasPrefix(path, "/echo"):
		randomString := path[6:]
		if acceptEncoding, ok := requestHeaders["Accept-Encoding"]; ok {
			acceptEncoding = strings.TrimSpace(acceptEncoding)
			if acceptEncoding == "gzip" || strings.Contains(acceptEncoding, "gzip") {
				compressedData := compressData(randomString)
				response = responseWithEncoding(compressedData)
			} else {
				response = responseWithBody(randomString)
			}
		} else {
			response = responseWithBody(randomString)
		}
	default:
		log.Printf("Path not found: %s", path)
		response = []byte("HTTP/1.1 404 Not Found\r\n\r\n")
	}

	// Write response to connection
	if err == nil {
		_, err = connection.Write(response)
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}

func handleConnection(connection net.Conn, dir_path ...string) {
	// Set read deadline to prevent hanging connections
	connection.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Create a buffer to accumulate data
	readBuffer := make([]byte, 0, 4096)
	tempBuffer := make([]byte, 1024)

	defer func() {
		connection.Close()
		log.Printf("Connection closed for %s", connection.RemoteAddr().String())
	}()

	for {
		// Read from the connection
		n, err := connection.Read(tempBuffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Reset the deadline for the next read
				connection.SetReadDeadline(time.Now().Add(5 * time.Second))
				continue
			}
			log.Printf("Error reading from connection: %v", err)
			return
		}

		// Append new data to our buffer
		readBuffer = append(readBuffer, tempBuffer[:n]...)

		// Process complete requests from the buffer
		for {
			// Try to find the end of a request (double newline)
			endOfRequest := bytes.Index(readBuffer, []byte("\r\n\r\n"))
			if endOfRequest == -1 {
				// No complete request found, wait for more data
				break
			}

			// Check if we have the Content-Length header
			requestStr := string(readBuffer[:endOfRequest])
			contentLength := 0
			if strings.Contains(requestStr, "Content-Length:") {
				// Extract Content-Length
				contentLengthStr := strings.Split(strings.Split(requestStr, "Content-Length:")[1], "\r\n")[0]
				contentLengthStr = strings.TrimSpace(contentLengthStr)
				contentLength, err = strconv.Atoi(contentLengthStr)
				if err != nil {
					log.Printf("Invalid Content-Length header: %v", err)
					return
				}
			}

			// Calculate total request length
			totalLength := endOfRequest + 4 // +4 for \r\n\r\n
			if contentLength > 0 {
				totalLength += contentLength
			}

			// Check if we have the complete request
			if len(readBuffer) < totalLength {
				// Don't have the complete request yet
				break
			}

			// Process the complete request
			parsedRequest := parseBuffer(readBuffer[:totalLength])
			if parsedRequest == nil {
				log.Printf("Error parsing request")
				return
			}

			// Handle the request
			if len(dir_path) > 0 {
				handleRequest(parsedRequest, connection, dir_path[0])
			} else {
				handleRequest(parsedRequest, connection)
			}

			// Remove the processed request from the buffer
			readBuffer = readBuffer[totalLength:]

			// If buffer is too large, trim it
			if cap(readBuffer) > 4096 && len(readBuffer) < 1024 {
				newBuffer := make([]byte, len(readBuffer), 4096)
				copy(newBuffer, readBuffer)
				readBuffer = newBuffer
			}
		}
	}
}

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		log.Fatalf("Failed to bind to port 4221: %v", err)
	}
	defer l.Close()

	log.Printf("Server started on port 4221")

	dir_path := ""
	if len(os.Args) > 1 {
		dir_path = os.Args[2]
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		log.Printf("New connection from %s", connection.RemoteAddr().String())

		if dir_path != "" {
			go handleConnection(connection, dir_path)
		} else {
			go handleConnection(connection)
		}
	}
}
