package miners

import "github.com/Truenya/dota-search-back/data"

type miner interface {
	Spoof()
	Spoofing(messages *chan data.Messages)
	Stop()
}
