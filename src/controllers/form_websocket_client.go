package controllers

import (
	"bytes"
	"context"
	"strings"
	"time"

	websocket "github.com/gorilla/websocket"
	models "github.com/nocodeleaks/quepasa/models"
	log "github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	OnClose func(int, string) error

	Cancel func()

	ctx context.Context
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(pairing models.QpWhatsappPairing) {
	defer func() {
		c.hub.unregister <- c
		// c.conn.Close() // issue on form synchronize qrcode
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseNoStatusReceived,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Errorf("(websocket): unexpected, %s", err.Error())
			}
			return
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		normalized := string(message)
		normalized = strings.ToLower(normalized)
		normalized = strings.TrimSpace(normalized)
		log.Infof("(websocket): message received: %s", normalized)

		if normalized == "start" {

			go c.QrCodeScanner(pairing)
		} else {
			log.Warnf("(websocket): received unknown msg: %s", normalized)
		}
	}
}

func (c *Client) QrCodeScanner(pairing models.QpWhatsappPairing) {
	out := make(chan []byte)
	defer close(out)
	go func() {
		for qrcode := range out {
			c.hub.broadcast <- qrcode
		}
	}()

	// show qrcode
	err := models.SignInWithQRCode(c.ctx, pairing, out)
	if err != nil {
		if strings.Contains(err.Error(), "time") {
			c.hub.broadcast <- []byte("timeout")
		} else {
			log.Errorf("unknown error on sign in with qrcode: %s", err.Error())
		}
	} else {
		log.Info("complete to read qr code on websocket")
		c.hub.broadcast <- []byte("complete")
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func WebSocketStart(pairing models.QpWhatsappPairing, conn *websocket.Conn) {
	execContext, cancel := context.WithCancel(context.Background())

	hub := newHub()
	go hub.run()

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), ctx: execContext}
	client.hub.register <- client
	client.OnClose = client.conn.CloseHandler()
	client.conn.SetCloseHandler(client.OnWebSocketClosed)
	client.Cancel = cancel

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump(pairing)
}

func (c *Client) OnWebSocketClosed(code int, text string) error {
	c.Cancel()
	return c.OnClose(code, text)
}
