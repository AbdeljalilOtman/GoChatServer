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

// ANSI color codes for better readability
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorBold   = "\033[1m"
)

// Message structure for communication
type Message struct {
	Sender   string
	RoomName string
	Content  string
	Type     string // "text", "file", "command", "file-request", "file-chunk"
	FileData []byte // Used for file transfer
	FileName string // Used for file transfer
}

// Add file transfer state tracking
type fileTransferState struct {
	active      bool
	recipient   string
	filePath    string
	fileName    string
	fileSize    int64
	totalSent   int64
	lastPercent float64
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

	clearScreen()
	printBanner()
	printHelp()

	// Create channels for communication between goroutines
	done := make(chan struct{})
	inputChan := make(chan string)
	quitChan := make(chan struct{}) // New channel for quit signaling

	// Keep track of current input and chat state
	currentInput := ""
	currentRoom := "general"
	loggedIn := false
	username := ""

	// Add file transfer state
	fileTransfer := fileTransferState{}

	// Start goroutine to read messages from the server
	go func() {
		defer func() {
			close(done)
			fmt.Println("Reader goroutine finished.")
		}()

		reader := bufio.NewReader(conn)

		for {
			// Check for quit signal
			select {
			case <-quitChan:
				return
			default:
				// Continue with normal operation
			}

			// Read raw message bytes with timeout
			conn.SetReadDeadline(time.Now().Add(1 * time.Hour)) // 1 hour timeout

			msgBytes, err := reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF || strings.Contains(err.Error(), "connection") {
					fmt.Println("\n" + colorRed + "Server connection closed." + colorReset)
				} else {
					fmt.Printf("\n"+colorRed+"Error reading from server: %v\n"+colorReset, err)
				}
				return
			}

			// Reset read deadline after successful read
			conn.SetReadDeadline(time.Time{})

			var message Message
			err = json.Unmarshal(msgBytes, &message)
			if err != nil {
				fmt.Printf("\n"+colorRed+"Error parsing message: %v\n"+colorReset, err) // Ensure newline if error
				continue
			}

			// Clear the current line (input prompt or progress bar)
			if fileTransfer.active {
				// If in middle of file transfer, preserve progress bar
				fmt.Print("\r\033[K") // Clears current line
			} else {
				fmt.Print("\r\033[K") // Clears current line
			}

			// Check for status messages that should update client state
			if message.Type == "text" && message.Sender == "Server" {
				if strings.Contains(message.Content, "Login successful") ||
					strings.Contains(message.Content, "Registered and logged in") {
					loggedIn = true
					// Extract username from the login command that was sent
					if currentInput != "" && strings.HasPrefix(currentInput, "/login") {
						parts := strings.Fields(currentInput)
						if len(parts) >= 2 {
							username = parts[1]
						}
					}
				} else if strings.HasPrefix(message.Content, "You have joined room:") {
					parts := strings.Fields(message.Content)
					if len(parts) > 4 {
						currentRoom = parts[len(parts)-1]
					}
				}
			}

			// Format output based on message type
			switch message.Type {
			case "text":
				// Skip displaying own messages since server will echo them back
				if message.Sender == username && !strings.Contains(message.Content, "has joined") &&
					!strings.Contains(message.Content, "has left") {
					continue
				}

				// If in file transfer mode and this is a non-critical message, buffer it
				if fileTransfer.active && !strings.Contains(message.Content, "File transfer") {
					// Skip non-critical messages during transfer to avoid breaking progress display
					continue
				}

				// Format room messages nicely
				if message.RoomName != "" && message.RoomName != "general" {
					fmt.Printf(colorYellow+"\n[%s] "+colorBold+"%s: "+colorReset+"%s\n", // Starts with \n
						message.RoomName, message.Sender, message.Content)
				} else if message.Sender == "Server" {
					// System messages in different color
					fmt.Printf(colorPurple+"\n%s: "+colorReset+"%s\n", message.Sender, message.Content) // Starts with \n
				} else {
					// Regular chat messages
					fmt.Printf(colorCyan+"\n%s: "+colorReset+"%s\n", message.Sender, message.Content) // Starts with \n
				}

			case "file-request":
				fmt.Printf(colorPurple+"\n%s wants to send file: %s\nType /accept or /reject\n"+colorReset, // Starts with \n
					message.Sender, message.FileName)

			case "file-accepted":
				fmt.Printf(colorGreen+"\nFile transfer accepted by %s. Starting transfer...\n"+colorReset,
					message.Sender)

				// Resume or start file transfer
				if fileTransfer.active && fileTransfer.recipient == message.Sender {
					go sendFileInChunks(conn, fileTransfer.filePath, fileTransfer.fileName, &fileTransfer)
				}

			case "file-rejected":
				fmt.Printf(colorRed+"\nFile transfer rejected by %s.\n"+colorReset, message.Sender)
				fileTransfer.active = false

			case "file-complete":
				fmt.Printf(colorGreen+"\nFile transfer to %s completed successfully!\n"+colorReset, message.Sender)
				fileTransfer.active = false

			case "file-chunk":
				fmt.Printf(colorBlue+"\nReceived chunk of file: %s\n"+colorReset, message.FileName) // Starts with \n

			default:
				fmt.Printf("\n%s: %s\n", message.Sender, message.Content) // Starts with \n
			}

			// Re-display the current input with prompt or progress bar
			if fileTransfer.active {
				// If in file transfer, redraw the progress bar
				drawProgressBar(fileTransfer.lastPercent, 50, false)
			} else {
				// Otherwise show the input prompt
				prompt := colorGreen + "You > " + colorReset
				if loggedIn && currentRoom != "general" {
					prompt = colorGreen + "[" + currentRoom + "] You > " + colorReset
				} else if loggedIn {
					prompt = colorGreen + "You > " + colorReset
				}
				fmt.Print(prompt + currentInput)
			}
		}
	}()

	// Start goroutine to handle user input
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		// Initial prompt
		initialPrompt := colorGreen + "You > " + colorReset
		if loggedIn && currentRoom != "general" {
			initialPrompt = colorGreen + "[" + currentRoom + "] You > " + colorReset
		} else if loggedIn {
			initialPrompt = colorGreen + "You > " + colorReset
		}
		fmt.Print(initialPrompt)

		for scanner.Scan() {
			// Check if we should quit
			select {
			case <-quitChan:
				return
			default:
				// Continue with normal operation
			}

			text := scanner.Text()
			currentInput = text

			// Check for quit command directly here too
			if strings.HasPrefix(text, "/quit") {
				fmt.Println(colorYellow + "\nExiting..." + colorReset)
				close(quitChan) // Signal all goroutines
				os.Exit(0)      // Force exit the program
				return
			}

			// Provide feedback for sent messages
			if text != "" && !strings.HasPrefix(text, "/") && loggedIn {
				// For normal chat messages, show them in the UI immediately
				// Clear the line first, then print message, then print prompt
				fmt.Print("\r\033[K")
				if currentRoom != "general" {
					fmt.Printf(colorYellow+"\n[%s] "+colorBold+"%s: "+colorReset+"%s\n",
						currentRoom, username, text)
				} else {
					fmt.Printf(colorCyan+"\n%s: "+colorReset+"%s\n", username, text)
				}
			}

			// Send to processing channel
			inputChan <- text

			// Update prompt based on room
			prompt := colorGreen + "You > " + colorReset
			if loggedIn && currentRoom != "general" {
				prompt = colorGreen + "[" + currentRoom + "] You > " + colorReset
			} else if loggedIn {
				prompt = colorGreen + "You > " + colorReset
			}

			// Print the prompt for next input
			fmt.Print(prompt)
			currentInput = ""
		}

		// If scanner.Scan() returns false, handle the error
		if err := scanner.Err(); err != nil {
			fmt.Printf(colorRed+"\nError reading input: %v\n"+colorReset, err)
		}

		// Signal that we're done
		close(quitChan)
	}()

	// Process input from the channel
	for text := range inputChan {
		// Handle special commands
		if strings.HasPrefix(text, "/quit") {
			fmt.Println(colorYellow + "Exiting..." + colorReset)
			// Send goodbye message to server if connected
			conn.Write([]byte("/quit\n"))
			// Close connection cleanly
			conn.Close()
			// Signal all goroutines to stop
			close(quitChan)
			// Exit the program directly
			os.Exit(0)
			break
		} else if strings.HasPrefix(text, "/help") {
			printHelp()
			continue
		} else if strings.HasPrefix(text, "/clear") {
			clearScreen()
			continue
		} else if strings.HasPrefix(text, "/join") {
			// Track room changes locally for better UI
			parts := strings.Fields(text)
			if len(parts) > 1 {
				// Will be confirmed by server message
				currentRoom = parts[1]
			}
		} else if strings.HasPrefix(text, "/sendfile") {
			parts := strings.Fields(text)
			if len(parts) < 3 {
				fmt.Print("\r\033[K") // Clear line before printing error
				fmt.Println(colorRed + "Usage: /sendfile <username> <filepath>" + colorReset)
				printPrompt(loggedIn, currentRoom)
				continue
			}

			recipient := parts[1]
			filePath := parts[2]

			// Read the file
			file, err := os.Open(filePath)
			if err != nil {
				fmt.Print("\r\033[K")
				fmt.Println(colorRed+"Error opening file:"+colorReset, err)
				printPrompt(loggedIn, currentRoom)
				continue
			}

			// Get file info
			fileInfo, err := file.Stat()
			if err != nil {
				fmt.Print("\r\033[K")
				fmt.Println(colorRed+"Error getting file info:"+colorReset, err)
				file.Close()
				printPrompt(loggedIn, currentRoom)
				continue
			}

			fileSize := fileInfo.Size()
			fileName := fileInfo.Name()
			file.Close() // Close the file for now, we'll reopen it when sending

			// Set up file transfer state
			fileTransfer = fileTransferState{
				active:      true,
				recipient:   recipient,
				filePath:    filePath,
				fileName:    fileName,
				fileSize:    fileSize,
				totalSent:   0,
				lastPercent: 0,
			}

			// Send file transfer request
			fmt.Print("\r\033[K") // Clear line before printing status
			fmt.Printf(colorBlue+"Initiating file transfer: %s (%.2f KB)\n"+colorReset,
				fileName, float64(fileSize)/1024)
			_, err = conn.Write([]byte(fmt.Sprintf("/sendfile %s %s %d\n",
				recipient, fileName, fileSize)))

			if err != nil {
				fmt.Print("\r\033[K")
				fmt.Println(colorRed+"Error sending file transfer request:"+colorReset, err)
				fileTransfer.active = false
				printPrompt(loggedIn, currentRoom)
			}
			continue
		} else if strings.HasPrefix(text, "/accept") || strings.HasPrefix(text, "/reject") {
			// Add explicit handling for accept/reject commands
			// Just pass these commands through to the server
			_, err := conn.Write([]byte(text + "\n"))
			if err != nil {
				fmt.Println(colorRed+"Error sending command:"+colorReset, err)
				break
			}
			continue
		}

		// Send the message to the server
		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			fmt.Println(colorRed+"Error sending message:"+colorReset, err)
			break
		}
	}

	// Make sure we wait for the reader goroutine to finish cleanly
	select {
	case <-done:
		// Reader exited normally
	case <-time.After(2 * time.Second):
		// Timeout to avoid hanging
	}
}

