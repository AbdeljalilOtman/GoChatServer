package server

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	chunkSize = 1024 * 8 // 8KB chunks
)

// ProcessFileTransfer handles file data reception
func (c *Client) ProcessFileTransfer(data []byte, fileName string, isLastChunk bool) {
	c.fileBuffer.Write(data)
	c.receivedSize += int64(len(data))

	progress := float64(c.receivedSize) / float64(c.fileSize) * 100
	if c.receivedSize%(c.fileSize/10+1) == 0 || isLastChunk {
		c.send <- Message{
			Sender:  "Server",
			Content: fmt.Sprintf("Receiving %s: %.1f%% complete", fileName, progress),
			Type:    "text",
		}
	}

	if isLastChunk {
		// Save the file
		err := os.MkdirAll("downloads", 0755)
		if err != nil {
			c.send <- Message{
				Sender:  "Server",
				Content: "Error saving file: " + err.Error(),
				Type:    "text",
			}
			return
		}

		// Generate unique filename to avoid overwriting
		savePath := filepath.Join("downloads", fmt.Sprintf("%d_%s", time.Now().Unix(), fileName))
		file, err := os.Create(savePath)
		if err != nil {
			c.send <- Message{
				Sender:  "Server",
				Content: "Error saving file: " + err.Error(),
				Type:    "text",
			}
			return
		}
		defer file.Close()

		_, err = io.Copy(file, c.fileBuffer)
		if err != nil {
			c.send <- Message{
				Sender:  "Server",
				Content: "Error saving file: " + err.Error(),
				Type:    "text",
			}
			return
		}

		c.send <- Message{
			Sender:  "Server",
			Content: fmt.Sprintf("File saved as %s", savePath),
			Type:    "text",
		}

		// Reset file transfer state
		c.fileBuffer.Reset()
		c.receivingFile = false
		c.fileSize = 0
		c.receivedSize = 0
		c.fileName = ""
	}
}
