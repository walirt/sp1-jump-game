package main

import (
	"bufio"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Connection close:", err)
			break
		}

		symbol := string(message)
		var cmdStr string

		if symbol == "GENERATE" {
			cmdStr = "cargo run --release -- --prove"
		} else {
			cmdStr = "cowsay"
		}

		parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}
		cmdName := parts[0]
		args := parts[1:]

		go executeCommand(conn, cmdName, args)
	}
}

func executeCommand(conn *websocket.Conn, cmdName string, args []string) {
	cmd := exec.Command(cmdName, args...)
	cmd.Dir = "../prove"

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Get stdout error:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR"))
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Println("Get stderr error:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR"))
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println("Start error:", err)
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR"))
		return
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}()

	err = cmd.Wait()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("ERROR"))
	} else {
		conn.WriteMessage(websocket.TextMessage, []byte("DONE"))
	}
}

func main() {
	http.HandleFunc("/proof", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade failed:", err)
			return
		}
		handleWebSocket(conn)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./index.html")
	})

	fsh := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fsh))

	log.Println("Server startup at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Server startup failed:", err)
	}
}
