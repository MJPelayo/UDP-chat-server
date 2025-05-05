package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  Server: go run . server")
		fmt.Println("  Client: go run . client <server-address> <username>")
		return
	}

	mode := os.Args[1]
	switch mode {
	case "server":
		startServer()
	case "client":
		if len(os.Args) < 4 {
			fmt.Println("Client usage: go run . client <server-address> <username>")
			return
		}
		startClient(os.Args[2], os.Args[3])
	default:
		fmt.Println("Invalid mode. Use 'server' or 'client'")
	}
}
