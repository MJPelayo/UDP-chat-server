// client.go
package main

import (
	"bufio" // NEW: For reading user input
	"fmt"
	"log"
	"net"
	"os"
)

// NEW: Now takes server address as parameter
func startClient(serverAddr string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// NEW: Goroutine to handle incoming messages concurrently
	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("Receive error:", err)
				return
			}
			// NEW: Print incoming messages with prefix
			fmt.Println(">", string(buf[:n]))
		}
	}()

	// NEW: Continuously read user input
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		// NEW: Send each line of input to server
		conn.Write([]byte(scanner.Text()))
	}
}
