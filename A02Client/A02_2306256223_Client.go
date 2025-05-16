/*
 * FWS/CSL/VauLSMorg 2025
 * CompNetSec 2024-2 A02
 * Don't forget to rename the file to A02_2306256223_Client.go
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

// ─── MAIN ─────────────────────────────────────────────────────────────

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("input the url: ")
	rawURL, _ := reader.ReadString('\n')
	rawURL = strings.TrimSpace(rawURL)

	fmt.Print("input the data type: ")
	dataType, _ := reader.ReadString('\n')
	dataType = strings.TrimSpace(strings.ToLower(dataType))

	acceptHeader := getAcceptHeader(dataType)
	if acceptHeader == "" {
		fmt.Println("unsupported data type:", dataType)
		return
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("invalid URL:", err)
		return
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "80"
	}

	addr := net.JoinHostPort(host, port)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("connection failed:", err)
		return
	}
	defer conn.Close()

	path := u.RequestURI()
	if path == "" {
		path = "/"
	}

	// ─── Send HTTP Request ─────────────────────────────────────────────
	request := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nAccept: %s\r\nConnection: close\r\n\r\n",
		path, u.Hostname(), acceptHeader)

	_, err = conn.Write([]byte(request))
	if err != nil {
		fmt.Println("write failed:", err)
		return
	}

	// ─── Read Response ────────────────────────────────────────────────
	responseBytes, err := io.ReadAll(conn)
	if err != nil {
		fmt.Println("read failed:", err)
		return
	}
	response := string(responseBytes)

	// ─── Split Headers & Body ─────────────────────────────────────────
	parts := strings.SplitN(response, "\r\n\r\n", 2)
	if len(parts) < 2 {
		fmt.Println("invalid response")
		return
	}

	headerLines := strings.Split(parts[0], "\r\n")
	body := parts[1]
	status := headerLines[0]
	fmt.Println("STATUS:", status)

	contentType := ""
	for _, line := range headerLines[1:] {
		if strings.HasPrefix(strings.ToLower(line), "content-type:") {
			contentType = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
		}
	}

	// ─── Process Response Body ────────────────────────────────────────
	switch {
	case strings.Contains(contentType, "application/json"):
		var obj interface{}
		if err := json.Unmarshal([]byte(body), &obj); err != nil {
			fmt.Println("invalid JSON:", err)
			return
		}
		for _, line := range flattenJSON(obj) {
			fmt.Println(line)
		}
	case strings.Contains(contentType, "application/xml"):
		for _, line := range flattenXML([]byte(body)) {
			fmt.Println(line)
		}
	default:
		fmt.Println(body)
	}
}

// ─── HELPERS ─────────────────────────────────────────────────────────

// getAcceptHeader returns the appropriate Accept header for input type
func getAcceptHeader(dataType string) string {
	switch dataType {
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	case "text/html", "html":
		return "text/html"
	default:
		return ""
	}
}

// flattenJSON flattens nested JSON into lines of dot notation
func flattenJSON(v interface{}) []string {
	var out []string
	var walk func(prefix string, val interface{})
	walk = func(prefix string, val interface{}) {
		switch val := val.(type) {
		case map[string]interface{}:
			for k, v2 := range val {
				walk(prefix+"."+k, v2)
			}
		case []interface{}:
			for i, v2 := range val {
				walk(fmt.Sprintf("%s[%d]", prefix, i), v2)
			}
		default:
			out = append(out, fmt.Sprintf("%s: %v", prefix, val))
		}
	}
	walk("response", v)
	return out
}

// flattenXML flattens XML recursively into lines
func flattenXML(data []byte) []string {
	type AnyXML struct {
		XMLName xml.Name
		Content []byte   `xml:",innerxml"`
		Nodes   []AnyXML `xml:",any"`
	}

	var root AnyXML
	xml.Unmarshal(data, &root)

	var out []string
	var walk func(prefix string, node AnyXML)
	walk = func(prefix string, node AnyXML) {
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
