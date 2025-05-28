package testutil

import (
	"bufio"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/keybittech/awayto-v3/go/pkg/util"
)

func GetSocketTicket(client *http.Client, userId string) (string, string) {
	req, err := http.NewRequest("GET", util.E_APP_HOST_URL+"/api/v1/sock/ticket", nil)
	if err != nil {
		log.Fatalf("could not make ticket request %v", err)
	}

	body, err := doAndRead(client, req)
	if err != nil {
		log.Fatal("failed get socket ticket", err)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatal("failed marshal ticket body", err)
	}

	ticket, ok := result["ticket"].(string)
	if !ok {
		log.Fatal("ticket not found in response")
	}

	ticketParts := strings.Split(ticket, ":")
	_, connId := ticketParts[0], ticketParts[1]

	return ticket, connId
}

func GetClientSocketConnection(ticket string) (net.Conn, error) {
	u, err := url.Parse("wss://" + util.E_APP_HOST_NAME + "/sock?ticket=" + ticket)
	if err != nil {
		return nil, err
	}

	sockConn, err := tls.Dial("tcp", u.Host, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}

	// Generate WebSocket key
	keyBytes := make([]byte, 16)
	rand.Read(keyBytes)
	secWebSocketKey := base64.StdEncoding.EncodeToString(keyBytes)

	// Send HTTP Upgrade request
	fmt.Fprintf(sockConn, "GET %s HTTP/1.1\r\n", u.RequestURI())
	fmt.Fprintf(sockConn, "Host: %s\r\n", u.Host)
	fmt.Fprintf(sockConn, "Upgrade: websocket\r\n")
	fmt.Fprintf(sockConn, "Connection: Upgrade\r\n")
	fmt.Fprintf(sockConn, "Sec-WebSocket-Key: %s\r\n", secWebSocketKey)
	fmt.Fprintf(sockConn, "Sec-WebSocket-Version: 13\r\n")
	fmt.Fprintf(sockConn, "\r\n")

	reader := bufio.NewReader(sockConn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if line == "\r\n" {
			break
		}
	}

	return sockConn, nil
}
