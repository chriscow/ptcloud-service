package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeError(w, 500, "Failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	fmt.Println("wsHandler closing connection")

	// for {

	// 	msgType, msg, err := conn.ReadJSON(&)
	// }
}
