package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"
)

const maxMessageSize = 1000
const clientReadTimeout = time.Second * 5

func main() {
	// Note: This client does not support empty keys even though they are supported by the server

	portPtr := flag.Uint("port", 8000, "specify the port to connect to")
	hostPtr := flag.String("host", "0.0.0.0", "specify the address of remote server")
	flag.Parse()
	address := fmt.Sprintf("%s:%d", *hostPtr, *portPtr)

	ctx := context.Background()
	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "udp", address)
	if err != nil {
		slog.Error("create udp dialer failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Unusual Database Program client\n")
	fmt.Printf("Type \"exit\" to stop the program\n")
	var buffer [maxMessageSize]byte

	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			}
			break
		}
		request := scanner.Text()
		if request == "exit" {
			break
		}
		if request == "" {
			continue
		}
		if len(request) > maxMessageSize {
			fmt.Fprintf(os.Stderr, "Error message too large size: %d, should be less than %d\n", len(request), maxMessageSize)
			continue
		}
		req := []byte(request)
		_, err := conn.Write(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while sending request %q to server: %s\n", request, err)
			continue
		}

		// If the request is a query, wait for a response from the server
		if !strings.ContainsRune(request, '=') {
			conn.SetReadDeadline(time.Now().Add(clientReadTimeout))
			n, err := conn.Read(buffer[:])
			if n > 0 {
				fmt.Printf("%s\n", string(buffer[:n]))
			}
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Timeout() {
					fmt.Fprintf(os.Stderr, "Response timed out\n")
				} else {
					fmt.Fprintf(os.Stderr, "Error occured while waiting for response: %s\n", err)
				}
			}
		}
	}
}
