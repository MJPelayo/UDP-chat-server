package main

import (
	"fmt" // For formatted I/O
	"os"  // For OS operations
)

// main is the entry point of the application
func main() {
	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  Server: go run . server")
		fmt.Println("  Client: go run . client <server-address> <username>")
		return
	}

	mode := os.Args[1] // First argument is mode (server/client)
	switch mode {
	case "server":
		startServer() // Start in server mode
	case "client":
		// Client mode requires additional arguments
		if len(os.Args) < 4 {
			fmt.Println("Client usage: go run . client <server-address> <username>")
			return
		}
		// Start client with provided server address and username
		startClient(os.Args[2], os.Args[3])
	default:
		fmt.Println("Invalid mode. Use 'server' or 'client'")
	}
}
