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
	"net/url"
	"os"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	// TODO: Get URL and accept type from user input
	fmt.Print("input the url: ")
	rawURL, _ := reader.ReadString('\n')
	rawURL = strings.TrimSpace(rawURL)

	fmt.Print("input the data type: ")
	accept, _ := reader.ReadString('\n')
	accept = strings.TrimSpace(accept)

	// TODO: Parse URL and extract host/port
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		fmt.Println("invalid url")
		return
	}
	host, port, err := splitHostPort(parsedURL.Host)
	if err != nil {
		fmt.Println("invalid host:port")
		return
	}
	ip, err := parseIPv4(host)
	if err != nil {
		fmt.Println("invalid ipv4")
		return
	}
	portNum, _ := strconv.Atoi(port)

	// TODO: Create socket and establish connection
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		fmt.Println("socket error:", err)
		return
	}
	defer syscall.Close(fd)

	sa := &syscall.SockaddrInet4{Port: portNum, Addr: ip}
	if err := syscall.Connect(fd, sa); err != nil {
		fmt.Println("connect error:", err)
		return
	}

	// TODO: Send HTTP request
	request := fmt.Sprintf("GET %s HTTP/1.1\r\n", parsedURL.RequestURI())
	request += fmt.Sprintf("Host: %s\r\n", parsedURL.Host)
	request += fmt.Sprintf("Accept: %s\r\n", accept)
	request += "Connection: close\r\n\r\n"

	syscall.Write(fd, []byte(request))

	// TODO: Read HTTP response (status, headers, body)
	var buf [4096]byte
	n, _ := syscall.Read(fd, buf[:])
	resp := string(buf[:n])

	parts := strings.SplitN(resp, "\r\n\r\n", 2)
	if len(parts) != 2 {
		fmt.Println("invalid response")
		return
	}
	headerPart := parts[0]
	body := parts[1]

	// TODO: Parse and display response based on content type
	headers := strings.Split(headerPart, "\r\n")
	contentType := ""
	for _, h := range headers {
		if strings.HasPrefix(strings.ToLower(h), "content-type:") {
			contentType = strings.TrimSpace(h[len("content-type:"):])
			break
		}
	}

	if strings.Contains(contentType, "json") {
		var data interface{}
		if err := json.Unmarshal([]byte(body), &data); err != nil {
			fmt.Println("invalid json")
			return
		}
		for _, line := range flattenJSON(data) {
			fmt.Println(line)
		}
	} else if strings.Contains(contentType, "xml") {
		lines := flattenXML([]byte(body))
		for _, line := range lines {
			fmt.Println(line)
		}
	} else {
		fmt.Println(body)
	}
}

// helper to split "host:port" (default port 80)
func splitHostPort(h string) (host, port string, _ error) {
	// TODO: Implement host:port splitting
	if strings.Contains(h, ":") {
		parts := strings.Split(h, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid hostport")
		}
		return parts[0], parts[1], nil
	}
	return h, "80", nil
}

// only dotted IPv4
func parseIPv4(s string) ([4]byte, error) {
	// TODO: Implement IPv4 parsing
	var b [4]byte
	parts := strings.Split(s, ".")
	if len(parts) != 4 {
		return b, fmt.Errorf("invalid ipv4")
	}
	for i := 0; i < 4; i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil || n < 0 || n > 255 {
			return b, fmt.Errorf("invalid segment")
		}
		b[i] = byte(n)
	}
	return b, nil
}

func nonEmpty(s, d string) string {
	// TODO: Implement non-empty string check
	if s != "" {
		return s
	}
	return d
}

func flattenJSON(v interface{}) []string {
	// TODO: Implement JSON flattening
	var out []string
	var f func(interface{}, string)
	f = func(val interface{}, prefix string) {
		switch vv := val.(type) {
		case map[string]interface{}:
			for k, v2 := range vv {
				f(v2, prefix+k+".")
			}
		case []interface{}:
			for i, v2 := range vv {
				f(v2, fmt.Sprintf("%s%d.", prefix, i))
			}
		default:
			out = append(out, fmt.Sprintf("%s%v", prefix, vv))
		}
	}
	f(v, "")
	return out
}

func flattenXML(data []byte) []string {
	// TODO: Implement XML flattening
	var out []string
	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var stack []string
	for {
		t, err := decoder.Token()
		if err != nil {
			break
		}
		switch tok := t.(type) {
		case xml.StartElement:
			stack = append(stack, tok.Name.Local)
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			s := strings.TrimSpace(string(tok))
			if s != "" {
				out = append(out, fmt.Sprintf("%s: %s", strings.Join(stack, "."), s))
			}
		}
	}
	return out
}
