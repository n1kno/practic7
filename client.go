package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Не удалось подключиться: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	stdinReader := bufio.NewReader(os.Stdin)
	connReader := bufio.NewReader(conn)

	fmt.Print("Введите ваш ник: ")
	nick, _ := stdinReader.ReadString('\n')
	nick = strings.TrimSpace(nick)
	if nick == "" {
		fmt.Println("Ник не может быть пустым")
		return
	}

	fmt.Fprintln(conn, nick)

	// Ответ сервера на регистрацию
	resp, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Println("Соединение разорвано")
		return
	}
	resp = strings.TrimSpace(resp)
	if strings.HasPrefix(resp, "ERROR") {
		fmt.Println("Ошибка сервера:", resp)
		return
	}
	if resp != "OK" {
		fmt.Println("Неожиданный ответ:", resp)
		return
	}

	fmt.Printf("Подключены как %s. Ожидание сообщений...\n", nick)

	// Горутина для приёма сообщений
	go func() {
		for {
			line, err := connReader.ReadString('\n')
			if err != nil {
				fmt.Println("\nСоединение с сервером потеряно.")
				os.Exit(0)
			}
			line = strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(line, "MESSAGE "):
				fmt.Printf("\r[Сообщение] %s\n> ", strings.TrimPrefix(line, "MESSAGE "))
			case strings.HasPrefix(line, "ERROR"):
				fmt.Printf("\r[Ошибка] %s\n> ", line)
			case line == "SENT":
				fmt.Printf("\r[Отправлено]\n> ")
			default:
				fmt.Printf("\r[Сервер] %s\n> ", line)
			}
		}
	}()

	fmt.Println("Введите получателя и сообщение. Для выхода введите 'exit' как получателя.")
	for {
		fmt.Print("Кому (ник): ")
		recipient, _ := stdinReader.ReadString('\n')
		recipient = strings.TrimSpace(recipient)
		if recipient == "exit" {
			fmt.Println("До свидания!")
			return
		}
		if recipient == "" {
			continue
		}

		fmt.Print("Сообщение: ")
		msg, _ := stdinReader.ReadString('\n')
		msg = strings.TrimSpace(msg)

		fmt.Fprintf(conn, "SEND %s\n", recipient)
		fmt.Fprintf(conn, "%s\n", msg)
	}
}
