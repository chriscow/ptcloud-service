package commands

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/chriscow/strucim/internal/messages"
	"github.com/gorilla/websocket"

	"github.com/urfave/cli/v2"
)

// Identify reads in the file given as an argument.
func Identify(ctx *cli.Context) error {

	if err := validateArgs(ctx); err != nil {
		return err
	}

	// connect the websocket for notification when identification is complete
	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()

	c.SetPingHandler(func(appData string) error {
		log.Println("[ping]", appData)
		return nil
	})
	c.SetPongHandler(func(appData string) error {
		log.Println("[pong]", appData)
		return nil
	})

	rc, err := status(c)
	if err != nil {
		return err
	}

	filename := ctx.Args().Get(0)
	endpoint := os.Getenv("LOCATOR_ENDPOINT")
	url := fmt.Sprintf("http://%s/v1/identify", endpoint)

	encoded, _ := encodeFile(filename, endpoint, url)

	req, err := json.Marshal(&messages.IdentifyRequest{File: encoded})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Server return an error: %s", string(body))
	}

	log.Printf("[Identify] waiting for result channel")
	result, ok := <-rc
	if !ok {
		// channel got closed
		log.Fatal("[Identify] result channel got closed")
	}

	fmt.Println(result)

	return wsClose(c, nil)
}

func validateArgs(ctx *cli.Context) error {
	if ctx.Args().Len() == 0 {
		return cli.Exit("You forgot to pass in the file name", 1)
	}

	if ctx.Args().Len() > 1 {
		return cli.Exit("Too many arguments were passed in. Just pass the filename", 1)
	}

	return nil
}

func encodeFile(filename, endpoint, uri string) (string, error) {

	csv, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("Failed to read file: %v", err)
	}

	return base64.StdEncoding.EncodeToString(csv), nil
}

type statusResult struct {
	Result string
	Err    error
}

func connect() (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: os.Getenv("LOCATOR_ENDPOINT"), Path: "/v1/identify"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return c, err
}

func status(c *websocket.Conn) (<-chan string, error) {

	result := make(chan string)

	go readPump(c, result)

	// write a heartbeat ping every second while we wait for results
	go heartbeat(c, result)

	return result, subscribe(c)
}

func subscribe(c *websocket.Conn) error {
	topic := os.Getenv("IDENTIFY_POINTCLOUD_STATUS_TOPIC")
	log.Println("Sending subscribe request for", topic)
	msg := strings.Join([]string{"subscribe", topic}, ":")

	return c.WriteMessage(websocket.TextMessage, []byte(msg))
}

func readPump(c *websocket.Conn, result chan string) error {
	for {
		log.Println("Wating for ReadMessage...")
		_, message, err := c.ReadMessage()
		if err != nil {

			log.Println("closing result channel")
			close(result)

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil
			}

			return fmt.Errorf("[readPump] read error: %v", err)
		}

		result <- string(message)
	}
}

func heartbeat(c *websocket.Conn, done chan string) error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.PingMessage, []byte(t.String()))
			if err != nil {
				return fmt.Errorf("[wsHeartbeat] Error writing heartbeat: %v", err)
			}

		case <-interrupt:
			log.Println("interrupt")
			if err := wsClose(c, done); err != nil {
				return fmt.Errorf("[interrupt] Received error from wsClose: %v", err)
			}
		}
	}
}

func wsClose(c *websocket.Conn, done chan string) error {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("normal close")
			return nil
		}

		return err
	}

	log.Println("waiting for done to close or timeout")

	select {
	case <-done: // wait for the read goroutine to exit
	case <-time.After(time.Second): // or timeout after a second
	}

	log.Println("closed")

	return nil
}
