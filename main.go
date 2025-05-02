// main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  Server: go run . server")
		fmt.Println("  Client: go run . client <address> <username>")
		return
	}

	switch os.Args[1] {
	case "server":
		startServer()
	case "client":
		if len(os.Args) < 4 {
			fmt.Println("Client needs server address and username")
			return
		}
		startClient(os.Args[2], os.Args[3])
	}
}
