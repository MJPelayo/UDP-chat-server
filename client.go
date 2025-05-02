// client.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings" // For parsing whisper command
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

	// NEW: Show admin commands if username is "admin"
	if username == "admin" {
		fmt.Println("Admin commands: /menu, /kick")
	}

	fmt.Println("Commands:")
	fmt.Println("/users - List online users")
	fmt.Println("/quit - Exit the chat")
	fmt.Println("/rename <name> - Change your username")
	fmt.Println("/whisper <user> <msg> - Private message") // NEW: Whisper command
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
		switch {
		case text == "/help":
			// NEW: Pass admin status to help function
			printHelp(username == "admin")
		case text == "/quit":
			conn.Write([]byte("/quit"))
			return
		// NEW: Handle whisper command
		case strings.HasPrefix(text, "/whisper "):
			parts := strings.SplitN(strings.TrimPrefix(text, "/whisper "), " ", 2)
			if len(parts) == 2 {
				// Format: WHISPER:target:message
				conn.Write([]byte(fmt.Sprintf("WHISPER:%s:%s", parts[0], parts[1])))
			}
		default:
			conn.Write([]byte(text))
		}
	}
}

// NEW: Updated help to show admin commands conditionally
func printHelp(isAdmin bool) {
	help := `Available commands:
/users    - List online users
/quit     - Exit the chat
/rename   - Change your username
/whisper  - Send private message
/help     - Show this help`

	if isAdmin {
		help += `
Admin commands:
/kick <user> - Kick a user
/menu - Admin menu`
	}

	fmt.Println(help)
}
