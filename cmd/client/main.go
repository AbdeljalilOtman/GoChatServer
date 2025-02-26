package main

import (
	"bufio"
	"log"
	"net"
)

func main() {
	conn, _ := net.Dial("tcp", "localhost:8080")

	client_reader := bufio.NewReader(conn)

	message := "Abdeljalil !!!!!\n"

	conn.Write([]byte(message))

	response, _ := client_reader.ReadString('\n')

	log.Print(response)

	defer conn.Close()

}
