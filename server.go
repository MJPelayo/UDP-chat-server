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

type Client struct {
	addr     *net.UDPAddr
	name     string
	lastSeen time.Time
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

			// NEW: Command handling system
			switch {
			case msg == "/users":
				// Handle /users command - list all connected users
				list := "Connected users:\n"
				for _, c := range clients {
					list += fmt.Sprintf("- %s\n", c.name)
				}
				conn.WriteToUDP([]byte(list), addr)

			case msg == "/quit":
				// Handle /quit command - remove user gracefully
				delete(clients, addr.String())
				broadcast(conn, fmt.Sprintf("%s left", client.name))

			case strings.HasPrefix(msg, "/rename "):
				// Handle /rename command - change username
				newName := strings.TrimPrefix(msg, "/rename ")
				client.name = newName
				broadcast(conn, fmt.Sprintf("%s changed name to %s", client.name, newName))

			default:
				// Regular message - broadcast normally
				broadcast(conn, fmt.Sprintf("%s: %s", client.name, msg))
			}
		} else if strings.HasPrefix(msg, "REGISTER:") {
			// Existing registration logic
			name := strings.TrimPrefix(msg, "REGISTER:")
			clients[addr.String()] = &Client{
				addr:     addr,
				name:     name,
				lastSeen: time.Now(),
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
