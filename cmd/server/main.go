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

func client_create(IP string, Port int, message string) {
	conn, _ := net.Dial("tcp", fmt.Sprintf("%s:%d", IP, Port))

	client_reader := bufio.NewReader(conn)

	conn.Write([]byte(message))

	response, _ := client_reader.ReadString('\n')

	log.Print(response)

	defer conn.Close()
}

func server_create(s server) {
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

func main() {
	s := server{
		IP:   "localhost",
		Port: 8080,
	}

	go server_create(s)
	client_create(
		s.IP,
		s.Port,
		"Abdeljalil !!!!!\n",
	)

}
