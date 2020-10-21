package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/kataras/iris/v12/websocket"
)

const (
	endpoint              = "ws://localhost:8080/v1/identify"
	namespace             = "identify.v1"
	dialAndConnectTimeout = 5 * time.Second
)

var wg *sync.WaitGroup

// this can be shared with the server.go's.
// `NSConn.Conn` has the `IsClient() bool` method which can be used to
// check if that's is a client or a server-side callback.
var clientEvents = websocket.Namespaces{
	namespace: websocket.Events{
		websocket.OnNamespaceConnected: func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("connected to namespace: %s", msg.Namespace)

			return nil
		},
		websocket.OnNamespaceDisconnect: func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("disconnected from namespace: %s", msg.Namespace)
			return nil
		},
		"publish": func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("publish: %s", string(msg.Body))
			wg.Done()
			return nil
		},
		"subscribe": func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("subscribe: %s", string(msg.Body))
			wg.Done()
			return nil
		},
		"error": func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("error: %s", string(msg.Body))
			wg.Done()
			return nil
		},
		namespace: func(c *websocket.NSConn, msg websocket.Message) error {
			log.Printf("[%s] received message: %s", namespace, string(msg.Body))
			wg.Done()
			return nil
		},
	},
}

func main() {

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(dialAndConnectTimeout))
	defer cancel()

	// username := "my_username"
	// dialer := websocket.GobwasDialer(websocket.GobwasDialerOptions{Header: websocket.GobwasHeader{"X-Username": []string{username}}})
	dialer := websocket.DefaultGobwasDialer
	client, err := websocket.Dial(ctx, dialer, endpoint, clientEvents)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer client.Close()

	c, err := client.Connect(ctx, namespace) // namespace is the topic identify.v1
	if err != nil {
		panic(err)
	}
	// defer c.Conn.Close()

	wg = &sync.WaitGroup{}
	wg.Add(4)
	c.Emit("publish", []byte("publish this please"))
	c.Emit("publish", []byte("publish this please #2"))

	fmt.Println("subscribing...")
	c.Emit("subscribe", []byte("ignored"))

	wg.Wait()

	fmt.Println("exiting")
	// c.Emit("heartbeat", []byte("Hello from Go client side!"))

	// fmt.Fprint(os.Stdout, ">> ")
	// scanner := bufio.NewScanner(os.Stdin)
	// for {
	// 	if !scanner.Scan() {
	// 		log.Printf("ERROR: %v", scanner.Err())
	// 		return
	// 	}

	// 	text := scanner.Bytes()

	// 	if bytes.Equal(text, []byte("exit")) {
	// 		if err := c.Disconnect(nil); err != nil {
	// 			log.Printf("reply from server: %v", err)
	// 		}
	// 		break
	// 	}

	// 	ok := c.Emit("heartbeat", text)
	// 	if !ok {
	// 		break
	// 	}

	// 	fmt.Fprint(os.Stdout, ">> ")
	// }
} // try running this program twice or/and run the server's http://localhost:8080 to check the browser client as well.
