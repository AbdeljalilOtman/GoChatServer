package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type server struct {
	IP   string
	Port int
}

func main() {
	s := server{
		IP:   "localhost",
		Port: 8080,
	}

	// creating the server
	server_listner, _ := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))

	defer server_listner.Close()
	for {
		conn, _ := server_listner.Accept()

		reader := bufio.NewReader(conn)

		message, _ := reader.ReadString('\n')
		log.Print(message)

		message = fmt.Sprintf("%s fuck i got it", message[:len(message)-1])
		message += "\n"

		conn.Write([]byte(message))
	}

}
