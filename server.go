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

type Client struct {
	addr     *net.UDPAddr
	name     string
	lastSeen time.Time
	isAdmin  bool
}

type Server struct {
	clients   map[string]*Client
	mu        sync.RWMutex
	messages  chan string
	startTime time.Time
	shutdown  chan struct{}
}

func newServer() *Server {
	return &Server{
		clients:   make(map[string]*Client),
		messages:  make(chan string, 100),
		startTime: time.Now(),
		shutdown:  make(chan struct{}),
	}
}

func (s *Server) start(port string) {
	addr, err := net.ResolveUDPAddr("udp", port)
	if err != nil {
		log.Fatal("Resolve address error:", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("Listen error:", err)
	}
	defer conn.Close()

	log.Printf("Server started on %s", port)

	go s.broadcastMessages(conn)
	go s.cleanupClients()

	buf := make([]byte, 1024)
	for {
		select {
		case <-s.shutdown:
			log.Println("Shutting down server...")
			s.mu.RLock()
			for _, client := range s.clients {
				conn.WriteToUDP([]byte("\033[31mServer is shutting down. Goodbye!\033[0m\n"), client.addr)
			}
			s.mu.RUnlock()
			return
		default:
			conn.SetReadDeadline(time.Now().Add(1 * time.Minute))
			n, clientAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				log.Printf("Read error: %v", err)
				continue
			}
			go s.handleMessage(conn, clientAddr, string(buf[:n]))
		}
	}
}

