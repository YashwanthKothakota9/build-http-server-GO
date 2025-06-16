[![progress-banner](https://backend.codecrafters.io/progress/http-server/35633b1d-9461-4048-b571-2fdacaf174e8)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is a starting point for Python solutions to the
["Build Your Own HTTP server" Challenge](https://app.codecrafters.io/courses/http-server/overview).

[HTTP](https://en.wikipedia.org/wiki/Hypertext_Transfer_Protocol) is the
protocol that powers the web. In this challenge, you'll build a HTTP/1.1 server
that is capable of serving multiple clients.

Along the way you'll learn about TCP servers,
[HTTP request syntax](https://www.w3.org/Protocols/rfc2616/rfc2616-sec5.html),
and more.

<h1 align="center">HTTP Server from scratch in GO</h1>

<div align="center">
    <img src="/image.png" alt="Project completion image">
</div>









### Stage 1: Bind server to a port
- Create a server using `socket` module
- bind it to a port `4221`
- wait for incoming client connections using `accept()`

### Stage 2: HTTP Response to the connection
- Server should respond to the accepted connection with `200` http response.

### Stage 3: URL Path parsing from request
- Server should send response based on the URL path in the HTTP request.
- HTTP `200` response for `GET \` request
- HTTP `404` response for anything other than above url path.

### Stage 4: Response with body
- Server should reponse with body
- HTTP `200` response for `GET \echo\{str}` request with response body of `{str}`

### Stage 5: Read header 
- Server should response with body the value of `User-agent` from headers
- HTTP `200` response for `GET \user-agent` request with response body with `User-agent`s value from headers. 

### Stage 6: Concurrent connections
- Server should support `concurrent connections`.
- Used `ThreadPoolExecutor` to support `concurrent connections`.
- Used `threads` instead of `processes` because http server is a `I/O` heavy one.

### Stage 7: Return a file
- Read `directory name` from `command line arguments` - `argv`
- HTTP `200` response for `GET \files\{file_name}` request with response body of file content.
- `Content-Type` header set to `application/octet-stream`.
- `Content-Length` header set to the `size of the file, in bytes`.

### Stage 8: Read Request body
- HTTP `201` response for `POST \files\{file_name}` request with request body contains `data` to write to the file mentioned in the URL path.

### Stage 9: Accept-Encoding header
- Server should respond based on `Accept-Encoding` header value with a `Content-Encoding` header value in response


### Stage 10: Accept-Encoding header with multiple values
- Server should respond based on multiple values in `Accept-Encoding` header 

### Stage 11: Add gzip compression of request body
- Server should compress the request body using `gzip` and append those with `headers` in bytes
- This should be done when `Accept-Encoding` header present with value `gzip`

### Stage 12: Persistent Connection
- Keep the TCP connection open after the first request
- Handle any subsequent requests on the same connection
- Return appropriate responses for both requests

### Stage 13: Add Support for Multiple Persistent Connections
- Handle multiple concurrent TCP connections
- Keep each connection open for multiple requests
- Process requests independently on each connection
- Return appropriate responses for all requests
  
### Stage 14: Close the Connection
- Server should close the connection if request header contains `Connection:close`



### Extra features
- Added logging
- Every connection has read deadline to remove hanging connections
- Error Handling
