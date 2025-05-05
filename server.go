package main

import (
	"bufio"   // For reading input from console
	"fmt"     // For formatted I/O
	"log"     // For logging errors
	"net"     // For network operations
	"os"      // For OS operations
	"strings" // For string manipulation
	"sync"    // For synchronization
	"time"    // For time operations
)

// Client represents a connected chat client
type Client struct {
	addr     *net.UDPAddr // Network address of client
	name     string       // Username
	lastSeen time.Time    // Last activity timestamp
	isAdmin  bool         // Admin privileges flag
}

// Server manages the chat server state
type Server struct {
	clients   map[string]*Client // Map of connected clients (key: address string)
	mu        sync.RWMutex       // Mutex for thread-safe client access
	messages  chan string        // Channel for broadcasting messages
	startTime time.Time          // Server start time
	shutdown  chan struct{}      // Channel for graceful shutdown
}

// newServer creates and initializes a new Server instance
func newServer() *Server {
	return &Server{
		clients:   make(map[string]*Client), // Initialize empty client map
		messages:  make(chan string, 100),   // Buffered message channel
		startTime: time.Now(),               // Set current time as start time
		shutdown:  make(chan struct{}),      // Initialize shutdown channel
	}
}

// formatMessage formats a chat message with timestamp and username
func (s *Server) formatMessage(client *Client, msg string) string {
	timestamp := time.Now().Format("15:04") // Format time as HH:MM
	username := client.name
	if client.isAdmin {
		username = "👑 " + username // Add crown emoji for admins
	}

	// Format: [time] username │ message (with colors)
	return fmt.Sprintf(
		"\033[90m%s\033[0m \033[36m%-15s\033[0m │ %s",
		timestamp,
		username,
		msg,
	)
}

// start begins the UDP server on specified port
func (s *Server) start(port string) {
	// Resolve UDP address
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatal("Resolve address error:", err)
	}

	// Create UDP listener
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer conn.Close() // Ensure connection closes when function exits

	log.Printf("Server started on %s", port)

	// Start message broadcaster goroutine
	go s.broadcastMessages(conn)
	// Start client cleanup goroutine
	go s.cleanupClients()

	buf := make([]byte, 1024) // Buffer for incoming messages
	for {
		select {
		case <-s.shutdown: // Shutdown signal received
			log.Println("Shutting down server...")
			s.mu.RLock() // Read lock for clients map
			// Notify all clients of shutdown
			for _, client := range s.clients {
				conn.WriteToUDP([]byte("\033[31mServer is shutting down. Goodbye!\033[0m\n"), client.addr)
			}
			s.mu.RUnlock()
			return // Exit server loop

		default: // Normal operation
			// Set read timeout to 1 minute
			conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
			// Read incoming UDP packet
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue // Timeout is normal, continue waiting
				}
				log.Printf("Read error: %v", err)
				continue
			}
			// Handle message in new goroutine
			go s.handleMessage(conn, clientAddr, string(buf[:n]))
		}
	}
}

