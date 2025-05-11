package main

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkUDPChat(b *testing.B) {
	// Start UDP server (using your UDP server implementation)
	server := newServer()
	go server.start(":8080")

	// Allow server to start
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		// Start UDP client
		go startClient("localhost:8080", fmt.Sprintf("user%d", i))
		time.Sleep(10 * time.Millisecond) // Stagger connections
	}

	// Run benchmark for 5 seconds
	time.Sleep(5 * time.Second)
}

func BenchmarkTCPChat(b *testing.B) {
	// Start TCP server (using your TCP server implementation)
	server := newServer()
	go server.start(":8081") // Different port to avoid conflict

	// Allow server to start
	time.Sleep(100 * time.Millisecond)

	for i := 0; i < b.N; i++ {
		// Start TCP client
		go startClient("localhost:8081", fmt.Sprintf("user%d", i))
		time.Sleep(10 * time.Millisecond) // Stagger connections
	}

	// Run benchmark for 5 seconds
	time.Sleep(5 * time.Second)
}
