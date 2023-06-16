package main

import (
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var port = ":8080"
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
var users = []*User{}
var messages = make(chan Message)

type User struct {
	conn *websocket.Conn
	id   string
}

type Message struct {
	message string
	user    *User
}

// Receive messages from a connection
func receiveMsg(user *User, messages chan<- Message) {
	for {
		_, p, err := user.conn.ReadMessage()

		if err != nil {
			// TODO add logic for removing connection from pool upon error
			log.Println("Failed to read message from server")
			return
		}

		log.Println(user.id + ": " + string(p))
		messages <- Message{user: user, message: string(p)}
	}
}

// Broadcast messages to all connections
func broadcastMsg(users *[]*User, messages <-chan Message) {
	for msg := range messages {
		for _, user := range *users {
			if err := user.conn.WriteMessage(1, []byte(user.id+": "+msg.message)); err != nil {
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

	user := User{conn: ws, id: uuid.NewString()}

	users = append(users, &user)
	receiveMsg(&user, messages)
}

func main() {
	http.HandleFunc("/", handleWs)

	log.Println("Listening on port", port)

	go broadcastMsg(&users, messages)
	log.Fatal(http.ListenAndServe(port, nil))
}
