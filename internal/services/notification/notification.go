package notification

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Client is a higher level wrapper around a websocket connection
type Client struct {
	conn *websocket.Conn
	ch   chan []byte
	sub  string

	Messages chan []byte
}

// Connect to the specified host and path using a websocket
func (c *Client) Connect(host, path, subName string) error {
	u := url.URL{Scheme: "ws", Host: host, Path: path}
	c.sub = subName

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.conn = conn

	c.conn.SetPingHandler(func(appData string) error {
		log.Println("[ping]", appData)
		return nil
	})
	c.conn.SetPongHandler(func(appData string) error {
		log.Println("[pong]", appData)
		return nil
	})

	c.Messages = make(chan []byte)

	go c.readPump()

	// write a heartbeat ping every second while we wait for results
	go c.heartbeat()

	return err
}

// Subscribe tells the server what topic to get notifications for
func (c *Client) Subscribe(topic string) error {
	log.Println("Sending subscribe request for", topic)
	msg := strings.Join([]string{"subscribe", topic}, ":")

	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (c *Client) readPump() error {
	for {
		log.Println("Wating for ReadMessage...")
		_, message, err := c.conn.ReadMessage()
		if err != nil {

			log.Println("closing result channel")
			close(c.Messages)

			if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				return nil
			}

			return fmt.Errorf("[readPump] read error: %v", err)
		}

		c.Messages <- message
	}
}

func (c *Client) heartbeat() error {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case t := <-ticker.C:
			err := c.conn.WriteMessage(websocket.PingMessage, []byte(t.String()))
			if err != nil {
				return fmt.Errorf("[wsHeartbeat] Error writing heartbeat: %v", err)
			}

		case <-interrupt:
			log.Println("interrupt")
			if err := c.Close(); err != nil {
				return fmt.Errorf("[interrupt] Received error from wsClose: %v", err)
			}
		}
	}
}

// Close performs a graceful close
func (c *Client) Close() error {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		if websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Println("normal close")
			return nil
		}

		return err
	}

	log.Println("waiting for done to close or timeout")

	select {
	case <-c.Messages: // wait for the read goroutine to exit
	case <-time.After(time.Second): // or timeout after a second
	}

	log.Println("closed")

	return nil
}
