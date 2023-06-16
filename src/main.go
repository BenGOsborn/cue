package main

import (
	"log"
	"net/http"

	"github.com/bengosborn/cue/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var port = ":8080"
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Receive messages from a connection
func receiveMsg(id string, connections *utils.Connections, messages chan<- utils.Message) {
	for {
		// Read and process messages
		if ok, err := connections.Apply(id, func(id string, conn *websocket.Conn) error {
			_, p, err := conn.ReadMessage()

			if err != nil {
				return err
			}

			msg := string(p)
			log.Println(id + ": " + msg)
			messages <- utils.Message{Id: id, Message: msg}

			return nil
		}); !ok || err != nil {
			log.Println("Connection failed.")
			connections.Remove(id)
			return
		}
	}
}

// Broadcast message to all connections
func broadcastMsg(connections *utils.Connections, messages <-chan utils.Message) {
	for msg := range messages {
		connections.ForEach(func(id string, conn *websocket.Conn) error {
			if err := conn.WriteMessage(1, []byte(id+": "+msg.Message)); err != nil {
				log.Println("Failed to send message to user", id)
			} else {
				log.Println("Sent message to user", id)
			}

			return nil
		})
	}
}

// Handle incoming connection
func handleWs(connections *utils.Connections, messages chan<- utils.Message) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Println("Could not upgrade websocket.")
			return
		}

		// Add connection to connection pool
		id := uuid.NewString()
		connections.Add(id, conn)

		receiveMsg(id, connections, messages)
	}
}

func main() {
	messages := make(chan utils.Message)
	connections := utils.NewConnections()

	http.HandleFunc("/", handleWs(connections, messages))

	log.Println("Listening on port", port)

	go broadcastMsg(connections, messages)
	log.Fatal(http.ListenAndServe(port, nil))
}
