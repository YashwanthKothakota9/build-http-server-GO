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

func main() {

	// fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	connection, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	buffer := make([]byte, 1024)
	connection.Read(buffer)
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

	connection.Close()
}
