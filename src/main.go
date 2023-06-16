package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var port = ":8080"
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var messages = make(chan Message)

var users = make(map[string]*websocket.Conn)
var userMutex sync.Mutex

type Message struct {
	userId  string
	message string
}

// **** I need to add mutexes to all of these elements to ensure only one access for writing and one for read (make a separate struct for this)

// Receive messages from a connection
func receiveMsg(userId string, messages chan<- Message) {
	// Get the user connection
	conn, ok := users[userId]

	if !ok {
		log.Println("Could not lookup user connection")
		return
	}

	for {
		// Read the message
		_, p, err := conn.ReadMessage()

		if err != nil {
			log.Println("Failed to read message from server")
			delete(users, userId)
			return
		}

		// Print the message
		log.Println(userId + ": " + string(p))
		messages <- Message{userId: userId, message: string(p)}
	}
}

// Broadcast message to all connections
func broadcastMsg(messages <-chan Message) {
	for msg := range messages {
		for userId, conn := range users {
			if err := conn.WriteMessage(1, []byte(userId+": "+msg.message)); err != nil {
				log.Println("Failed to send message to user")
			}
		}
	}
}

// Handle incoming connection
func handleWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Could not upgrade websocket.")
		return
	}

	// Add user to connection pool
	userId := uuid.NewString()
	users[userId] = ws

	receiveMsg(userId, messages)
}

func main() {
	http.HandleFunc("/", handleWs)

	log.Println("Listening on port", port)

	go broadcastMsg(messages)
	log.Fatal(http.ListenAndServe(port, nil))
}
