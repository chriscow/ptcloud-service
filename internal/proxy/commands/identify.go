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

	c, err := connect()
	if err != nil {
		return err
	}
	defer c.Close()

	done, err := status(c)
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

	log.Printf("waiting for done channel")
	result := <-done

	fmt.Println(result)

	return nil
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
	if err != nil {
		log.Fatal("dial:", err)
		return nil, err
	}

	return c, nil
}

func status(c *websocket.Conn) (<-chan string, error) {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	result := make(chan string)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			log.Println("Wating for ReadMessage...")
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			result <- string(message)
		}
	}()

	ticker := time.NewTicker(time.Second)

	// heartbeat
	go func() {
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				log.Println("Writing time heartbeat at ", t)

				hb, err := json.Marshal(&messages.WSHeartbeat{Timestamp: t.String()})
				if err != nil {
					result <- fmt.Sprintf("failed to marshal heartbeat message: %v", err)
					return
				}

				b, err := json.Marshal(&messages.WSEvent{
					MsgType: "WSHeartbeat",
					Message: hb,
				})
				if err != nil {
					result <- fmt.Sprintf("Failed to marshal heartbeat message envelope: %v", err)
					return
				}

				err = c.WriteMessage(websocket.BinaryMessage, b)
				if err != nil {
					result <- fmt.Sprintf("Received error on WriteMessage: %v", err)
					return
				}
			case <-interrupt:
				defer close(done)
				defer close(result)
				log.Println("interrupt")
				if err := wsClose(c, done); err != nil {
					log.Printf("Received error from wsClose: %v", err)
					return
				}
			}
		}
	}()

	return result, nil
}

func wsClose(c *websocket.Conn, done chan struct{}) error {
	log.Printf("[wsClose] called")

	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("[wsClose] error writing: %v", err)
		return err
	}

	log.Printf("[wsClose] waiting for done or time.Second")

	select {
	case <-done: // wait for the read goroutine to exit
	case <-time.After(time.Second): // or timeout after a second
	}

	log.Printf("[wsClose] done")

	return nil
}
