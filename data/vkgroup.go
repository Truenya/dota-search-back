package data

type GroupData struct {
	ID           int `json:"id"`
	TopicID      int `json:"topic_id"`
	MaxMessageID int `json:"last_message_id"`
}
