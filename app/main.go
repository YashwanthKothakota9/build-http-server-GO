package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

func responseWithBody(body string, file ...bool) []byte {
	contentType := "text/plain"
	if len(file) > 0 && file[0] {
		contentType = "application/octet-stream"
	}
	return []byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: %s\r\nContent-Length: %d\r\n\r\n%s", contentType, len(body), body))
}

func responseWithEncoding(compressedData []byte) []byte {
	return []byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Encoding: gzip\r\nContent-Length: %d\r\n\r\n%s", len(compressedData), compressedData))
}

func compressData(data string) []byte {
	compressedData := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(compressedData)
	gz.Write([]byte(data))
	gz.Close()
	return compressedData.Bytes()
}

func handleConnection(connection net.Conn, dir_path ...string) {
	defer connection.Close()

	buffer := make([]byte, 1024)
	_, err := connection.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err.Error())
		return
	}

	requestLines := strings.Split(string(buffer), "\r\n")
	requestLine := requestLines[0]
	requestParts := strings.Split(requestLine, " ")
	requestMethod := requestParts[0]
	path := requestParts[1]

	if path == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if path == "/user-agent" {
		userAgentHeader := requestLines[2]
		userAgentValue := strings.Split(userAgentHeader, ": ")[1]
		connection.Write(responseWithBody(userAgentValue))
	} else if strings.HasPrefix(path, "/files") {
		fileName := path[7:]
		filePath := filepath.Join(dir_path[0], fileName)
		if requestMethod == "GET" {
			content, err := os.ReadFile(filePath)
			if err != nil {
				connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			connection.Write(responseWithBody(string(content), true))
		} else if requestMethod == "POST" {
			contentLength := strings.Split(requestLines[2], ": ")[1]
			contentLengthInt, err := strconv.Atoi(string(contentLength))
			if err != nil {
				connection.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
				return
			}

			// Split the request by double CRLF to separate headers and body
			parts := strings.SplitN(string(buffer), "\r\n\r\n", 2)
			if len(parts) != 2 {
				connection.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
				return
			}

			// Take exactly contentLengthInt bytes from the body
			requestBody := parts[1][:contentLengthInt]

			if len(requestBody) != contentLengthInt {
				connection.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
				return
			}
			os.WriteFile(filePath, []byte(requestBody), 0644)
			connection.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}

	} else if strings.HasPrefix(path, "/echo") {
		randomString := path[6:]
		if len(requestLines) > 2 && requestLines[2] != "" {
			compressionMethods := strings.Split(strings.Split(requestLines[2], ": ")[1], ", ")
			if len(compressionMethods) != 0 {
				if slices.Contains(compressionMethods, "gzip") {
					compressedData := compressData(randomString)
					connection.Write(responseWithEncoding(compressedData))
				} else {
					connection.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\n"))
				}
			}
		}
		connection.Write(responseWithBody(randomString))
	} else {
		connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	// fmt.Println("Server started on port 4221")

	dir_path := ""
	if len(os.Args) > 1 {
		dir_path = os.Args[2]
	}

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		if dir_path != "" {
			go handleConnection(connection, dir_path)
		} else {
			go handleConnection(connection)
		}
	}
}
