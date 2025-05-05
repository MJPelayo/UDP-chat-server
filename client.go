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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func showInteractiveHelp(isAdmin bool) string {
	clearScreen()
	fmt.Println("\033[1;36m┌──────────────────────────────────────┐")
	fmt.Println("│           \033[1;35mGOCHAT HELP\033[1;36m           │")
	fmt.Println("├──────────────────────────────────────┤")

	options := []struct {
		key, desc string
	}{
		{"1", "List online users"},
		{"2", "Send private message"},
		{"3", "Change username"},
		{"4", "View server stats"},
	}

	for _, opt := range options {
		fmt.Printf("│ \033[32m%s\033[0m - %-25s \033[36m│\n", opt.key, opt.desc)
	}

	if isAdmin {
		fmt.Println("├──────────────────────────────────────┤")
		adminOpts := []struct {
			key, desc string
		}{
			{"5", "Kick user"},
			{"6", "Server broadcast"},
			{"7", "View admin menu"},
		}
		for _, opt := range adminOpts {
			fmt.Printf("│ \033[31m%s\033[0m - %-25s \033[36m│\n", opt.key, opt.desc)
		}
	}

	fmt.Println("└──────────────────────────────────────┘\033[0m")
	fmt.Print("Select option (q to quit): ")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		choice := scanner.Text()
		switch choice {
		case "1":
			return "/users"
		case "2":
			fmt.Print("Enter username: ")
			scanner.Scan()
			user := scanner.Text()
			fmt.Print("Enter message: ")
			scanner.Scan()
			msg := scanner.Text()
			return fmt.Sprintf("WHISPER:%s:%s", user, msg)
		case "3":
			fmt.Print("Enter new username: ")
			scanner.Scan()
			return fmt.Sprintf("RENAME:%s", scanner.Text())
		case "4":
			return "/stats"
		case "5":
			if isAdmin {
				fmt.Print("Enter username to kick: ")
				scanner.Scan()
				return fmt.Sprintf("KICK:%s", scanner.Text())
			}
		case "6":
			if isAdmin {
				fmt.Print("Enter broadcast message: ")
				scanner.Scan()
				return fmt.Sprintf("BROADCAST:%s", scanner.Text())
			}
		case "7":
			if isAdmin {
				return "/menu"
			}
		case "q":
			return ""
		default:
			fmt.Print("Invalid choice, try again: ")
		}
	}
	return ""
}

func showPrompt(username string) {
	fmt.Printf("\033[35m[%s]\033[0m > ", username)
}

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

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	_, err = conn.Write([]byte("REGISTER:" + username))
	if err != nil {
		log.Fatal("Registration failed:", err)
	}

	fmt.Printf("\033[32mConnected to %s as %s\033[0m\n", serverAddr, username)
	if username == "admin" {
		fmt.Println("Type /menu for admin commands")
	} else {
		fmt.Println("Type /help for commands")
		fmt.Println("You will be automatically disconnected after 10 minutes of inactivity")
	}

	var wg sync.WaitGroup
	wg.Add(2)
	shutdown := make(chan struct{})

	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		for {
			select {
			case <-shutdown:
				return
			default:
				conn.SetReadDeadline(time.Now().Add(30 * time.Second))
				n, err := conn.Read(buf)
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					log.Println("Receive error:", err)
					close(shutdown)
					return
				}
				fmt.Print("\r\033[K")
				fmt.Printf("%s\n", string(buf[:n]))
				showPrompt(username)
			}
		}
	}()

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
				switch {
				case text == "/quit":
					conn.Write([]byte("QUIT:" + username))
					close(shutdown)
					return
				case text == "/help":
					cmd := showInteractiveHelp(username == "admin")
					if cmd != "" {
						conn.Write([]byte(cmd))
					}
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

	wg.Wait()
	fmt.Println("\033[33mDisconnected from server\033[0m")
}

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
