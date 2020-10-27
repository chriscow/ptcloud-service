package gateway

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/chriscow/strucim/internal/messages"

	"github.com/gorilla/websocket"
)

func WSHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[WSHandler] called")
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	log.Println("[WSHandler] upgrading...")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		writeError(w, 500, "Failed to upgrade websocket", err)
		return
	}
	defer conn.Close()

	log.Println("[WSHandler] read loop starting")

	for {

		_, b, err := conn.ReadMessage()
		if err != nil {
			log.Println("ReadMessage error:", err)
			break
		}

		msg := &messages.WSEvent{}
		if err := json.Unmarshal(b, &msg); err != nil {
			log.Println("Failed to unmarshal message")
			break
		}

		switch msg.MsgType {
		case "WSHeartbeat":
			heartbeat := &messages.WSHeartbeat{}
			if err := json.Unmarshal(msg.Message, &heartbeat); err != nil {
				log.Println("Failed to unmarshal heartbeat")
				break
			}
		}

		// if err := conn.WriteMessage(msgType, msg); err != nil {
		// 	log.Println("WriteMessage error:", err)
		// 	break
		// }
	}
}
