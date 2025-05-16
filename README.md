# GoChatServer - TCP-based Chat Application

A robust, feature-rich chat server built with Go that supports multiple rooms, direct messaging, user authentication, and file transfers.

## Collaborators

- Abdeljalil Otman

## Features

- **User Authentication**: Secure login system with password hashing
- **Multiple Chat Rooms**: Create and join different chat rooms
- **Room Management**: List available rooms and users in a room
- **Direct Messaging**: Send private messages to specific users
- **File Transfer**: Share files between users
- **Command System**: Simple command interface for all operations

## Technical Requirements

- Go 1.21 or later
- TCP/IP networking support
- No external dependencies required (uses only standard library)

## Installation Instructions

1. Clone the repository:
   ```bash
   git clone https://github.com/abdeljalil/GoChatServer.git
   cd GoChatServer
   ```

2. Build the server:
   ```bash
   go build -o chat_server main.go
   ```

3. Build the client:
   ```bash
   cd client
   go build -o chat_client client.go
   ```

## Usage 🚀

### 🖥️ Starting the Server

```bash
./chat_server -port 8080 -v
```

**Options:**
| Flag | Description | Default |
|------|-------------|---------|
| `-port` | Port number to listen on | `8080` |
| `-v` | Enable verbose logging | `false` |

---

### 📱 Running the Client

```bash
./chat_client -server localhost:8080
```

**Options:**
| Flag | Description | Default |
|------|-------------|---------|
| `-server` | Server address (host:port) | `localhost:8080` |

---

### 🔤 Client Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/login <username> <password>` | Authenticate with the server | `/login alice secret123` |
| `/join <roomname>` | Enter a specific chat room | `/join general` |
| `/rooms` | Show list of available rooms | `/rooms` |
| `/users` | List users in current room | `/users` |
| `/sendfile <username> <filepath>` | Send a file to a user | `/sendfile bob /path/to/file.txt` |
| `/quit` | Exit the client | `/quit` |

## Project Structure

- `server/`: Server implementation
  - `server.go`: Main server logic
  - `client.go`: Client connection handling
  - `room.go`: Chat room implementation
  - `user.go`: User authentication
  - `file.go`: File transfer functionality
- `client/`: Client implementation
- `cmd/`: Alternative client/server implementations
- `main.go`: Server entry point

## Development Roadmap

### **Steps to Build a TCP Chat Server in Go**  

#### **1. Project Setup**  
- Install Go and set up a new project.  
- Initialize a Go module for dependency management.  

#### **2. Build the TCP Server**  
- Create a TCP server that listens for incoming connections.  
- Handle client connections in separate goroutines.  
- Implement message broadcasting to all connected clients.  

#### **3. Support Multiple Clients**  
- Use a map to track connected clients.  
- Use mutex locks to manage concurrent access.  

#### **4. Implement Room Management**  
- Create a structure to manage chat rooms.  
- Allow clients to join, leave, and list rooms.  
- Restrict message broadcasting to room members.  

#### **5. Implement User Authentication**  
- Create a user system with usernames and passwords.  
- Require authentication before allowing users to send messages.  
- Store user credentials securely.  

#### **6. Implement File Transfer**  
- Enable users to send files to others in the chat.  
- Read and write file data in chunks.  
- Display file transfer progress and status.  

#### **7. Implement a Client**  
- Develop a CLI-based client to connect to the server.  
- Allow sending and receiving messages.  
- Support commands for authentication, rooms, and file transfer.  

#### **8. Deploy and Test**  
- Run multiple clients to test communication.  
- Debug and optimize performance.  
- Ensure proper error handling and security measures.  

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see [choosealicense.com](https://choosealicense.com/licenses/mit/) for details.
