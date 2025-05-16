package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var config = struct {
	Name string
	NPM  string
	Port int
}{
	Name: "Ameera Khaira Tawfiqa",
	NPM:  "2306256223",
	Port: 6223,
}

// ─── HTTP STRUCTS ─────────────────────────────
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

// ─── DECODER & ENCODER ───────────────────────
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
			key := strings.ToLower(strings.TrimSpace(lines[i][:idx]))
			val := strings.TrimSpace(lines[i][idx+1:])
			req.Headers[key] = val
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

// ─── HANDLER ─────────────────────────────────
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

	// Root
	if req.URI == "/" {
		res := HttpResponse{
			Proto:   "HTTP/1.1",
			Headers: make(map[string]string),
		}
		res.StatusCode = 200
		res.StatusText = "OK"
		res.Headers["Content-Type"] = "text/html"
		res.Body = []byte(fmt.Sprintf("<html><body><h1>Halo, dunia! Aku %s</h1></body></html>", config.Name))
		res.Headers["Content-Length"] = strconv.Itoa(len(res.Body))
		return res
	}

	u, err := url.Parse(req.URI)
	if err != nil {
		return resp
	}

	// /greet/{npm}?name=...
	parts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(parts) >= 2 && parts[0] == "greet" {
		if parts[1] != config.NPM {
			return resp
		}
		greeter := u.Query().Get("name")
		greeter = strings.Trim(greeter, `"`) // remove extra quotes

		if greeter == "" {
			greeter = config.Name
		}

		student := Student{Nama: config.Name, Npm: config.NPM}
		gresp := GResp{Student: student, Greeter: greeter}
		accept := strings.ToLower(req.Headers["accept"])
		var body []byte
		var ctype string

		if strings.Contains(accept, "application/xml") && !strings.Contains(accept, "application/json") {
			body, _ = xml.MarshalIndent(gresp, "", "  ")
			ctype = "application/xml"
		} else if strings.Contains(accept, "application/json") {
			body, _ = json.MarshalIndent(gresp, "", "  ")
			ctype = "application/json"
		} else {
			return resp
		}

		resp.StatusCode = 200
		resp.StatusText = "OK"
		resp.Body = body
		resp.Headers["Content-Type"] = ctype
		resp.Headers["Content-Length"] = fmt.Sprint(len(body))
	}

	return resp
}

// ─── MAIN ─────────────────────────────────────
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
