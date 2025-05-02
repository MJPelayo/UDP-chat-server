p// server.go
package main

import (
    "log"
    "net"
    "sync" // NEW: For thread-safe client management
)

// NEW: Track all connected clients by their address string
var clients = make(map[string]*net.UDPAddr)

// NEW: Mutex to protect clients map from concurrent access
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
        
        // NEW: Lock clients map for thread safety
        mu.Lock()
        
        // NEW: Remember new clients
        if _, exists := clients[addr.String()]; !exists {
            clients[addr.String()] = addr
        }
        
        // NEW: Broadcast message to all clients except sender
        msg := string(buf[:n])
        for _, clientAddr := range clients {
            if clientAddr.String() != addr.String() {
                conn.WriteToUDP([]byte(msg), clientAddr)
            }
        }
        
        // NEW: Unlock when done
        mu.Unlock()
    }
}