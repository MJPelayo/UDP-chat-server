package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

const (
	testMessages = 1000 // Number of messages to send
	message      = "test message for benchmarking"
)

func BenchmarkUDPPerformance(b *testing.B) {
	// Start UDP server
	udpServer := newServer()
	go udpServer.start(":8080")
	defer close(udpServer.shutdown)
	time.Sleep(100 * time.Millisecond) // Wait for server to start

	// Create client
	client, err := net.Dial("udp", "localhost:8080")
	if err != nil {
		b.Fatalf("Failed to create UDP client: %v", err)
	}
	defer client.Close()

	// Register client
	client.Write([]byte("REGISTER:bench_user"))

	// Warm up
	time.Sleep(50 * time.Millisecond)

	b.ResetTimer() // Start measuring performance from here

	startTime := time.Now()

	// Send 1000 messages
	for i := 0; i < testMessages; i++ {
		_, err := client.Write([]byte(message))
		if err != nil {
			b.Errorf("Failed to send message: %v", err)
		}
	}

	// Calculate metrics
	duration := time.Since(startTime)

	// Report metrics
	b.ReportMetric(float64(testMessages), "msgs")                                    // Total messages
	b.ReportMetric(float64(testMessages)/duration.Seconds(), "msg/s")                // Message rate
	b.ReportMetric(float64(duration.Microseconds())/float64(testMessages), "μs/msg") // Latency per message
}

func BenchmarkTCPPerformance(b *testing.B) {
	// Start TCP server
	tcpServer := newServer()
	go tcpServer.start(":8081")
	defer close(tcpServer.shutdown)
	time.Sleep(100 * time.Millisecond) // Wait for server to start

	// Create client
	client, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		b.Fatalf("Failed to create TCP client: %v", err)
	}
	defer client.Close()

	// Register client
	fmt.Fprintf(client, "REGISTER:bench_user\n")

	// Warm up
	time.Sleep(50 * time.Millisecond)

	b.ResetTimer() // Start measuring performance from here

	startTime := time.Now()

	// Send 1000 messages
	for i := 0; i < testMessages; i++ {
		_, err := fmt.Fprintf(client, message+"\n")
		if err != nil {
			b.Errorf("Failed to send message: %v", err)
		}
	}

	// Calculate metrics
	duration := time.Since(startTime)

	// Report metrics
	b.ReportMetric(float64(testMessages), "msgs")                                    // Total messages
	b.ReportMetric(float64(testMessages)/duration.Seconds(), "msg/s")                // Message rate
	b.ReportMetric(float64(duration.Microseconds())/float64(testMessages), "μs/msg") // Latency per message
}
