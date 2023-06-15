package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleMessages(ws *websocket.Conn) {
	for {
		_, p, err := ws.ReadMessage()

		if err != nil {
			log.Println("Failed to read message from server")
			return
		}

		log.Println(string(p))

		if err := ws.WriteMessage(1, []byte("Welcome")); err != nil {
			log.Println("Failed to send message to client")
			return
		}
	}
}

func handleWs(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		panic("Could not upgrade websocket.")
	}

	handleMessages(ws)
}

func main() {
	port := ":8080"
	log.Println("Listening on port", port)

	http.HandleFunc("/", handleWs)

	log.Fatal(http.ListenAndServe(port, nil))
}
