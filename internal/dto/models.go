package dto

import (
	"encoding/json"

	"github.com/google/uuid"
)

// EmployeeProfile — данные профиля сотрудника
type EmployeeProfile struct {
	EmployeeID    string  `json:"employee_id"`
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
	BirthDate     *string `json:"birth_date,omitempty"`
	Email         *string `json:"email,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Title         *string `json:"title,omitempty"`
	Department    *string `json:"department,omitempty"`
	Grade         *string `json:"grade,omitempty"`
	EffectiveFrom *string `json:"effective_from,omitempty"`
	UpdatedAt     string  `json:"updated_at"`
}

// EmploymentHistory — запись истории работы
type EmploymentHistory struct {
	ID         int64    `json:"id"`
	EmployeeID string   `json:"employee_id"`
	Company    string   `json:"company"`
	Position   *string  `json:"position,omitempty"`
	PeriodFrom string   `json:"period_from"`
	PeriodTo   string   `json:"period_to"`
	Stack      []string `json:"stack"`
	CreatedAt  string   `json:"created_at"`
}

// KafkaEvent — сырое событие
type KafkaEvent struct {
	ID         int64           `json:"id"`
	MessageID  uuid.UUID       `json:"message_id"`
	Topic      string          `json:"topic"`
	Key        string          `json:"key"`
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
