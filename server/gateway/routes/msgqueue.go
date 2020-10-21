package routes

import (
	"context"
	"log"
	"strucim/gateway/google"

	"cloud.google.com/go/pubsub"
	"github.com/kataras/iris/websocket"
	"github.com/kataras/neffos"
)

var serverEvents = websocket.Namespaces{
	"identify.v1": websocket.Events{

		websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			// get the Iris' `Context`.
			// ctx := websocket.GetContext(nsConn.Conn)
			log.Printf("[%s] connected to namespace [%s] with IP []",
				nsConn, msg.Namespace)

			return nil
		},

		websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
			return nil
		},

		"publish": func(nsConn *websocket.NSConn, msg websocket.Message) error {
			topic := msg.Namespace

			google.Publish(topic, msg.Body)
			log.Printf("[%s] publish to topic %q msg:%s", nsConn, topic, msg.Body)

			nsConn.Emit("publish", msg.Body)
			return nil
		},

		"subscribe": func(nsConn *websocket.NSConn, msg websocket.Message) error {
			topic := msg.Namespace

			google.Subscribe(topic, func(ctx context.Context, msg *pubsub.Message) (bool, error) {
				ok := true
				log.Printf("[%s] topic %q received message %s", nsConn, topic, string(msg.Data))
				nsConn.Emit(topic, msg.Data)
				msg.Ack()
				return ok, nil
			})

			return nil
		},
	},
}

// MsgQueue returns a websocket handler
func MsgQueue() *neffos.Server {
	return websocket.New(websocket.DefaultGorillaUpgrader, serverEvents)
}
