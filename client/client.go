package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

type Message struct {
	Sender   string
	RoomName string
	Content  string
	Type     string // "text", "file", "command", "file-request", "file-chunk"
	FileData []byte // Used for file transfer
	FileName string // Used for file transfer
}

func main() {
	serverAddr := flag.String("server", "localhost:8080", "Server address in the form host:port")
	flag.Parse()

	// Connect to the server
	conn, err := net.Dial("tcp", *serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to", *serverAddr)
	fmt.Println("Welcome to the chat client!")
	fmt.Println("Commands:")
	fmt.Println("  /login <username> <password> - Log in to the server")
	fmt.Println("  /join <roomname> - Join a chat room")
	fmt.Println("  /rooms - List available rooms")
	fmt.Println("  /users - List users in current room")
	fmt.Println("  /sendfile <username> <filepath> - Send a file to a user")
	fmt.Println("  /quit - Exit the client")

	// Create a channel to signal when the read goroutine exits
	done := make(chan struct{})

	// Start goroutine to read messages from the server
	go func() {
		defer func() {
			close(done)
			fmt.Println("Reader goroutine finished.")
		}()

		reader := bufio.NewReader(conn)

		for {
			// Read raw message bytes with timeout
			conn.SetReadDeadline(time.Now().Add(1 * time.Hour)) // 1 hour timeout

			msgBytes, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "connection") {
					fmt.Println("Server connection closed.")
				} else {
					fmt.Printf("Error reading from server: %v\n", err)
				}
				return
			}

			// Reset read deadline after successful read
			conn.SetReadDeadline(time.Time{})

			// Debug: Print raw received message
			fmt.Printf("RAW: %s", string(msgBytes))

			var message Message
			err = json.Unmarshal(msgBytes, &message)
			if err != nil {
				fmt.Printf("Error parsing message: %v\n", err)
				continue
			}

			// Format output based on message type
			switch message.Type {
			case "text":
				if message.RoomName != "" && message.RoomName != "general" {
					fmt.Printf("\n[%s] %s: %s\n", message.RoomName, message.Sender, message.Content)
				} else {
					fmt.Printf("\n%s: %s\n", message.Sender, message.Content)
				}

			case "file-request":
				fmt.Printf("\n%s wants to send file: %s\nType /accept or /reject\n",
					message.Sender, message.FileName)

			case "file-chunk":
				fmt.Printf("\nReceived chunk of file: %s\n", message.FileName)

			default:
				fmt.Printf("\n%s: %s\n", message.Sender, message.Content)
			}

			fmt.Print("> ")
		}
	}()

	// Read input from the user and send to the server
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")

	for scanner.Scan() {
		text := scanner.Text()

		// Handle special commands
		if strings.HasPrefix(text, "/quit") {
			fmt.Println("Exiting...")
			break
		} else if strings.HasPrefix(text, "/sendfile") {
			parts := strings.Fields(text)
			if len(parts) < 3 {
				fmt.Println("Usage: /sendfile <username> <filepath>")
				continue
			}

			recipient := parts[1]
			filePath := parts[2]

			// Read the file
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Println("Error opening file:", err)
				continue
			}

			// Get file info
			fileInfo, err := file.Stat()
			if err != nil {
				fmt.Println("Error getting file info:", err)
				file.Close()
				continue
			}

			fileSize := fileInfo.Size()
			fileName := fileInfo.Name()

			// Send file transfer request
			fmt.Printf("Initiating file transfer: %s (%.2f KB)\n", fileName, float64(fileSize)/1024)
			_, err = conn.Write([]byte(fmt.Sprintf("/sendfile %s %s %d\n", recipient, fileName, fileSize)))
			if err != nil {
				fmt.Println("Error sending file transfer request:", err)
				file.Close()
				continue
			}

			// Send file in chunks
			buffer := make([]byte, 8*1024) // 8KB chunks
			var totalSent int64
			for {
				n, err := file.Read(buffer)
				if err == io.EOF {
					break
				}
				if err != nil {
					fmt.Println("Error reading file:", err)
					break
				}

				// Send chunk
				fileMessage := Message{
					Type:     "file-chunk",
					FileName: fileName,
					FileData: buffer[:n],
				}

				jsonFileMsg, err := json.Marshal(fileMessage)
				if err != nil {
					fmt.Println("Error marshaling file message:", err)
					break
				}

				jsonFileMsg = append(jsonFileMsg, '\n')
				_, err = conn.Write(jsonFileMsg)
				if err != nil {
					fmt.Println("Error sending file chunk:", err)
					break
				}

				totalSent += int64(n)
				progress := float64(totalSent) / float64(fileSize) * 100
				if totalSent%(fileSize/10+1) == 0 || totalSent >= fileSize {
					fmt.Printf("Sending: %.1f%% complete\n", progress)
				}
			}

			file.Close()
			fmt.Println("File sent!")
			continue
		}

		// Send the message to the server
		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}

		// Don't print prompt immediately for commands that expect responses
		if !strings.HasPrefix(text, "/") {
			fmt.Print("> ")
		}
	}

	// Wait for the reader goroutine to finish
	<-done
}
