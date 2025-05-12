# UDP-chat-server https://github.com/MJPelayo/UDP-chat-server

# TCP -chat-server https://github.com/ScarletSalinas/CMPS2242SemesterProject

# YouTube video link: https://www.youtube.com/watch?v=wSEWHL-QNBE

# GoChat - A UDP/TCP Chat Application

GoChat is a high-performance, command-line chat application built in Go that supports both UDP and TCP protocols. It features a robust server-client architecture with advanced functionality including private messaging, admin controls, and performance benchmarking.

## Features

### 🚀 Core Features
- Multi-client support via UDP with TCP benchmarking
- Admin privileges system (username "admin")
- Interactive help menu with command auto-completion
- Color-coded messages for better readability
- Typing indicators for active users

### ⚙️ Server Features
- User management (list, kick, timeout)
- Server announcements and broadcasts
- Graceful shutdown capability
- Automatic cleanup of inactive clients (10-minute timeout)
- Detailed server statistics

### 💻 Client Features
- Private messaging (`/whisper` command)
- Username changing (`/rename` command)
- Admin menu for privileged users
- Connection status indicators
- Clean, colorized interface

### 📊 Benchmarking
- UDP performance testing
- TCP performance comparison
- Message rate calculation
- Latency measurement per message

## Installation

### Prerequisites
- Go 1.16 or higher
- Git (for cloning the repository)

### Steps
1. Clone the repository:
   git clone https://github.com/yourusername/gochat.git
   cd gochat

  ## Build the application:
go build

## Usage
# Starting the Server

./gochat server

# Starting a Client

./gochat client <server-address> <username>

## Example:

./gochat client localhost:8080 alice

# Admin Access
- Use the username "admin" for privileged access:

./gochat client localhost:8080 admin

## Command Reference
General Commands
Command	Description
/help	Show interactive help menu
/users	List all online users
/stats	Display server statistics
/quit	Disconnect from the server
/rename <newname>	Change your username
/whisper <user> <msg>	Send a private message

## Admin Commands
Command	Description
/menu	Show admin control panel
/kick <username>	Remove a user from the server
/broadcast <msg>	Send a server-wide announcement
/shutdown	Shut down the server

# Benchmarking
To run performance tests:

go test -bench=.

# Benchmark metrics include:

Total messages processed

Messages per second rate

Per-message latency in microseconds
