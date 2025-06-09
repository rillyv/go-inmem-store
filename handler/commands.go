package handler

import (
	"fmt"
	"go-inmem-store/store"
	"net"
	"strconv"
	"strings"
)

func HandleCommand(conn net.Conn, line string, s *store.InMemoryStore) {
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

		fmt.Fprintln(conn, s.Set(parts[1], parts[2]))
	case "GET":
		if len(parts) != 2 {
			fmt.Fprintln(conn, "ERR usage: GET key")
			return
		}

		fmt.Fprintln(conn, s.Get(parts[1]))
	case "DEL":
		if len(parts) != 2 {
			fmt.Fprintln(conn, "ERR usage: DEL key")
			return
		}
		fmt.Println("hi")

		fmt.Fprintln(conn, s.Delete(parts[1]))
	case "PING":
		if len(parts) != 1 {
			fmt.Fprintln(conn, "ERR usage: PING")
			return
		}

		fmt.Fprintln(conn, "PONG")
	case "EXPIRE":
		if len(parts) != 3 {
			fmt.Fprintln(conn, "ERR usage: EXPIRE key ttl")
			return
		}

		ttlInSeconds, err := strconv.Atoi(parts[2])
		if err != nil {
			fmt.Fprintln(conn, "ERR ttl must be an integer")
		}

		s.Expire(parts[1], ttlInSeconds)
	case "TTL":
		if len(parts) != 2 {
			fmt.Fprintln(conn, "ERR usage: TTL key")
			return
		}

		fmt.Fprintln(conn, s.GetTtl(parts[1]))
	case "KEYS":
		if len(parts) != 3 {
			fmt.Fprintln(conn, "ERR usage: KEYS offset limit")
			return
		}

		offset, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Fprintln(conn, "ERR offset must be an integer")
		}

		limit, err := strconv.Atoi(parts[2])
		if err != nil {
			fmt.Fprintln(conn, "ERR limit must be an integer")
		}

		keys := s.Keys(offset, limit)

		for _, v := range keys {
			fmt.Fprintln(conn, v)
		}
	case "FLUSHALL":
		if len(parts) != 1 {
			fmt.Fprintln(conn, "ERR usage: FLUSHALL")
		}

		s.FlushAll()
	default:
		fmt.Fprintln(conn, "ERR unknown command")
	}
}
