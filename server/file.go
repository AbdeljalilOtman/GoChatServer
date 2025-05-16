package server

import (
	"fmt"
	"sync"
	"time"
)

const (
	chunkSize = 1024 * 8 // 8KB chunks
)

// FileTransfer tracks ongoing file transfers
type FileTransfer struct {
	Sender       *Client
	Receiver     *Client
	FileName     string
	FileSize     int64
	ReceivedSize int64
	Status       string // "pending", "accepted", "rejected", "complete", "failed"
	StartTime    time.Time
}

// Global map to track file transfers
var (
	activeTransfers = make(map[string]*FileTransfer) // Key is Sender.username + "_" + Receiver.username + "_" + FileName
	transferMutex   sync.Mutex
)

// InitiateFileTransfer sets up a new file transfer
func InitiateFileTransfer(sender, receiver *Client, fileName string, fileSize int64) *FileTransfer {
	if sender == nil || receiver == nil || fileName == "" || fileSize <= 0 {
		return nil
	}

	transferMutex.Lock()
	defer transferMutex.Unlock()

	key := fmt.Sprintf("%s_%s_%s", sender.username, receiver.username, fileName)

	// Check if transfer already exists
	if existing, found := activeTransfers[key]; found {
		if existing.Status == "pending" {
			// Already have a pending transfer, update it
			existing.FileSize = fileSize
			existing.StartTime = time.Now()
			return existing
		}
		// Otherwise, clean up old transfer first
		delete(activeTransfers, key)
	}

	transfer := &FileTransfer{
		Sender:       sender,
		Receiver:     receiver,
		FileName:     fileName,
		FileSize:     fileSize,
		ReceivedSize: 0,
		Status:       "pending",
		StartTime:    time.Now(),
	}

	activeTransfers[key] = transfer

	return transfer
}

// GetActiveTransfer retrieves an active transfer if it exists
func GetActiveTransfer(sender, receiver *Client, fileName string) *FileTransfer {
	transferMutex.Lock()
	defer transferMutex.Unlock()

	key := fmt.Sprintf("%s_%s_%s", sender.username, receiver.username, fileName)
	return activeTransfers[key]
}

// UpdateTransferStatus changes the status of a transfer
func UpdateTransferStatus(transfer *FileTransfer, status string) {
	transferMutex.Lock()
	defer transferMutex.Unlock()

	if transfer != nil {
		transfer.Status = status
	}
}

// RemoveTransfer removes a completed transfer
func RemoveTransfer(transfer *FileTransfer) {
	if transfer == nil {
		return
	}

	transferMutex.Lock()
	defer transferMutex.Unlock()

	key := fmt.Sprintf("%s_%s_%s",
		transfer.Sender.username,
		transfer.Receiver.username,
		transfer.FileName)

	delete(activeTransfers, key)
}

// AcceptFileTransfer marks a transfer as accepted
func (c *Client) AcceptFileTransfer(senderUsername string) {
	// Find the pending transfer
	var transfer *FileTransfer

	transferMutex.Lock()
	for k, t := range activeTransfers {
		if t.Receiver == c && t.Sender.username == senderUsername && t.Status == "pending" {
			transfer = t
			transfer.Status = "accepted"
			fmt.Printf("File transfer accepted: %s\n", k) // Fixed: using k from the loop now
			break
		}
	}
	transferMutex.Unlock()

	if transfer == nil {
		// Debug: List all active transfers
		fmt.Printf("No file transfer found for %s from %s. Active transfers:\n", c.username, senderUsername)
		transferMutex.Lock()
		for k, t := range activeTransfers {
			fmt.Printf("- Transfer %s: sender=%s, receiver=%s, status=%s\n",
				k, t.Sender.username, t.Receiver.username, t.Status)
		}
		transferMutex.Unlock()

		c.directSend(Message{
			Sender:  "Server",
			Content: "No pending file transfer from " + senderUsername,
			Type:    "text",
		})
		return
	}

	// Notify the sender that the transfer was accepted
	transfer.Sender.directSend(Message{
		Sender:  transfer.Receiver.username,
		Content: fmt.Sprintf("File transfer request for %s accepted", transfer.FileName),
		Type:    "file-accepted",
	})

	c.directSend(Message{
		Sender:  "Server",
		Content: "File transfer accepted. Receiving file...",
		Type:    "text",
	})
}

