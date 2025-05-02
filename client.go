// client.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

// NEW: Now takes username as second parameter
func startClient(serverAddr, username string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// NEW: Register username with server
	conn.Write([]byte("REGISTER:" + username))
	fmt.Printf("Connected as %s\n", username)

	// Goroutine to handle incoming messages
	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("Receive error:", err)
				return
			}
			// Messages now include usernames
			fmt.Println(string(buf[:n]))
		}
	}()

	// Read user input continuously
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		conn.Write([]byte(scanner.Text()))
	}
}
