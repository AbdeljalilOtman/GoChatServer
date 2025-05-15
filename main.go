package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/abdeljalil/GoChatServer/server"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	// Create and start the server
	s := server.NewServer(*port)
	fmt.Printf("Chat server running on port %d\n", *port)
	log.Fatal(s.Run())
}
