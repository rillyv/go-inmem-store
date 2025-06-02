package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var mu = sync.RWMutex{}

var store = make(map[string]string)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Unable to listen to port 8080", err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("Unable to accept a connection", err)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Fields(line)
		if len(parts) == 0 {
			fmt.Fprintln(conn, "ERR empty command")
			return
		}

		switch strings.ToUpper(parts[0]) {
		case "SET":
			if len(parts) != 3 {
				fmt.Fprintln(conn, "ERR usage: SET key value")
				return
			}

			mu.Lock()
			store[parts[1]] = parts[2]
			mu.Unlock()

			fmt.Fprintln(conn, "OK")
		case "GET":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR usage: GET key")
				return
			}

			mu.RLock()
			val, ok := store[parts[1]]
			mu.RUnlock()

			if !ok {
				fmt.Fprintln(conn, "NULL")
			} else {
				fmt.Fprintln(conn, val)
			}
		case "DEL":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR usage: DEL key")
				return
			}

			mu.Lock()
			delete(store, parts[1])
			mu.Unlock()

			fmt.Fprintln(conn, "OK")
		case "PING":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR usage: PING")
				return
			}
			fmt.Fprintln(conn, "PONG")
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Connection error:", err)
	}
}
