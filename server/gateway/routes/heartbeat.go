package routes

import (
	"bytes"
	"encoding/json"
	"log"
	"time"

	"github.com/kataras/iris/websocket"
)

// const CHANNEL string := "heartbeat"

const (
	StatusOK = 100
)

type statusResp struct {
	Status    int   `json:'status'`
	Timestamp int64 `json:'timestamp'`
}

// Heartbeat returns JSON with the current system status and timestamp
// for monitoring
func Heartbeat(nsConn *websocket.NSConn, msg websocket.Message) error {

	// room.String() returns -> NSConn.String() returns -> Conn.String() returns -> Conn.ID()
	log.Printf("[%s] sent: %s", nsConn, string(msg.Body))

	status := statusResp{Status: StatusOK, Timestamp: time.Now().UTC().Unix()}

	// encode the status into json
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(status); err != nil {
		return err
	}

	// Write message back to the client message owner with:
	// nsConn.Emit("chat", msg)
	nsConn.Emit("heartbeat", buf.Bytes())

	// Write message to all except this client with:
	// nsConn.Conn.Server().Broadcast(nsConn, msg)
	return nil
}
