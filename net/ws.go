package net

import (
	"encoding/json"
	"strings"
	"time"

	col "github.com/Truenya/dota-search-back/condlog"
	"github.com/Truenya/dota-search-back/data"
	"github.com/gorilla/websocket"
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
	// maxMessageSize = 512.
)

type Client struct {
	Conn    *websocket.Conn
	Send    chan data.WsMessage
	IP      string
	Manager *WsManager
}

func NewClient(c *websocket.Conn, ip string) *Client {
	return &Client{
		Conn: c,
		Send: make(chan data.WsMessage, 256),
		IP:   ip,
	}
}

func (c *Client) Read() {
	defer func() {
		log.Debugf("[websocket] Read for %s ended", c.IP)
		c.Manager.unregister <- c
		c.Conn.Close()
	}()
	col.TraceCond(c.Conn.SetReadDeadline(time.Now().Add(pongWait)))
	c.Conn.SetPongHandler(func(string) error { col.TraceCond(c.Conn.SetReadDeadline(time.Now().Add(pongWait))); return nil })

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Warnf("[websocket] Read failed: %s", err)
			}

			return
		}

		var p data.WsMessage

		err = json.Unmarshal(msg, &p)
		if err != nil {
			log.Errorf("[websocket] Message parsing failed: %s", err)
			log.Debugf("[websocket] Bad message is: %s", string(msg))

			return
		}

		p.IP = c.IP
		c.Manager.fromClients <- &p

		col.TraceCond(c.Conn.SetReadDeadline(time.Now().Add(pongWait)))
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Debugf("[websocket] Write for %s ended", c.IP)
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case p, ok := <-c.Send:
			col.WarnCond("[websocket] Failed to set write deadline: %v", c.Conn.SetWriteDeadline(time.Now().Add(writeWait)))

			if !ok {
				// Channel closed
				col.TraceCond(c.Conn.WriteMessage(websocket.CloseMessage, []byte{}))

				return
			}

			err := c.Conn.WriteJSON(&p)
			if err != nil {
				if !strings.Contains(err.Error(), "going away") {
					log.Warn("[websocket] Write failed:", err)
				}

				return
			}
		case <-ticker.C:
			col.WarnCond("[websocket] Failed to set write deadline: %v", c.Conn.SetWriteDeadline(time.Now().Add(writeWait)))

			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Errorf("[websocket] Ping failed: %v", err)
				return
			}
		}
	}
}