// Helper functions for a better UI
func clearScreen() {
	fmt.Print("\033[H\033[2J") // ANSI escape sequence to clear screen
}

func printBanner() {
	banner := `
` + colorBlue + `
   ______       ________          __   _________             __
  / ____/____  / ____/ /_  ____ _/ /_ / ____/ (_)__  ____  / /_
 / / __/ __ \/ /   / __ \/ __ '/ __// /   / / / _ \/ __ \/ __/
/ /_/ / /_/ / /___/ / / / /_/ / /_ / /___/ / /  __/ / / / /_
\____/\____/\____/_/ /_/\__,_/\__/ \____/_/_/\___/_/ /_/\__/
` + colorReset + `
`
	fmt.Println(banner)
	fmt.Println(colorYellow + "Welcome to the GoChatServer client!" + colorReset)
	fmt.Println(colorCyan + "Type /help to see available commands" + colorReset)
	fmt.Println()
}

func printHelp() {
	help := `
` + colorBold + colorCyan + `AVAILABLE COMMANDS:` + colorReset + `
  ` + colorGreen + `/login <username> <password>` + colorReset + ` - Log in to the server
  ` + colorGreen + `/join <roomname>` + colorReset + `          - Join a chat room
  ` + colorGreen + `/rooms` + colorReset + `                   - List available rooms
  ` + colorGreen + `/users` + colorReset + `                   - List users in current room
  ` + colorGreen + `/sendfile <username> <filepath>` + colorReset + ` - Send a file to a user
  ` + colorGreen + `/accept` + colorReset + `                  - Accept an incoming file transfer
  ` + colorGreen + `/reject` + colorReset + `                  - Reject an incoming file transfer
  ` + colorGreen + `/clear` + colorReset + `                   - Clear the screen
  ` + colorGreen + `/help` + colorReset + `                    - Show this help message
  ` + colorGreen + `/quit` + colorReset + `                    - Exit the client

Type your message and press Enter to send it to the current room.
`
	fmt.Println(help)
}

