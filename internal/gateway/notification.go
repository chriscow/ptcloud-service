package gateway

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/pubsub"
	"github.com/gorilla/websocket"
)

// TODO: This is a Notification service.
//
// Other services can publish a notification message
// { 'type':'notification', 'source':'identification-service',
// 		...
//		service specific message
// }
//
func NotifyHandler(w http.ResponseWriter, r *http.Request) {

	log.Println("[WSHandler] called")
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

	// conn.SetPingHandler(func(appData string) error {
	// 	log.Println("[ping]", appData)
	// 	return nil
	// })
	// conn.SetPongHandler(func(appData string) error {
	// 	log.Println("[pong]", appData)
	// 	return nil
	// })

	ctx, cancel := context.WithCancel(context.Background())

	for {

		_, b, err := conn.ReadMessage()
		if err != nil {

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Println("client closed")
				cancel()
				return
			}

			log.Println("ReadMessage error:", err)
			cancel()
			return
		}

		tokens := strings.Split(string(b), ":")
		if len(tokens) != 2 {
			conn.WriteMessage(websocket.TextMessage, []byte("Bad command"))
			cancel()
			return
		}

		command, param := tokens[0], tokens[1]

		switch command {
		case "subscribe":
			go Subscribe(ctx, cancel, "gateway", param, func(ctx context.Context, msg *pubsub.Message) (bool, error) {
				log.Println("Received message from pubsub")
				err := conn.WriteMessage(websocket.BinaryMessage, msg.Data)
				msg.Ack()
				return true, err
			})
			log.Println("Received subscribe:", param)
			break
		case "unsubscribe":
			log.Println("Received unsubscribe:", param)
			cancel()
			conn.WriteMessage(websocket.TextMessage, []byte("unsubscribe:OK"))
			break
		}
	}
}
