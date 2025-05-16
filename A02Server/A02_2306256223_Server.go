package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
	"strings"
	"syscall"
)

// ─── CONFIG ───────────────────────────────────────────────────────
var config = struct {
	Name string
	NPM  string
	Port int
}{
	Name: "Ameera Khaira Tawfiqa",
	NPM:  "2306256223",
	Port: 8080,
}

// ─── HTTP STRUCTS & ENCODERS ────────────────────────────────────
type HttpRequest struct {
	Method  string
	URI     string
	Proto   string
	Headers map[string]string
	Body    []byte
}

type HttpResponse struct {
	Proto      string
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       []byte
}

type Student struct {
	Nama string `json:"Nama" xml:"Nama"`
	Npm  string `json:"Npm"  xml:"Npm"`
}

type GResp struct {
	XMLName xml.Name `xml:"Response" json:"-"`
	Student Student  `json:"Student" xml:"Student"`
	Greeter string   `json:"Greeter"  xml:"Greeter"`
}

// ─── DECODER & ENCODER ──────────────────────────────────────────

func RequestDecoder(bytestream []byte) HttpRequest {
	// TODO: Implement request decoding
	req := HttpRequest{Headers: make(map[string]string)}
	lines := strings.Split(string(bytestream), "\r\n")

	// parse request line
	if len(lines) < 1 {
		return req
	}
	parts := strings.Split(lines[0], " ")
	if len(parts) == 3 {
		req.Method = parts[0]
		req.URI = parts[1]
		req.Proto = parts[2]
	}

	// parse headers
	i := 1
	for ; i < len(lines); i++ {
		if lines[i] == "" {
			break
		}
		colon := strings.Index(lines[i], ":")
		if colon != -1 {
			key := strings.TrimSpace(lines[i][:colon])
			val := strings.TrimSpace(lines[i][colon+1:])
			req.Headers[strings.ToLower(key)] = val
		}
	}

	if i+1 < len(lines) {
		req.Body = []byte(strings.Join(lines[i+1:], "\r\n"))
	}

	return req
}

func ResponseEncoder(res HttpResponse) []byte {
	// TODO: Implement response encoding
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s %d %s\r\n", res.Proto, res.StatusCode, res.StatusText))
	for k, v := range res.Headers {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buffer.WriteString("\r\n")
	buffer.Write(res.Body)
	return buffer.Bytes()
}

// ─── HANDLER ────────────────────────────────────────────────────
func handleRequest(req HttpRequest) HttpResponse {

	// Default response
	resp := HttpResponse{
		Proto:      "HTTP/1.1",
		StatusCode: 404,
		StatusText: "Not Found",
		Headers:    map[string]string{"Content-Length": "0"},
	}

	// TODO: Implement request handlers
	// 1) GET /
	if req.Method != "GET" {
		return resp
	}

	// 2) GET /greet/{npm}[?name=...]
	if req.URI == "/" {
		body := []byte(fmt.Sprintf("Halo, dunia! Aku %s", config.Name))
		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Headers["Content-Type"] = "text/html"
		resp.Body = body
		resp.Headers["Content-Length"] = fmt.Sprint(len(body))
		return resp
	}

	// parse URI
	uri := req.URI
	u, err := url.Parse(uri)
	if err != nil {
		return resp
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) >= 2 && parts[0] == "greet" {
		npm := parts[1]
		if npm != config.NPM {
			return resp
		}

		name := u.Query().Get("name")
		if name == "" {
			name = config.Name
		}

		student := Student{Nama: config.Name, Npm: config.NPM}
		gresp := GResp{Student: student, Greeter: name}

		accept := strings.ToLower(req.Headers["accept"])
		var body []byte
		var contentType string

		if strings.Contains(accept, "application/xml") && !strings.Contains(accept, "application/json") {
			body, _ = xml.MarshalIndent(gresp, "", "  ")
			contentType = "application/xml"
		} else {
			body, _ = json.MarshalIndent(gresp, "", "  ")
			contentType = "application/json"
		}

		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Body = body
		resp.Headers["Content-Type"] = contentType
		resp.Headers["Content-Length"] = fmt.Sprint(len(body))
		return resp
	}

	return resp
}

// ─── MAIN & SOCKET ──────────────────────────────────────────────
func main() {
	// Create socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Fprintln(os.Stderr, "socket:", err)
		return
	}
	defer syscall.Close(fd)

	// Bind to address
	sa := &syscall.SockaddrInet4{Port: config.Port}
	copy(sa.Addr[:], []byte{0, 0, 0, 0})
	if err := syscall.Bind(fd, sa); err != nil {
		fmt.Fprintln(os.Stderr, "bind:", err)
		return
	}

	// TODO: Implement bind, listen, and accept connections
	if err := syscall.Listen(fd, 10); err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		return
	}

	fmt.Printf("Listening on port %d ...\n", config.Port)

	// TODO: Implement connection handling
	for {
		connFd, _, err := syscall.Accept(fd)
		if err != nil {
			fmt.Fprintln(os.Stderr, "accept:", err)
			continue
		}

		go func(cfd syscall.Handle) {
			defer syscall.Close(cfd)
			buf := make([]byte, 4096)
			n, _ := syscall.Read(cfd, buf)

			req := RequestDecoder(buf[:n])
			resp := handleRequest(req)
			respBytes := ResponseEncoder(resp)
			syscall.Write(cfd, respBytes)
		}(connFd)
	}
}
