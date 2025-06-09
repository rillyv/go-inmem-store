package main

import (
	"bufio"
	"fmt"
	"go-inmem-store/handler"
	"go-inmem-store/store"
	"log"
	"net"
)

func main() {
	inMemoryStore := store.NewInMemoryStore()
	go inMemoryStore.RunTTLCleanup()

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Unable to listen to port 8080", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Unable to accept a connection", err)
		}

		go handleConnection(conn, inMemoryStore)
	}
}

func handleConnection(conn net.Conn, s *store.InMemoryStore) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		handler.HandleCommand(conn, line, s)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Connection error:", err)
	}
}
