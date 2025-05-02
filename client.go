// client.go
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func startClient(serverAddr, username string) {
	addr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		log.Fatal("Resolve address error:", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatal("Connection error:", err)
	}
	defer conn.Close()

	// Set initial read timeout
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	// Register with server
	_, err = conn.Write([]byte("REGISTER:" + username))
	if err != nil {
		log.Fatal("Registration failed:", err)
	}

	// Color-coded connection message
	fmt.Printf("\033[32mConnected to %s as %s\033[0m\n", serverAddr, username)
	if username == "admin" {
		fmt.Println("Type /menu for admin commands")
	} else {
		fmt.Println("Type /help for commands")
		fmt.Println("You will be automatically disconnected after 10 minutes of inactivity")
	}

	// Use WaitGroup to manage goroutines
	var wg sync.WaitGroup
	wg.Add(2)                       // We'll launch 2 goroutines
	shutdown := make(chan struct{}) // Channel for graceful shutdown

	// Goroutine 1: Handle incoming messages
	go func() {
		defer wg.Done() // Notify WaitGroup when done
		buf := make([]byte, 1024)
		for {
			select {
			case <-shutdown:
				return
			default:
				// Refresh read timeout periodically
				conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				n, err := conn.Read(buf)
				if err != nil {
					// Handle timeout errors differently
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					log.Println("Receive error:", err)
					close(shutdown)
					return
				}
				// Clear line and print message
				fmt.Print("\r\033[K")
				fmt.Printf("%s\n", string(buf[:n]))
				showPrompt(username)
			}
		}
	}()

	// Goroutine 2: Handle user input
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(os.Stdin)
		for {
			select {
			case <-shutdown:
				return
			default:
				showPrompt(username)
				if !scanner.Scan() {
					close(shutdown)
					return
				}

				text := scanner.Text()
				// Enhanced command handling
				switch {
				case text == "/quit":
					conn.Write([]byte("QUIT:" + username))
					close(shutdown)
					return
				case text == "/help":
					printHelp(username == "admin")
				case text == "/menu" && username == "admin":
					conn.Write([]byte("/menu"))
				case strings.HasPrefix(text, "/rename "):
					newName := strings.TrimPrefix(text, "/rename ")
					conn.Write([]byte("RENAME:" + newName))
				case text == "/users":
					conn.Write([]byte("/users"))
				case text == "/stats":
					conn.Write([]byte("/stats"))
				case strings.HasPrefix(text, "/whisper "):
					parts := strings.SplitN(strings.TrimPrefix(text, "/whisper "), " ", 2)
					if len(parts) == 2 {
						conn.Write([]byte("WHISPER:" + parts[0] + ":" + parts[1]))
					} else {
						fmt.Println("\033[31mUsage: /whisper username message\033[0m")
					}
				case strings.HasPrefix(text, "/kick ") && username == "admin":
					target := strings.TrimPrefix(text, "/kick ")
					conn.Write([]byte("KICK:" + target))
				case strings.HasPrefix(text, "/broadcast ") && username == "admin":
					msg := strings.TrimPrefix(text, "/broadcast ")
					conn.Write([]byte("BROADCAST:" + msg))
				case text == "/shutdown" && username == "admin":
					conn.Write([]byte("SHUTDOWN:"))
					close(shutdown)
					return
				default:
					if strings.HasPrefix(text, "/") {
						fmt.Println("\033[31mInvalid command. Type /help for available commands\033[0m")
					} else {
						conn.Write([]byte(text))
					}
				}
			}
		}
	}()

	// Wait for both goroutines to finish
	wg.Wait()
	fmt.Println("\033[33mDisconnected from server\033[0m")
}

// Helper function to show username prompt
func showPrompt(username string) {
	fmt.Printf("\033[35m[%s]\033[0m > ", username)
}

// Enhanced help with color and admin commands
func printHelp(isAdmin bool) {
	help := `
\033[1mAvailable Commands:\033[0m
/help               - Show this help message
/users              - List online users
/stats              - Show server statistics
/quit               - Exit the chat
/rename <newname>   - Change your username
/whisper <user> <msg> - Send private message
`

	if isAdmin {
		help += `
\033[1mAdmin Commands:\033[0m
/menu               - Show admin menu
/kick <username>    - Remove a user
/broadcast <msg>    - Server announcement
/shutdown           - Shutdown server
`
	}
	fmt.Println(help)
}
