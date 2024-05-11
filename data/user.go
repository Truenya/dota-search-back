package data

import "time"

type PossiblePos []int

type Player struct {
	IP          string      `json:"ip,omitempty"`
	Data        string      `json:"data,omitempty"`
	Link        string      `json:"link,omitempty"`
	MMR         int         `json:"mmr,omitempty"`
	PossiblePos PossiblePos `json:"possible_pos,omitempty"`
	Timestamp   time.Time   `json:"timestamp,omitempty"`
}

type Players struct {
	Users []Player `json:"list"`
}

func ToPlayers(players ...Player) Players {
	return Players{Users: players}
}
