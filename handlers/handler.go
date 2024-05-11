package handlers

import (
	"time"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/db"
	log "github.com/sirupsen/logrus"
)

var TooEarly = 30 * time.Second

func CheckIP(ip string) (bool, bool) {
	log.Debugf("client with ip %s posting", ip)

	prev, found := cache.GetI(ip)
	if found && time.Since(prev) < TooEarly {
		return found, true
	}

	cache.StoreI(ip)

	return found, false
}

type ActionDid uint8

const (
	JustAdded ActionDid = iota
	Updated
	Deleted
	DeletedAndAdded
	TooManyRequests
	NotFound
)

func AddPlayer(p data.Player) ActionDid { //nolint: dupl
	ip := p.IP

	found, tooManyRequests := CheckIP(ip)
	if tooManyRequests {
		return TooManyRequests
	}

	cache.StoreI(ip)
	log.Debug("Got player to save ", p)

	if !found {
		log.Debugf("ip %s not found", ip)
		cache.StoreU(p)
		db.StoreP(p)

		log.Debugf("User %v stored to db", p)

		return JustAdded
	}

	log.Debugf("ip %s found", ip)

	if cache.ContainsC(ip) {
		// Удалить команду в кэше и базе
		cache.DeleteC(ip)
		db.DeleteC(ip)
		log.Debugf("Deleted command with ip %s", ip)
	}

	if cache.ContainsP(ip) {
		// Заменить игрока в кэше и базе
		db.UpdateP(p)
		cache.StoreU(p)
		log.Debugf("player %v updated", p)

		return Updated
	}

	log.Debugf("player with ip %s not found", ip)
	cache.StoreU(p)
	db.StoreP(p)

	log.Debugf("User %v stored to db", p)

	return DeletedAndAdded
}

func AddCommand(c data.Command) ActionDid { //nolint: dupl
	ip := c.IP

	found, tooManyRequests := CheckIP(ip)
	if tooManyRequests {
		return TooManyRequests
	}

	cache.StoreI(ip)
	log.Debug("Got command to save ", c)

	if !found {
		log.Debugf("ip %s not found", ip)
		cache.StoreC(c)
		db.StoreC(c)

		log.Debugf("User %v stored to db", c)

		return JustAdded
	}

	log.Debugf("ip %s found", ip)

	if cache.ContainsP(ip) {
		// Удалить игрока в кэше и базе
		cache.DeleteP(ip)
		db.DeleteP(ip)
		log.Debugf("Deleted player with ip %s", ip)
	}

	if cache.ContainsC(ip) {
		// Заменить игрока в кэше и базе
		db.UpdateC(c)
		cache.StoreC(c)
		log.Debugf("command %v updated", c)

		return Updated
	}

	log.Debugf("player with ip %s not found", ip)
	cache.StoreC(c)
	db.StoreC(c)

	log.Debugf("Command %v stored to db", c)

	return DeletedAndAdded
}

func AddMessage(m data.Message) ActionDid {
	log.Infoln("Send message to everyone:", m.Data)
	cache.StoreM(m)
	db.StoreM(m)

	return JustAdded
}

func DeleteByKey(key string) ActionDid {
	if cache.ContainsC(key) {
		cache.DeleteC(key)
		db.DeleteC(key)

		return Deleted
	}

	if cache.ContainsP(key) {
		cache.DeleteP(key)
		db.DeleteP(key)

		return Deleted
	}

	return NotFound
}
