package consumer

import (
	"fmt"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func NewPersonalRunner(
	bootstrap string,
	topic string,
	groupID string,
	events EventsRepository,
	profiles ProfileRepository,
	log zerolog.Logger,
) *Runner {
	h := &handler{
		kind:        kindPersonal,
		events:      events,
		profiles:    profiles,
		history:     nil,
		log:         log.With().Str("consumer", "personal").Logger(),
		commitOnDLQ: false,
	}

	return newRunner(bootstrap, groupID, topic, h, log)
}

func (h *handler) processPersonal(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, messageId uuid.UUID, personal PersonalPayload) bool {
	ctx := sess.Context()

	if messageId == uuid.Nil {
		h.toDLQ(ctx, msg, "missing required field message_id")
		return h.commitOnDLQ
	}

	if personal.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_id")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, messageId)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: %v", err))
		return h.commitOnDLQ
	}

	if exists {
		h.log.Info().
			Str("message_id", messageId.String()).
			Str("employee_id", personal.EmployeeID).
			Msg("duplicate message, skip (idempotency)")
		return true // коммитим — событие уже обработано ранее
	}

	if verr := validatePersonal(personal); verr != "" {
		h.toDLQ(ctx, msg, verr)
		return h.commitOnDLQ
	}

	if err := h.events.InsertEvent(ctx, dto.KafkaEvent{
		MessageID: messageId,
		Topic:     msg.Topic,
		Partition: int(msg.Partition),
		Offset:    msg.Offset,
		Payload:   append([]byte(nil), msg.Value...),
	}); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert personal: %s", err.Error()))

		return h.commitOnDLQ
	}

	employee := dto.EmployeeProfile{
		EmployeeID: personal.EmployeeID,
		FirstName:  personal.FirstName,
		LastName:   personal.LastName,
		BirthDate:  personal.BirthDate,
		Email:      personal.Contacts.Email,
		Phone:      personal.Contacts.Phone,
	}

	if err := h.profiles.UpsertPersonal(ctx, employee); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("profiles.UpsertPersonal: %v", err))
		return h.commitOnDLQ
	}

	return true
}
