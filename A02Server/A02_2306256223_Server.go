package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
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
	req := HttpRequest{Headers: make(map[string]string)}
	lines := strings.Split(string(bytestream), "\r\n")

	if len(lines) < 1 {
		return req
	}
	parts := strings.Split(lines[0], " ")
	if len(parts) == 3 {
		req.Method = parts[0]
		req.URI = parts[1]
		req.Proto = parts[2]
	}

	i := 1
	for ; i < len(lines); i++ {
		if lines[i] == "" {
			break
		}
		if idx := strings.Index(lines[i], ":"); idx != -1 {
			req.Headers[strings.ToLower(strings.TrimSpace(lines[i][:idx]))] = strings.TrimSpace(lines[i][idx+1:])
		}
	}

	if i+1 < len(lines) {
		req.Body = []byte(strings.Join(lines[i+1:], "\r\n"))
	}

	return req
}

func ResponseEncoder(res HttpResponse) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s %d %s\r\n", res.Proto, res.StatusCode, res.StatusText))
	for k, v := range res.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	buf.WriteString("\r\n")
	buf.Write(res.Body)
	return buf.Bytes()
}

// ─── HANDLER ────────────────────────────────────────────────────
func handleRequest(req HttpRequest) HttpResponse {
	resp := HttpResponse{
		Proto:      "HTTP/1.1",
		StatusCode: 404,
		StatusText: "Not Found",
		Headers:    map[string]string{"Content-Length": "0"},
	}

	if req.Method != "GET" {
		return resp
	}

	if req.URI == "/" {
		body := []byte(fmt.Sprintf("Halo, dunia! Aku %s", config.Name))
		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Headers["Content-Type"] = "text/html"
		resp.Body = body
		resp.Headers["Content-Length"] = fmt.Sprint(len(body))
		return resp
	}

	u, err := url.Parse(req.URI)
	if err != nil {
		return resp
	}
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "greet" {
		if parts[1] != config.NPM {
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
		var ctype string
		if strings.Contains(accept, "application/xml") && !strings.Contains(accept, "application/json") {
			body, _ = xml.MarshalIndent(gresp, "", "  ")
			ctype = "application/xml"
		} else {
			body, _ = json.MarshalIndent(gresp, "", "  ")
			ctype = "application/json"
		}
		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Body = body
		resp.Headers["Content-Type"] = ctype
		resp.Headers["Content-Length"] = fmt.Sprint(len(body))
	}

	return resp
}

// ─── MAIN & SOCKET ──────────────────────────────────────────────
func main() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "listen:", err)
		return
	}
	defer ln.Close()

	fmt.Printf("Listening on port %d ...\n", config.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "accept:", err)
			continue
		}
		go func(c net.Conn) {
			defer c.Close()
			buf := make([]byte, 4096)
			n, _ := c.Read(buf)
			req := RequestDecoder(buf[:n])
			res := handleRequest(req)
			c.Write(ResponseEncoder(res))
		}(conn)
	}
}
