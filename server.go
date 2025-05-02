// server.go
package main

import (
	"fmt"
	"log"
	"net"
	"strings" // NEW: For parsing registration messages
	"sync"
	"time" // NEW: For tracking last activity
)

// NEW: Client struct to store more user information
type Client struct {
	addr     *net.UDPAddr // Network address
	name     string       // Username
	lastSeen time.Time    // Last activity timestamp
}

var clients = make(map[string]*Client) // Now stores Client structs
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

		// NEW: Registration system
		if strings.HasPrefix(msg, "REGISTER:") {
			name := strings.TrimPrefix(msg, "REGISTER:")
			// NEW: Create client with username and timestamp
			clients[addr.String()] = &Client{
				addr:     addr,
				name:     name,
				lastSeen: time.Now(),
			}
			// NEW: Notify all users about new join
			broadcast(conn, fmt.Sprintf("User %s joined", name))
		} else if client, exists := clients[addr.String()]; exists {
			// Existing user - update last seen and handle message
			client.lastSeen = time.Now()
			// NEW: Messages now show username instead of address
			broadcast(conn, fmt.Sprintf("%s: %s", client.name, msg))
		}

		mu.Unlock()
	}
}

// NEW: Helper function for broadcasting to all clients
func broadcast(conn *net.UDPConn, msg string) {
	for _, client := range clients {
		conn.WriteToUDP([]byte(msg), client.addr)
	}
}
