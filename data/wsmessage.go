package data

type WsMessage struct {
	Player
	// data.Command is same as Player
	// data.Message not aggregated to not ignore Link field by marshaller
	ID     int    `json:"id,omitempty"`
	Hash   string `json:"hash,omitempty"`
	Type   string `json:"msg_type,omitempty"` // Only ws thing {command|player|message}
	Action string `json:"action,omitempty"`   // Only ws thing {add|delete}
}

func FromPlayer(p Player) WsMessage {
	a := WsMessage{Player: p}
	a.Type = "player"

	return a
}

func FromCommand(c Command) WsMessage {
	a := WsMessage{Player: c.Player}
	a.Type = "command"

	return a
}

func FromMessage(m Message) WsMessage {
	a := WsMessage{}
	a.ID = m.ID
	a.Data = m.Data
	a.Link = m.Link
	a.Timestamp = m.Timestamp
	a.Hash = m.Hash
	a.Type = "message"

	return a
}

func (wm WsMessage) ToPlayer() (p Player) {
	return wm.Player
}

func (wm WsMessage) ToCommand() (c Command) {
	return Command{Player: wm.Player}
}

func (wm WsMessage) ToMessage() (m Message) {
	m.ID = wm.ID
	m.Data = wm.Data
	m.Link = wm.Link
	m.Timestamp = wm.Timestamp
	m.Hash = wm.Hash

	return
}
