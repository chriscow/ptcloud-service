package identify

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"strucim/gateway/google"

	"cloud.google.com/go/pubsub"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/websocket"
)

type idRequest struct {
	CsvFile  string `json:"file"`
	Callback string `json:"callback"`
}

type idMessage struct {
	Bucket   string `json:"bucket"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Result   string `json:"result"`
}

// PointCloud is called when the client wants to identify a part using point
// cloud data. The file is sent embedded in a JSON wrapper as a CSV.
func PointCloud(ctx iris.Context) {

	req := &idRequest{}
	if err := ctx.ReadJSON(&req); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Failed to read JSON body").DetailErr(err))
		return
	}

	bucket, err := google.GetBucket(os.Getenv("IDENTIFY_BUCKET"))
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("GetBucket failed").DetailErr(err))
		return
	}

	csvBytes, err := base64.StdEncoding.DecodeString(req.CsvFile)
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Failed to decode file").DetailErr(err))
		return
	}

	csv := string(csvBytes)
	sr := strings.NewReader(csv)

	filename := fmt.Sprintf("pointcloud-%d", time.Now().Unix())
	if err := bucket.Store(filename, sr); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("bucket.Store() failed").DetailErr(err))
		return
	}

	msg := idMessage{
		Bucket:   os.Getenv("IDENTIFY_BUCKET"),
		Filename: filename,
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("json.Marshal() pubsub message failed").DetailErr(err))
		return
	}

	if err := google.Publish(os.Getenv("IDENTIFY_POINTCLOUD_TOPIC"), msgJSON); err != nil {
		ctx.StopWithProblem(iris.StatusBadRequest, iris.NewProblem().
			Title("Publish() failed").DetailErr(err))
		return
	}

	// TODO: Send JSON response with

}

// ServerEvents returns a websocket handler
var ServerEvents = websocket.Namespaces{

	"identify.v1": websocket.Events{

		websocket.OnNamespaceConnected: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			// get the Iris' `Context`.
			ctx := websocket.GetContext(nsConn.Conn)
			log.Printf("[%s] connected to namespace [%s] with IP [%s]",
				nsConn, msg.Namespace, ctx.RemoteAddr())

			return nil
		},

		websocket.OnNamespaceDisconnect: func(nsConn *websocket.NSConn, msg websocket.Message) error {
			log.Printf("[%s] disconnected from namespace [%s]", nsConn, msg.Namespace)
			count := nsConn.Conn.Server().GetTotalConnections()
			fmt.Println("total connections:", count)
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

			log.Printf("[%s] subscribing to topic %q", nsConn, topic)

			err := google.Subscribe(topic, func(ctx context.Context, msg *pubsub.Message) (bool, error) {
				ok := true
				log.Printf("[%s] topic %q received message %s", nsConn, topic, string(msg.Data))

				wsmsg := websocket.Message{Namespace: topic, Event: topic, Body: msg.Data}
				nsConn.Conn.Server().Broadcast(nil, wsmsg)
				msg.Ack()
				return ok, nil
			})

			if err != nil {
				nsConn.Emit("error", []byte(err.Error()))
				return err
			}

			nsConn.Emit("subscribe", []byte("subscribed"))

			return nil
		},
	},
}
