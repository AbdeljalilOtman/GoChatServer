package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/abdeljalil/GoChatServer/server"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	verbose := flag.Bool("v", false, "Enable verbose logging")
	flag.Parse()

	// Set up logging
	if *verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		// Less verbose logging
		log.SetFlags(log.Ldate | log.Ltime)
	}

	// Create and start the server
	s := server.NewServer(*port)
	fmt.Printf("Chat server running on port %d\n", *port)
	fmt.Println("Press Ctrl+C to stop the server")

	// Log to both console and file
	logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	log.Fatal(s.Run())
}
