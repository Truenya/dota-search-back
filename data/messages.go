package data

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Message struct {
	ID        int `json:"id"`
	authorID  int
	Data      string    `json:"data"`
	Link      string    `json:"link"`
	Timestamp time.Time `json:"timestamp"`
	Hash      string    `json:"hash"`
}

func (m *Message) SetAuthorID(a int) {
	m.authorID = a
}

func (m *Message) CalculateHash() {
	hash := sha256.Sum256([]byte(m.Data + strconv.Itoa(m.authorID)))
	m.Hash = hex.EncodeToString(hash[:])
}

type Messages struct {
	Messages []Message `json:"list"`
}

func ToMessages(message ...Message) Messages {
	return Messages{message}
}
