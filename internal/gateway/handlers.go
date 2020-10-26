package gateway

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/chriscow/strucim/internal/messages"
)

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	w.WriteHeader(status)
	log.Printf("%s: %v", msg, err)
	fmt.Fprintf(w, "%s: %v", msg, err)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {

	req := messages.IdentifyRequest{}
	body, _ := ioutil.ReadAll(r.Body)

	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, 400, "Invalid request JSON", err)
		return
	}

	csvBytes, err := base64.StdEncoding.DecodeString(req.File)
	if err != nil {
		writeError(w, 400, "Failed to decode file: %s", err)
		return
	}

	sr := bytes.NewReader(csvBytes)
	bucket := os.Getenv("IDENTIFY_BUCKET")
	filename := fmt.Sprintf("pointcloud-%d", time.Now().Unix())

	if err := storeFile(bucket, filename, sr); err != nil {
		writeError(w, 500, "Failed to write to storage", err)
		return
	}

	msg := messages.IdentifyResponse{
		Bucket:   bucket,
		Filename: filename,
		Status:   "init",
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		writeError(w, 500, "Failed to marshal pubsub message", err)
		return
	}

	if err := publish(os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"), msgJSON); err != nil {
		writeError(w, 500, "Failed to publish identify job", err)
		return
	}

	fmt.Fprint(w, "OK")
}

// ServerEvents returns a websocket handler
// var ServerEvents = websocket.Namespaces{

// 	"identify.v1": websocket.Events{

// 		websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
// 			// get the Iris' `Context`.
// 			ctx := websocket.GetContext(nsConn.Conn)
// 			log.Printf("[%s] connected to namespace [%s] with IP [%s]",
// 				nsConn, msg.Namespace, ctx.RemoteAddr())

// 			return nil
// 		},

// 		websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
// 			log.Printf("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
// 			count := nsConn.Conn.Server().GetTotalConnections()
// 			fmt.Println("total connections:", count)
// 			return nil
// 		},

// 		"publish": func(nsConn *websocket.NSConn, msg websocket.Message) error {
// 			topic := msg.Namespace

// 			google.Publish(topic, msg.Body)
// 			log.Printf("[%s] publish to topic %q msg:%s", nsConn, topic, msg.Body)

// 			nsConn.Emit("publish", msg.Body)
// 			return nil
// 		},

// 		"subscribe": func(nsConn *websocket.NSConn, msg websocket.Message) error {
// 			topic := msg.Namespace

// 			log.Printf("[%s] subscribing to topic %q", nsConn, topic)

// 			err := google.Subscribe(topic, func(ctx context.Context, msg *pubsub.Message) (bool, error) {
// 				ok := true
// 				log.Printf("[%s] topic %q received message %s", nsConn, topic, string(msg.Data))

// 				wsmsg := websocket.Message{Namespace: topic, Event: topic, Body: msg.Data}
// 				nsConn.Conn.Server().Broadcast(nil, wsmsg)
// 				msg.Ack()
// 				return ok, nil
// 			})

// 			if err != nil {
// 				nsConn.Emit("error", []byte(err.Error()))
// 				return err
// 			}

// 			nsConn.Emit("subscribe", []byte("subscribed"))

// 			return nil
// 		},
// 	},
// }