// Helper to print the appropriate prompt
func printPrompt(loggedIn bool, currentRoom string) {
	prompt := colorGreen + "You > " + colorReset
	if loggedIn && currentRoom != "general" {
		prompt = colorGreen + "[" + currentRoom + "] You > " + colorReset
	} else if loggedIn {
		prompt = colorGreen + "You > " + colorReset
	}
	fmt.Print(prompt)
}

// Send file in chunks as a separate goroutine
func sendFileInChunks(conn net.Conn, filePath string, fileName string, state *fileTransferState) {
	// Create a copy of the state to avoid race conditions
	fileSize := state.fileSize

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Print("\r\033[K")
		fmt.Println(colorRed+"Error opening file:"+colorReset, err)
		state.active = false
		return
	}
	defer file.Close()

	// Send file in chunks
	buffer := make([]byte, 8*1024) // 8KB chunks
	var totalSent int64

	for {
		n, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Print("\r\033[K") // Clear progress bar line
			fmt.Println(colorRed+"Error reading file:"+colorReset, err)
			state.active = false
			return
		}

		// Send chunk
		fileMessage := Message{
			Type:     "file-chunk",
			FileName: fileName,
			FileData: buffer[:n],
		}

		jsonFileMsg, err := json.Marshal(fileMessage)
		if err != nil {
			fmt.Print("\r\033[K") // Clear progress bar line
			fmt.Println(colorRed+"Error marshaling file message:"+colorReset, err)
			state.active = false
			return
		}

		jsonFileMsg = append(jsonFileMsg, '\n')
		_, err = conn.Write(jsonFileMsg)
		if err != nil {
			fmt.Print("\r\033[K") // Clear progress bar line
			fmt.Println(colorRed+"Error sending file chunk:"+colorReset, err)
			state.active = false
			return
		}

		totalSent += int64(n)
		progress := float64(totalSent) / float64(fileSize) * 100

		// Thread-safely update the state
		state.totalSent = totalSent
		state.lastPercent = progress

		// Update progress bar (only update every ~2%)
		if totalSent%(fileSize/50+1) == 0 || totalSent >= fileSize {
			fmt.Print("\r\033[K") // Clear line
			drawProgressBar(progress, 50, false)
		}

		// Small sleep to avoid flooding the network
		time.Sleep(5 * time.Millisecond)
	}

	// File transfer complete
	fmt.Print("\r\033[K") // Clear line
	drawProgressBar(100, 50, true)
	fmt.Println(colorGreen + "File sent successfully!" + colorReset)
	state.active = false

	// Allow time for the final chunks to be processed
	time.Sleep(500 * time.Millisecond)
	printPrompt(true, "general") // Default prompt after transfer
}

// Improved progress bar that's more resilient to interference
func drawProgressBar(percent float64, width int, finalCall bool) {
	fmt.Print("\r[") // Carriage return to start of line
	completedWidth := int(percent / 100 * float64(width))
	for i := 0; i < width; i++ {
		if i < completedWidth {
			fmt.Print(colorGreen + "=" + colorReset)
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f%% ", percent) // Extra space at end for clean overwrite

	if finalCall {
		fmt.Println() // Only print newline on final call
	}
}
