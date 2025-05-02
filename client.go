// client.go
package main

import (
	"fmt" // For printing output
	"log" // For error handling
	"net" // For networking
	// For command-line args (not used yet)
)

func startClient() {
	// Connect to localhost on port 8080
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal(err) // Crash if address is invalid
	}

	// Establish UDP connection
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err) // Crash if connection fails
	}
	defer conn.Close() // Ensure connection closes when done

	// Send a single test message
	conn.Write([]byte("Hello, server!"))

	// Prepare to receive response
	buf := make([]byte, 1024) // Buffer for server response

	// Wait for and read server's echo response
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		log.Fatal(err) // Crash if read fails
	}

	// Print the echoed message
	fmt.Println("Server reply:", string(buf[:n]))
}
