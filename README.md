first commit

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
- Store user credentials securely (consider using a database).  

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

#### **9. Future Enhancements**  
- Store chat history in a database.  
- Implement message encryption.  
- Develop a WebSocket-based version for web support.  
- Add an admin panel for moderation.  

Would you like more details on a specific step? 🚀
