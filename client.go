// client.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	// NEW: For parsing commands
)

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

	conn.Write([]byte("REGISTER:" + username))
	fmt.Printf("Connected as %s\n", username)

	// NEW: Show available commands on startup
	fmt.Println("Commands:")
	fmt.Println("/users - List online users")
	fmt.Println("/quit - Exit the chat")
	fmt.Println("/rename <name> - Change your username")
	fmt.Println("/help - Show this help")

	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Println("Receive error:", err)
				return
			}
			fmt.Println(string(buf[:n]))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		// NEW: Client-side command handling
		switch {
		case text == "/help":
			printHelp() // Show help menu
		case text == "/quit":
			conn.Write([]byte("/quit")) // Send quit command
			return                      // Exit client
		default:
			conn.Write([]byte(text)) // Send normal message
		}
	}
}

// NEW: Help function to explain commands
func printHelp() {
	help := `Available commands:
/users    - List online users
/quit     - Exit the chat
/rename   - Change your username
/help     - Show this help`
	fmt.Println(help)
}
