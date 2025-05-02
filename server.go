// server.go
package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

// NEW: Added isAdmin flag to Client struct
type Client struct {
	addr     *net.UDPAddr
	name     string
	lastSeen time.Time
	isAdmin  bool // Flag for admin privileges
}

var clients = make(map[string]*Client)
var mu sync.Mutex

func startServer() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 8080})
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println(err)
			continue
		}

		msg := string(buf[:n])
		mu.Lock()

		if client, exists := clients[addr.String()]; exists {
			client.lastSeen = time.Now()

			switch {
			case msg == "/users":
				list := "Connected users:\n"
				for _, c := range clients {
					// NEW: Show admin status in user list
					adminTag := ""
					if c.isAdmin {
						adminTag = " (admin)"
					}
					list += fmt.Sprintf("- %s%s\n", c.name, adminTag)
				}
				conn.WriteToUDP([]byte(list), addr)

			case msg == "/quit":
				delete(clients, addr.String())
				broadcast(conn, fmt.Sprintf("%s left", client.name))

			case strings.HasPrefix(msg, "/rename "):
				newName := strings.TrimPrefix(msg, "/rename ")
				// NEW: Update admin status when renaming
				client.name = newName
				client.isAdmin = (newName == "admin") // "admin" username gets privileges
				broadcast(conn, fmt.Sprintf("%s changed name to %s", client.name, newName))

			// NEW: Private message handling
			case strings.HasPrefix(msg, "WHISPER:"):
				parts := strings.SplitN(strings.TrimPrefix(msg, "WHISPER:"), ":", 2)
				if len(parts) == 2 {
					target, message := parts[0], parts[1]
					for _, c := range clients {
						if c.name == target {
							// Send private message only to target user
							conn.WriteToUDP([]byte(fmt.Sprintf("[PM from %s] %s", client.name, message)), c.addr)
						}
					}
				}

			// NEW: Admin commands
			case msg == "/menu" && client.isAdmin:
				adminMenu := "ADMIN MENU:\n/kick <user>\n/shutdown"
				conn.WriteToUDP([]byte(adminMenu), addr)

			case strings.HasPrefix(msg, "/kick ") && client.isAdmin:
				target := strings.TrimPrefix(msg, "/kick ")
				for addrStr, c := range clients {
					if c.name == target {
						delete(clients, addrStr)
						// Notify kicked user
						conn.WriteToUDP([]byte("You were kicked by admin"), c.addr)
					}
				}

			default:
				broadcast(conn, fmt.Sprintf("%s: %s", client.name, msg))
			}
		} else if strings.HasPrefix(msg, "REGISTER:") {
			name := strings.TrimPrefix(msg, "REGISTER:")
			// NEW: Set admin status during registration
			isAdmin := (name == "admin")
			clients[addr.String()] = &Client{
				addr:     addr,
				name:     name,
				lastSeen: time.Now(),
				isAdmin:  isAdmin,
			}
			broadcast(conn, fmt.Sprintf("User %s joined", name))
		}

		mu.Unlock()
	}
}

func broadcast(conn *net.UDPConn, msg string) {
	for _, client := range clients {
		conn.WriteToUDP([]byte(msg), client.addr)
	}
}