// handleMessage processes incoming messages from clients
func (s *Server) handleMessage(conn *net.UDPConn, addr *net.UDPAddr, msg string) {
	s.mu.Lock()         // Acquire write lock
	defer s.mu.Unlock() // Ensure lock is released

	// Handle typing indicator
	if strings.HasPrefix(msg, "TYPING:") {
		name := strings.TrimPrefix(msg, "TYPING:")
		s.messages <- fmt.Sprintf("\033[2m%s is typing...\033[0m", name)
		return
	}

	clientKey := addr.String() // Use address string as client key
	client, exists := s.clients[clientKey]

	if !exists { // New client registration
		if strings.HasPrefix(msg, "REGISTER:") {
			name := strings.TrimPrefix(msg, "REGISTER:")
			// Check for duplicate usernames
			for _, c := range s.clients {
				if c.name == name {
					conn.WriteToUDP([]byte("\033[31mUsername already taken. Please choose another.\033[0m\n"), addr)
					return
				}
			}

			isAdmin := (name == "admin") // "admin" gets special privileges
			newClient := &Client{
				addr:     addr,
				name:     name,
				lastSeen: time.Now(),
				isAdmin:  isAdmin,
			}
			s.clients[clientKey] = newClient // Add new client

			// Format and broadcast join notification
			welcome := s.formatMessage(newClient, "joined the chat")
			s.messages <- "\033[32m" + welcome + "\033[0m"

			if isAdmin { // Send admin menu if admin
				adminMsg := "\n\033[33mADMIN MENU:\n" +
					"1. /users - List all users\n" +
					"2. /kick <username> - Kick a user\n" +
					"3. /broadcast <msg> - Server announcement\n" +
					"4. /stats - Server statistics\n" +
					"5. /shutdown - Shutdown server\033[0m\n"
				conn.WriteToUDP([]byte(adminMsg), addr)
			}
		}
		return
	}

	// Update last seen time for existing client
	client.lastSeen = time.Now()

	// Handle different command types
	switch {
	case msg == "/menu" && client.isAdmin:
		// Show admin menu
		adminMenu := "\n\033[33mADMIN MENU:\n" +
			"1. /users - List all users\n" +
			"2. /kick <username> - Kick a user\n" +
			"3. /broadcast <msg> - Server announcement\n" +
			"4. /stats - Server statistics\n" +
			"5. /shutdown - Shutdown server\033[0m\n"
		conn.WriteToUDP([]byte(adminMenu), addr)

	case msg == "/users":
		// List all connected users
		userList := "\033[1mConnected users:\033[0m\n"
		for _, c := range s.clients {
			adminTag := ""
			if c.isAdmin {
				adminTag = " (admin)"
			}
			userList += fmt.Sprintf("- %s%s\n", c.name, adminTag)
		}
		conn.WriteToUDP([]byte(userList), addr)

	case msg == "/help":
		// Show help message
		help := "\033[1mCommands:\033[0m\n" +
			"/users - List online users\n" +
			"/help - Show this help\n" +
			"/stats - Show server statistics\n" +
			"/quit - Disconnect from server\n"
		conn.WriteToUDP([]byte(help), addr)

	case msg == "/stats":
		// Show server statistics
		uptime := time.Since(s.startTime).Round(time.Second)
		stats := fmt.Sprintf("\033[1mServer Stats:\033[0m\n"+
			"Uptime: %s\n"+
			"Users connected: %d\n"+
			"Timeout: 10 minutes (except admins)\n", uptime, len(s.clients))
		conn.WriteToUDP([]byte(stats), addr)

	case strings.HasPrefix(msg, "RENAME:"):
		// Handle username change
		newName := strings.TrimPrefix(msg, "RENAME:")
		// Check for duplicate names
		for _, c := range s.clients {
			if c.name == newName {
				conn.WriteToUDP([]byte("\033[31mUsername already taken\033[0m\n"), addr)
				return
			}
		}
		oldName := client.name
		client.name = newName
		client.isAdmin = (newName == "admin") // Update admin status if name changed to "admin"
		// Broadcast name change notification
		s.messages <- fmt.Sprintf("\033[33m[%s] %s changed name to %s\033[0m",
			time.Now().Format("3:04 PM"), oldName, newName)

	case strings.HasPrefix(msg, "WHISPER:"):
		// Handle private messages
		parts := strings.SplitN(strings.TrimPrefix(msg, "WHISPER:"), ":", 2)
		if len(parts) == 2 {
			targetName := parts[0]
			whisperMsg := parts[1]
			// Find target user
			for _, c := range s.clients {
				if c.name == targetName {
					// Send private message to target
					privateMsg := fmt.Sprintf("\033[35m[WHISPER from %s] %s\033[0m",
						client.name, whisperMsg)
					conn.WriteToUDP([]byte(privateMsg), c.addr)
					// Send confirmation to sender
					conn.WriteToUDP([]byte(fmt.Sprintf("\033[36m[Whisper sent to %s]\033[0m",
						targetName)), addr)
					return
				}
			}
			// Target not found
			conn.WriteToUDP([]byte(fmt.Sprintf("\033[31mUser %s not found\033[0m",
				targetName)), addr)
		}

	case strings.HasPrefix(msg, "QUIT:"):
		// Handle client disconnection
		name := strings.TrimPrefix(msg, "QUIT:")
		delete(s.clients, clientKey) // Remove client from map
		// Broadcast leave notification
		s.messages <- fmt.Sprintf("\033[31m[%s] %s left the chat\033[0m",
			time.Now().Format("3:04 PM"), name)

	case strings.HasPrefix(msg, "KICK:") && client.isAdmin:
		// Admin kick command
		targetName := strings.TrimPrefix(msg, "KICK:")
		// Find and kick target user
		for key, c := range s.clients {
			if c.name == targetName {
				delete(s.clients, key) // Remove client
				// Notify kicked user
				conn.WriteToUDP([]byte("\033[31mYou have been kicked by admin\033[0m\n"), c.addr)
				// Broadcast kick notification
				s.messages <- fmt.Sprintf("\033[31m[%s] %s was kicked by admin\033[0m",
					time.Now().Format("3:04 PM"), targetName)
				break
			}
		}

	case strings.HasPrefix(msg, "BROADCAST:") && client.isAdmin:
		// Admin broadcast message
		message := strings.TrimPrefix(msg, "BROADCAST:")
		s.messages <- fmt.Sprintf("\033[33m[ADMIN ANNOUNCEMENT] %s\033[0m", message)

	case strings.HasPrefix(msg, "SHUTDOWN:") && client.isAdmin:
		// Admin shutdown command
		close(s.shutdown) // Trigger shutdown
		return

	default: // Regular chat message or invalid command
		if strings.HasPrefix(msg, "/") {
			// Invalid command
			conn.WriteToUDP([]byte("\033[31mInvalid command. Type /help for available commands\033[0m\n"), addr)
		} else {
			// Format and broadcast regular message
			fullMsg := s.formatMessage(client, msg)
			s.messages <- fullMsg
		}
	}
}

