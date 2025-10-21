package dto

import (
	"encoding/json"
	"github.com/google/uuid"
)

// KafkaEvent — сырое событие
type KafkaEvent struct {
	ID         int64           `json:"id"`
	MessageID  uuid.UUID       `json:"message_id"`
	Topic      string          `json:"topic"`
	Partition  int             `json:"partition"`
	Offset     int64           `json:"offset"`
	Payload    json.RawMessage `json:"payload"`
	ReceivedAt string          `json:"received_at"`
}

// KafkaDLQ — сообщение в DLQ
type KafkaDLQ struct {
	ID         int64           `json:"id"`
	Topic      string          `json:"topic"`
	Key        string          `json:"key"`
	Payload    json.RawMessage `json:"payload"`
	Error      string          `json:"error"`
	ReceivedAt string          `json:"received_at"`
}