// RejectFileTransfer marks a transfer as rejected
func (c *Client) RejectFileTransfer(senderUsername string) {
	// Find the pending transfer
	var transfer *FileTransfer

	transferMutex.Lock()
	for k, t := range activeTransfers {
		if t.Receiver == c && t.Sender.username == senderUsername && t.Status == "pending" {
			transfer = t
			transfer.Status = "rejected"
			fmt.Printf("File transfer rejected: %s\n", k) // Debug logging
			break
		}
	}
	transferMutex.Unlock()

	if transfer == nil {
		// Debug: List all active transfers
		fmt.Printf("No file transfer found for %s from %s to reject. Active transfers:\n", c.username, senderUsername)
		transferMutex.Lock()
		for k, t := range activeTransfers {
			fmt.Printf("- Transfer %s: sender=%s, receiver=%s, status=%s\n",
				k, t.Sender.username, t.Receiver.username, t.Status)
		}
		transferMutex.Unlock()

		c.directSend(Message{
			Sender:  "Server",
			Content: "No pending file transfer from " + senderUsername,
			Type:    "text",
		})
		return
	}

	// Notify the sender that the transfer was rejected
	transfer.Sender.directSend(Message{
		Sender:  transfer.Receiver.username,
		Content: fmt.Sprintf("File transfer request for %s rejected", transfer.FileName),
		Type:    "file-rejected",
	})

	c.directSend(Message{
		Sender:  "Server",
		Content: "File transfer rejected.",
		Type:    "text",
	})

	// Clean up
	RemoveTransfer(transfer)
}

// ProcessFileTransfer handles file data reception
func (c *Client) ProcessFileTransfer(data []byte, fileName string, isLastChunk bool) {
	// First check if this is part of an active transfer
	var transfer *FileTransfer

	transferMutex.Lock()
	for _, t := range activeTransfers {
		if t.Sender == c && t.FileName == fileName && t.Status == "accepted" {
			transfer = t
			break
		}
	}
	transferMutex.Unlock()

	if transfer == nil {
		// No active accepted transfer found
		c.directSend(Message{
			Sender:  "Server",
			Content: "No active file transfer found. Recipient may not have accepted yet.",
			Type:    "text",
		})
		return
	}

	// Write data to recipient
	transfer.Receiver.directSend(Message{
		Sender:   c.username,
		FileName: fileName,
		FileData: data,
		Type:     "file-chunk",
	})

	// Update progress
	transferMutex.Lock()
	transfer.ReceivedSize += int64(len(data))
	progress := float64(transfer.ReceivedSize) / float64(transfer.FileSize) * 100
	transferMutex.Unlock()

	// Notify sender of progress periodically
	if transfer.ReceivedSize%(transfer.FileSize/10+1) == 0 || isLastChunk {
		c.directSend(Message{
			Sender:  "Server",
			Content: fmt.Sprintf("Transfer progress: %.1f%%", progress),
			Type:    "text",
		})
	}

	// If complete, notify both parties
	if isLastChunk || transfer.ReceivedSize >= transfer.FileSize {
		// Mark as complete
		UpdateTransferStatus(transfer, "complete")

		// Notify sender and receiver
		c.directSend(Message{
			Sender: "Server",
			Content: fmt.Sprintf("File %s transferred successfully to %s",
				fileName, transfer.Receiver.username),
			Type: "text",
		})

		transfer.Receiver.directSend(Message{
			Sender: "Server",
			Content: fmt.Sprintf("File %s received successfully from %s",
				fileName, c.username),
			Type: "file-complete",
		})

		// Clean up
		RemoveTransfer(transfer)
	}
}
