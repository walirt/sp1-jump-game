package main

import (
	"bufio"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Concurrency and queue management variables
var (
	maxConcurrency = 3
	semaphore      = make(chan struct{}, maxConcurrency) // Semaphore to limit concurrency
	waitingQueue   = make([]*WaitingCommand, 0)          // Queue for waiting commands
	queueMutex     sync.Mutex                            // Mutex to protect the queue
)

// WaitingCommand represents a command waiting in the queue
type WaitingCommand struct {
	conn    *websocket.Conn
	cmdName string
	args    []string
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

		// Attempt to acquire a slot or add to queue
		select {
		case semaphore <- struct{}{}: // Slot available, execute immediately
			go executeCommand(conn, cmdName, args)
		default: // No slot available, add to waiting queue
			queueMutex.Lock()
			waitingQueue = append(waitingQueue, &WaitingCommand{conn, cmdName, args})
			position := len(waitingQueue)
			queueMutex.Unlock()
			conn.WriteMessage(websocket.TextMessage, []byte("WAITING: "+strconv.Itoa(position)))
		}
	}
}

func executeCommand(conn *websocket.Conn, cmdName string, args []string) {
	defer func() {
		<-semaphore // Release the slot
		checkWaitingQueue()
	}()

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

// checkWaitingQueue processes the next command in the queue when a slot becomes available
func checkWaitingQueue() {
	queueMutex.Lock()
	if len(waitingQueue) > 0 {
		nextCmd := waitingQueue[0]
		waitingQueue = waitingQueue[1:]
		// Update waiting positions for remaining commands
		for i, cmd := range waitingQueue {
			cmd.conn.WriteMessage(websocket.TextMessage, []byte("WAITING: "+strconv.Itoa(i+1)))
		}
		queueMutex.Unlock()
		// Execute the next command
		semaphore <- struct{}{} // Acquire a slot for the next command
		go executeCommand(nextCmd.conn, nextCmd.cmdName, nextCmd.args)
	} else {
		queueMutex.Unlock()
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