// broadcastMessages sends messages to all connected clients
func (s *Server) broadcastMessages(conn *net.UDPConn) {
	for msg := range s.messages { // Read from messages channel
		s.mu.RLock() // Read lock for clients map
		// Send to all clients
		for _, client := range s.clients {
			_, err := conn.WriteToUDP([]byte(msg+"\n"), client.addr)
			if err != nil {
				log.Printf("Error sending to %s: %v", client.name, err)
			}
		}
		s.mu.RUnlock()
	}
}

// cleanupClients periodically removes inactive clients
func (s *Server) cleanupClients() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown: // Stop on shutdown
			return
		case <-ticker.C: // Every tick
			s.mu.Lock() // Acquire write lock
			now := time.Now()
			timedOutUsers := make([]string, 0)

			// Find inactive clients (skip admins)
			for key, client := range s.clients {
				if client.isAdmin {
					continue
				}

				if now.Sub(client.lastSeen) >= 10*time.Minute {
					timedOutUsers = append(timedOutUsers, key)
				}
			}

			// Remove inactive clients
			for _, key := range timedOutUsers {
				name := s.clients[key].name
				delete(s.clients, key)
				// Broadcast timeout notification
				msg := fmt.Sprintf("\033[33m[%s] %s timed out (inactive for 10 minutes)\033[0m",
					now.Format("3:04 PM"), name)
				s.messages <- msg
				log.Printf("User %s timed out due to inactivity", name)
			}
			s.mu.Unlock()
		}
	}
}

// startServer initializes and starts the chat server
func startServer() {
	s := newServer() // Create server instance

	// Shutdown on Enter key press
	go func() {
		fmt.Println("Press Enter to shutdown server...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		close(s.shutdown) // Trigger shutdown
	}()

	s.start(":8080") // Start server on port 8080
}
