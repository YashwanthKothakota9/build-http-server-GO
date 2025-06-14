package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func responseWithBody(body string) []byte {
	return []byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(body), body))
}

func handleConnection(connection net.Conn) {
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
	path := requestParts[1]

	if path == "/" {
		connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if path == "/user-agent" {
		userAgentHeader := requestLines[2]
		userAgentValue := strings.Split(userAgentHeader, ": ")[1]
		connection.Write(responseWithBody(userAgentValue))
	} else if strings.HasPrefix(path, "/echo") {
		randomString := path[6:]
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

	fmt.Println("Server started on port 4221")

	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		// Handle each connection in a separate goroutine
		go handleConnection(connection)
	}
}
