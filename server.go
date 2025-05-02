package main

import (
	"log" // For basic error logging
	"net" // For core networking functions
)

func startServer() {
	// Create a UDP server listening on port 8080
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8080})
	if err != nil {
		log.Fatal(err) // Crash if server can't start
	}
	defer conn.Close() // Make sure connection closes when done

	buf := make([]byte, 1024) // Buffer to hold incoming messages

	// Infinite loop to keep server running
	for {
		// Wait for and read an incoming message
		// n = number of bytes, addr = sender address
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err) // Log errors but keep running
			continue
		}

		// Echo the same message back to sender
		// buf[:n] = only the received bytes (not entire buffer)
		conn.WriteToUDP(buf[:n], addr)
	}
}
