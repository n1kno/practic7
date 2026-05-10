package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var (
	clients   = make(map[string]net.Conn)
	clientsMu sync.RWMutex
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	fmt.Println("Сервер запущен на :8080")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Ошибка подключения:", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Чтение ника
	nickLine, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	nick := strings.TrimSpace(nickLine)
	if nick == "" {
		fmt.Fprintln(conn, "ERROR: пустой ник")
		return
	}

	clientsMu.Lock()
	if _, exists := clients[nick]; exists {
		clientsMu.Unlock()
		fmt.Fprintln(conn, "ERROR: nickname already taken")
		return
	}
	clients[nick] = conn
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, nick)
		clientsMu.Unlock()
		fmt.Printf("Пользователь %s отключился\n", nick)
	}()

	fmt.Fprintln(conn, "OK")
	fmt.Printf("Пользователь %s подключился\n", nick)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "SEND ") {
			recipient := strings.TrimPrefix(line, "SEND ")
			recipient = strings.TrimSpace(recipient)
			if recipient == "" {
				fmt.Fprintln(conn, "ERROR: не указан получатель")
				continue
			}

			// Чтение строки сообщения
			msgLine, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			msg := strings.TrimSpace(msgLine)

			clientsMu.RLock()
			recipientConn, ok := clients[recipient]
			clientsMu.RUnlock()
			if !ok {
				fmt.Fprintf(conn, "ERROR: user %s not found\n", recipient)
			} else {
				fmt.Fprintf(recipientConn, "MESSAGE from %s: %s\n", nick, msg)
				fmt.Fprintln(conn, "SENT")
			}
		} else {
			fmt.Fprintln(conn, "ERROR: неизвестная команда")
		}
	}
}
