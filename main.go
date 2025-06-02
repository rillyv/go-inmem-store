package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mu = sync.RWMutex{}

var store = make(map[string]string)
var ttlStore = make(map[string]time.Time)

func main() {
	go runTTLCleanup()

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

func runTTLCleanup() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		expiredKeys := []string{}

		mu.RLock()
		for k, v := range ttlStore {
			if v.Unix() < time.Now().Unix() {
				expiredKeys = append(expiredKeys, k)
			}
		}
		mu.RUnlock()

		mu.Lock()
		for _, v := range expiredKeys {
			delete(store, v)
			delete(ttlStore, v)
		}
		mu.Unlock()
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
			continue
		}

		switch strings.ToUpper(parts[0]) {
		case "SET":
			if len(parts) != 3 {
				fmt.Fprintln(conn, "ERR usage: SET key value")
				continue
			}

			mu.Lock()
			store[parts[1]] = parts[2]
			mu.Unlock()

			fmt.Fprintln(conn, "OK")
		case "GET":
			if len(parts) != 2 {
				fmt.Fprintln(conn, "ERR usage: GET key")
				continue
			}

			mu.RLock()
			ttl, ok := ttlStore[parts[1]]
			mu.RUnlock()

			if ok && (ttl.Unix() < time.Now().Unix()) {
				mu.Lock()
				delete(store, parts[1])
				delete(ttlStore, parts[1])
				mu.Unlock()

				fmt.Fprintln(conn, "NULL")
				continue
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
				continue
			}

			mu.RLock()
			ttl, ok := ttlStore[parts[1]]
			mu.RUnlock()

			if ok && (ttl.Unix() < time.Now().Unix()) {
				mu.Lock()
				delete(store, parts[1])
				delete(ttlStore, parts[1])
				mu.Unlock()

				fmt.Fprintln(conn, "NULL")
				continue
			}

			mu.Lock()
			delete(store, parts[1])
			delete(ttlStore, parts[1])
			mu.Unlock()

			fmt.Fprintln(conn, "OK")
		case "PING":
			if len(parts) != 1 {
				fmt.Fprintln(conn, "ERR usage: PING")
				continue
			}
			fmt.Fprintln(conn, "PONG")
		case "EXPIRE":
			if len(parts) != 3 {
				fmt.Fprintln(conn, "ERR usage: EXPIRE key ttl")
				continue
			}

			// first check if the key exists
			mu.RLock()
			_, ok := store[parts[1]]
			if !ok {
				fmt.Fprintln(conn, "NULL")
			}
			mu.RUnlock()

			seconds, err := strconv.Atoi(parts[2])
			if err != nil {
				fmt.Fprintln(conn, "ERR ttl must be an integer")
			}

			mu.Lock()
			ttlStore[parts[1]] = time.Now().Add(time.Duration(seconds) * time.Second)
			mu.Unlock()

			fmt.Fprintln(conn, "OK")
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Connection error:", err)
	}
}
