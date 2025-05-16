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
| `/accept [username]` | Accept an incoming file transfer | `/accept alice` |
| `/reject [username]` | Reject an incoming file transfer | `/reject alice` |
| `/clear` | Clear the terminal screen | `/clear` |
| `/help` | Display available commands | `/help` |
| `/quit` | Exit the client | `/quit` |

## 📁 Testing File Transfer

The file transfer feature allows users to send files to each other through the chat. Here's how to test it:

### Setting Up for Transfer

1. Start the server:
   ```bash
   cd GoChatServer
   go run main.go
   ```

2. Open two terminal windows and start two client instances:
   ```bash
   cd GoChatServer/client
   go run client.go
   ```

3. In each client, log in with different usernames:
   ```
   /login user1 password1
   ```
   ```
   /login user2 password2
   ```

### Sending Files

4. In the first client (user1), send a file to the second client:
   ```
   /sendfile user2 path/to/your/file.txt
   ```
   You can use the test.txt file included in the project:
   ```
   /sendfile user2 ../test.txt
   ```

5. In the second client (user2), you'll see a notification about the incoming file:
   ```
   user1 wants to send file: file.txt
   Type /accept user1 or /reject user1
   ```

6. To accept the file transfer, type:
   ```
   /accept user1
   ```
   Or simply:
   ```
   /accept
   ```
   (if you only have one pending transfer)

7. The file will be transferred and saved in the `downloads` folder in the server's directory.

### Transfer Progress

During the file transfer:
- The sender will see a progress bar showing the transfer status
- The receiver will be notified when the transfer is complete
- The terminal will display the saved file location

### Troubleshooting

- **"No pending file transfer"**: Make sure you've typed the correct username when accepting
- **File not found**: Check that the file path is correct and the file exists
- **Transfer stalls**: Ensure both clients remain connected during the transfer
- **Permission denied**: Ensure the server has write access to create the downloads directory

## 📝 Additional Information

- **File Storage**: Received files are saved in the `downloads` directory on the server, with a timestamp prefix to avoid name conflicts
- **Transfer Limits**: The default chunk size is 8KB, suitable for most files
- **Supported File Types**: All file types are supported
- **Maximum File Size**: There is no hard limit on file size, but very large files may take significant time to transfer

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