func (s *Server) handleMessage(conn *net.UDPConn, addr *net.UDPAddr, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.HasPrefix(msg, "TYPING:") {
		name := strings.TrimPrefix(msg, "TYPING:")
		s.messages <- fmt.Sprintf("\033[2m%s is typing...\033[0m", name)
		return
	}

	clientKey := addr.String()
	client, exists := s.clients[clientKey]

	if !exists {
		if strings.HasPrefix(msg, "REGISTER:") {
			name := strings.TrimPrefix(msg, "REGISTER:")
			for _, c := range s.clients {
				if c.name == name {
					conn.WriteToUDP([]byte("\033[31mUsername already taken. Please choose another.\033[0m\n"), addr)
					return
				}
			}

			isAdmin := (name == "admin")
			newClient := &Client{
				addr:     addr,
				name:     name,
				lastSeen: time.Now(),
				isAdmin:  isAdmin,
			}
			s.clients[clientKey] = newClient

			welcome := fmt.Sprintf("\033[32m[%s] %s joined the chat\033[0m",
				time.Now().Format("3:04 PM"), name)
			s.messages <- welcome

			if isAdmin {
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

	client.lastSeen = time.Now()

	switch {
	case msg == "/menu" && client.isAdmin:
		adminMenu := "\n\033[33mADMIN MENU:\n" +
			"1. /users - List all users\n" +
			"2. /kick <username> - Kick a user\n" +
			"3. /broadcast <msg> - Server announcement\n" +
			"4. /stats - Server statistics\n" +
			"5. /shutdown - Shutdown server\033[0m\n"
		conn.WriteToUDP([]byte(adminMenu), addr)

	case msg == "/users":
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
		help := "\033[1mCommands:\033[0m\n" +
			"/users - List online users\n" +
			"/help - Show this help\n" +
			"/stats - Show server statistics\n" +
			"/quit - Disconnect from server\n"
		conn.WriteToUDP([]byte(help), addr)

	case msg == "/stats":
		uptime := time.Since(s.startTime).Round(time.Second)
		stats := fmt.Sprintf("\033[1mServer Stats:\033[0m\n"+
			"Uptime: %s\n"+
			"Users connected: %d\n"+
			"Timeout: 10 minutes (except admins)\n", uptime, len(s.clients))
		conn.WriteToUDP([]byte(stats), addr)

	case strings.HasPrefix(msg, "RENAME:"):
		newName := strings.TrimPrefix(msg, "RENAME:")
		for _, c := range s.clients {
			if c.name == newName {
				conn.WriteToUDP([]byte("\033[31mUsername already taken\033[0m\n"), addr)
				return
			}
		}
		oldName := client.name
		client.name = newName
		client.isAdmin = (newName == "admin")
		s.messages <- fmt.Sprintf("\033[33m[%s] %s changed name to %s\033[0m",
			time.Now().Format("3:04 PM"), oldName, newName)

	case strings.HasPrefix(msg, "WHISPER:"):
		parts := strings.SplitN(strings.TrimPrefix(msg, "WHISPER:"), ":", 2)
		if len(parts) == 2 {
			targetName := parts[0]
			whisperMsg := parts[1]
			for _, c := range s.clients {
				if c.name == targetName {
					privateMsg := fmt.Sprintf("\033[35m[WHISPER from %s] %s\033[0m",
						client.name, whisperMsg)
					conn.WriteToUDP([]byte(privateMsg), c.addr)
					conn.WriteToUDP([]byte(fmt.Sprintf("\033[36m[Whisper sent to %s]\033[0m",
						targetName)), addr)
					return
				}
			}
			conn.WriteToUDP([]byte(fmt.Sprintf("\033[31mUser %s not found\033[0m",
				targetName)), addr)
		}

	case strings.HasPrefix(msg, "QUIT:"):
		name := strings.TrimPrefix(msg, "QUIT:")
		delete(s.clients, clientKey)
		s.messages <- fmt.Sprintf("\033[31m[%s] %s left the chat\033[0m",
			time.Now().Format("3:04 PM"), name)

	case strings.HasPrefix(msg, "KICK:") && client.isAdmin:
		targetName := strings.TrimPrefix(msg, "KICK:")
		for key, c := range s.clients {
			if c.name == targetName {
				delete(s.clients, key)
				conn.WriteToUDP([]byte("\033[31mYou have been kicked by admin\033[0m\n"), c.addr)
				s.messages <- fmt.Sprintf("\033[31m[%s] %s was kicked by admin\033[0m",
					time.Now().Format("3:04 PM"), targetName)
				break
			}
		}

	case strings.HasPrefix(msg, "BROADCAST:") && client.isAdmin:
		message := strings.TrimPrefix(msg, "BROADCAST:")
		s.messages <- fmt.Sprintf("\033[33m[ADMIN ANNOUNCEMENT] %s\033[0m", message)

	case strings.HasPrefix(msg, "SHUTDOWN:") && client.isAdmin:
		close(s.shutdown)
		return

	default:
		if strings.HasPrefix(msg, "/") {
			conn.WriteToUDP([]byte("\033[31mInvalid command. Type /help for available commands\033[0m\n"), addr)
		} else {
			fullMsg := fmt.Sprintf("\033[34m[%s] %s:\033[0m %s",
				time.Now().Format("3:04 PM"), client.name, msg)
			s.messages <- fullMsg
		}
	}
}

func (s *Server) broadcastMessages(conn *net.UDPConn) {
	for msg := range s.messages {
		s.mu.RLock()
		for _, client := range s.clients {
			_, err := conn.WriteToUDP([]byte(msg+"\n"), client.addr)
			if err != nil {
				log.Printf("Error sending to %s: %v", client.name, err)
			}
		}
		s.mu.RUnlock()
	}
}

func (s *Server) cleanupClients() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-s.shutdown:
			return
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			timedOutUsers := make([]string, 0)

			for key, client := range s.clients {
				if client.isAdmin {
					continue
				}

				if now.Sub(client.lastSeen) >= 10*time.Minute {
					timedOutUsers = append(timedOutUsers, key)
				}
			}

			for _, key := range timedOutUsers {
				name := s.clients[key].name
				delete(s.clients, key)
				msg := fmt.Sprintf("\033[33m[%s] %s timed out (inactive for 10 minutes)\033[0m",
					now.Format("3:04 PM"), name)
				s.messages <- msg
				log.Printf("User %s timed out due to inactivity", name)
			}
			s.mu.Unlock()
		}
	}
}

func startServer() {
	s := newServer()

	go func() {
		fmt.Println("Press Enter to shutdown server...")
		bufio.NewReader(os.Stdin).ReadString('\n')
		close(s.shutdown)
	}()

	s.start(":8080")
}
