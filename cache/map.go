package cache

import (
	"time"

	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
	cmap "github.com/orcaman/concurrent-map/v2"
	log "github.com/sirupsen/logrus"
)

type Result int

const (
	Stored Result = iota
	Updated
	Conflict
)

var m cmap.ConcurrentMap[string, data.Message]
var c cmap.ConcurrentMap[string, data.Command]
var p cmap.ConcurrentMap[string, data.Player]
var ip cmap.ConcurrentMap[string, time.Time]

func Init() {
	m = cmap.New[data.Message]()
	ip = cmap.New[time.Time]()
	p = cmap.New[data.Player]()
	c = cmap.New[data.Command]()

	InitMessages()
	InitPlayers()
	InitCommands()
	InitIps()
}

func InitMessages() {
	msgs := db.GetAllM()
	for _, mes := range msgs.Messages {
		m.Set(mes.Hash, mes)
	}
}

func InitPlayers() {
	ap := db.GetAllU()
	for _, pl := range ap.Users {
		p.Set(pl.IP, pl)
	}
}

func InitCommands() {
	ac := db.GetAllC()
	for _, com := range ac.Commands {
		c.Set(com.IP, com)
	}
}

func InitIps() {
	ips := db.GetAllI()
	for _, i := range ips {
		ip.Set(i, time.Now().Add(-30*time.Second))
	}

	log.Debugf("Got IP from db: %v", ips)
}

func StoreM(mes data.Message) Result {
	if _, found := m.Get(mes.Hash); found {
		return Conflict
	}

	m.Set(mes.Hash, mes)

	return Stored
}

func StoreC(com data.Command) (r Result) {
	r = Stored
	if _, found := c.Get(com.IP); found {
		r = Updated
	}

	c.Set(com.IP, com)

	return
}

func ContainsC(ip string) bool {
	return c.Has(ip)
}

func DeleteC(ip string) {
	c.Pop(ip)
}

func StoreU(pl data.Player) (status Result) {
	status = Stored
	if _, found := p.Get(pl.IP); found {
		status = Updated
	}

	p.Set(pl.IP, pl)

	return
}

func ContainsP(ip string) bool {
	return p.Has(ip)
}

func DeleteP(ip string) {
	p.Pop(ip)
}

func StoreI(addr string) (status Result) {
	status = Stored

	if _, found := p.Get(addr); found {
		status = Updated
	}

	ip.Set(addr, time.Now())

	return
}

func GetM(k string) (data.Message, bool) {
	return m.Get(k)
}

func ContainsM(mes data.Message) bool {
	return m.Has(mes.Hash)
}

func GetC(k string) (data.Command, bool) {
	return c.Get(k)
}

func GetU(k string) (data.Player, bool) {
	return p.Get(k)
}

func GetI(k string) (time.Time, bool) {
	return ip.Get(k)
}

func GetAllU() map[string]data.Player {
	return p.Items()
}

func GetAllC() map[string]data.Command {
	return c.Items()
}

func GetAllM() map[string]data.Message {
	return m.Items()
}
