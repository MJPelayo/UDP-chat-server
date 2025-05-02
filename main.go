// main.go
package main

import (
	"fmt"
	"os"
) // NEW: Now handles command-line arguments

func main() {
	// Check for required arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run . server|client <address>")
		return
	}

	// NEW: Supports both server and client modes
	switch os.Args[1] {
	case "server":
		startServer()
	case "client":
		// NEW: Client requires server address
		if len(os.Args) < 3 {
			fmt.Println("Client needs server address")
			return
		}
		startClient(os.Args[2]) // Pass server address
	}
}
