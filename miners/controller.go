package miners

import (
	"strings"

	"github.com/Truenya/dota-search-back/cache"
	"github.com/Truenya/dota-search-back/data"
	"github.com/Truenya/dota-search-back/handlers"
	"github.com/Truenya/dota-search-back/net"
)

type Controller struct {
	miners   []miner
	stopChan chan struct{}
}

func (c *Controller) Mine() {
	c.stopChan = make(chan struct{}, 1)
	messages := make(chan data.Messages, 100)

	c.miners = append(c.miners, newVKMiner())

	for _, s := range c.miners {
		go s.Spoofing(&messages)
	}

	go c.Store(&messages)
}

var invalids = []string{"прод", "купл", "сервер", "коммьюн", "сайт"}

func Valid(s string) bool {
	s = strings.ToLower(s)
	for _, bad := range invalids {
		if strings.Contains(s, bad) {
			return false
		}
	}

	return true
}

func (c *Controller) Store(messages *chan data.Messages) {
	for {
		select {
		case <-c.stopChan:
			return
		case ms := <-*messages:
			for _, m := range ms.Messages {
				if !Valid(m.Data) {
					continue
				}

				if cache.ContainsM(m) {
					continue
				}

				handlers.AddMessage(m)
				net.SendToWS(data.FromMessage(m))
			}
		}
	}
}
func (c *Controller) Stop() {
	c.stopChan <- struct{}{}
}
