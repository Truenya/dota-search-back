package net

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/handlers"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{CheckOrigin: func(_ *http.Request) bool { return true }} // use default options

func WsHandler(g *gin.Context) {
	log.Debugf("[websocket] Got connection with addr: %v", g.ClientIP())

	c, err := upgrader.Upgrade(g.Writer, g.Request, nil)
	if err != nil {
		log.Warn("Upgrade to websocket failed: ", err)
		return
	}

	am := cache.GetAllM()
	for id, mes := range am {
		if time.Since(mes.Timestamp) >= time.Hour*72 {
			delete(am, id)
		}
	}

	for _, m := range am {
		wsm := data.FromMessage(m)

		err := c.WriteJSON(wsm)
		if err != nil {
			log.Errorf("[websocket] failed to send: %v", err)
		}
	}

	au := cache.GetAllU()
	for _, u := range au {
		wsm := data.FromPlayer(u)
		log.Debugln("[websocket]", wsm)

		err := c.WriteJSON(wsm)
		if err != nil {
			log.Errorf("[websocket] failed to send: %v", err)
		}
	}

	ac := cache.GetAllC()
	for _, co := range ac {
		wsm := data.FromCommand(co)
		log.Debugln("[websocket]", wsm)

		err := c.WriteJSON(wsm)
		if err != nil {
			log.Errorf("[websocket] failed to send: %v", err)
		}
	}

	manager.AddClient(c, g.ClientIP())
}

var manager WsManager

// var m sync.RWMutex

func Init() {
	manager.clients = make(map[*Client]bool)
	manager.toClients = make(chan *data.WsMessage)
	manager.fromClients = make(chan *data.WsMessage)
	manager.register = make(chan *Client)
	manager.unregister = make(chan *Client)

	go manager.work()
}

type WsManager struct {
	clients     map[*Client]bool
	toClients   chan *data.WsMessage
	fromClients chan *data.WsMessage
	register    chan *Client
	unregister  chan *Client
}

func (wm *WsManager) AddClient(c *websocket.Conn, ip string) {
	nc := NewClient(c, ip)
	nc.Manager = wm

	go nc.Read()
	go nc.Write()
	go manager.Listen()
	wm.register <- nc
}

func SendToWS(msg data.WsMessage) {
	manager.toClients <- &msg
}

func (wm *WsManager) SendAll(msg data.WsMessage) {
	wm.toClients <- &msg
}

func (wm *WsManager) Listen() {
	for msg := range wm.fromClients {
		var result handlers.ActionDid

		if msg.Action == "delete" {
			result = handlers.DeleteByKey(msg.IP)
		} else { // Currently add only
			if msg.Type == "command" {
				result = handlers.AddCommand(msg.ToCommand())
			} else if msg.Type == "player" {
				result = handlers.AddPlayer(msg.ToPlayer())
			}
		}

		if result == handlers.TooManyRequests {
			prev, _ := cache.GetI(msg.IP)
			Data := fmt.Sprintf("Too many requests, wait for %.2f seconds before try again",
				(handlers.TooEarly - time.Since(prev)).Seconds())

			msg.Type = "error"
			msg.Data = Data
			log.Warnf("Got too many requests from %s", msg.IP)
		}

		if result == handlers.NotFound {
			msg.Type = "error"
			msg.Data = msg.Type + " to delete not found"
			log.Warnf("%s trying to delete not existing entity", msg.IP)
		}

		wm.SendAll(*msg)
	}
}

func (wm *WsManager) work() {
	for {
		select {
		case client := <-wm.register:
			log.Debugf("[websocket] Client %s registered", client.IP)

			wm.clients[client] = true
		case client := <-wm.unregister:
			if _, ok := wm.clients[client]; ok {
				log.Debugf("[websocket] Client %s unregistered", client.IP)
				delete(wm.clients, client)
				close(client.Send)
			}

			log.Debugf("[websocket] Client %s not found", client.IP)
		case message := <-wm.toClients:
			log.Debugf("[websocket] Got message to everyone: %v", *message)

			for client := range wm.clients {
				select {
				case client.Send <- *message:
				default:
					close(client.Send)
					delete(wm.clients, client)
				}
			}
		}
	}
}
