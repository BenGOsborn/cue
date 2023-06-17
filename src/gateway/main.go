package main

import (
	"log"
	"net/http"

	gwUtils "github.com/bengosborn/cue/gateway/src/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var port = ":8080"
var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Receive messages from a connection
func receiveMsg(id string, connections *gwUtils.Connections, messages chan<- gwUtils.Message) {
	for {
		// Read and process messages
		if ok, err := connections.Apply(id, func(id string, conn *websocket.Conn) error {
			_, p, err := conn.ReadMessage()

			if err != nil {
				return err
			}

			msg := string(p)
			log.Println(id + ": " + msg)
			messages <- gwUtils.Message{Id: id, Message: msg}

			return nil
		}); !ok || err != nil {
			log.Println("Connection failed.")
			connections.Remove(id)
			return
		}
	}
}

// Broadcast message to all connections
func broadcastMsg(connections *gwUtils.Connections, messages <-chan gwUtils.Message) {
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
func handleWs(connections *gwUtils.Connections, messages chan<- gwUtils.Message) func(w http.ResponseWriter, r *http.Request) {
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
	messages := make(chan gwUtils.Message)
	connections := gwUtils.NewConnections()

	http.HandleFunc("/", handleWs(connections, messages))

	log.Println("Listening on port", port)

	// **** So now we need to create a facade for a message queue (which supports fan out) and which sends all of its messages into one of these functions to be sorted
	// into another channel for where the message should go and what category it should be assigned ??? (this needs to be scalable)
	// Raw messagw from service -> message queue processing channel -> processed ready to be sent to message channel / broadcast to user channel / broadcast to all chanel

	// **** For efficiency, we will have in "express" Kafka channel which when the message is sent this server id is attached to it, and then when the message is sent back
	// it will go directly to a queue only listened to by this server which will process it - if the message does not belong to this user, we will add it to a global queue
	// which is listened to by all servers.
	// Also remove the broadcast to all - it will NOT work in a distributed environment and is not needed.

	for i := 0; i < 5; i++ {
		go broadcastMsg(connections, messages)
	}

	log.Fatal(http.ListenAndServe(port, nil))
}
