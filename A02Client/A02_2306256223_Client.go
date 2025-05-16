/*
 * FWS/CSL/VauLSMorg 2025
 * CompNetSec 2024-2 A02
 * Don't forget to rename the file to client.go
 */

package main

import (
	"bufio"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("input the url: ")
	rawURL, _ := reader.ReadString('\n')
	rawURL = strings.TrimSpace(rawURL)

	fmt.Print("input the data type: ")
	dtype, _ := reader.ReadString('\n')
	dtype = strings.TrimSpace(strings.ToLower(dtype))

	accept := dtype
	if accept != "application/json" && accept != "application/xml" && accept != "text/html" {
		fmt.Println("unsupported data type")
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("invalid URL:", err)
		return
	}
	h := u.Hostname()
	p := u.Port()
	if p == "" {
		p = "80"
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(h, p))
	if err != nil {
		fmt.Println("connection failed:", err)
		return
	}
	defer conn.Close()

	path := u.RequestURI()
	if path == "" {
		path = "/"
	}

	// Send request
	req := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nAccept: %s\r\nConnection: close\r\n\r\n", path, u.Hostname(), accept)
	conn.Write([]byte(req))

	// Read response
	respBytes, _ := io.ReadAll(conn)
	parts := strings.SplitN(string(respBytes), "\r\n\r\n", 2)
	headLines := strings.Split(parts[0], "\r\n")
	body := ""
	if len(parts) == 2 {
		body = parts[1]
	}

	// Parse status code
	status := strings.Fields(headLines[0])
	code := ""
	if len(status) >= 2 {
		code = status[1]
	}

	// Output
	fmt.Println("Status Code:", code)
	fmt.Println("Body:", body)

	// Parsed for JSON/XML
	ct := ""
	for _, h := range headLines[1:] {
		if strings.HasPrefix(strings.ToLower(h), "content-type:") {
			ct = strings.TrimSpace(strings.SplitN(h, ":", 2)[1])
		}
	}

	switch {
	case strings.Contains(ct, "application/json"):
		var obj interface{}
		json.Unmarshal([]byte(body), &obj)
		parsed := flattenJSON(obj)
		fmt.Println("Parsed:", parsed)
	case strings.Contains(ct, "application/xml"):
		parsed := flattenXML([]byte(body))
		fmt.Println("Parsed:", parsed)
	}
}

// flattenJSON and flattenXML implementations follow your template logic
// (retain your existing functions here)

func flattenJSON(v interface{}) []string {
	var out []string
	var walk func(prefix string, val interface{})
	walk = func(prefix string, val interface{}) {
		switch vv := val.(type) {
		case map[string]interface{}:
			for k, v2 := range vv {
				walk(prefix+"."+k, v2)
			}
		case []interface{}:
			for i, v2 := range vv {
				walk(fmt.Sprintf("%s[%d]", prefix, i), v2)
			}
		default:
			out = append(out, fmt.Sprintf("%s: %v", prefix, vv))
		}
	}
	walk("response", v)
	return out
}

func flattenXML(data []byte) []string {
	type Any struct {
		XMLName xml.Name
		Content []byte   `xml:",innerxml"`
		Nodes   []Any    `xml:",any"`
	}
	var root Any
	xml.Unmarshal(data, &root)
	var out []string
	var walk func(prefix string, node Any)
	walk = func(prefix string, node Any) {
		if len(node.Nodes) == 0 {
			out = append(out, fmt.Sprintf("%s: %s", prefix, strings.TrimSpace(string(node.Content))))
			return
		}
		for _, child := range node.Nodes {
			walk(prefix+"."+child.XMLName.Local, child)
		}
	}
	walk("response", root)
	return out
}
