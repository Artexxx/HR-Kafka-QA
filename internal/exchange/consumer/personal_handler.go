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

func (h *handler) processPersonal(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, event Envelope[PersonalPayload]) bool {
	ctx := sess.Context()

	msgID, err := uuid.Parse(event.MessageID)
	if err != nil || event.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_id")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, msgID)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: db error exists: %v", err))
		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().
			Str("message_id", event.MessageID).
			Str("employee_id", event.EmployeeID).
			Msg("duplicate message, skip (idempotency)")
		return true // коммитим — событие уже обработано ранее
	}

	if verr := validatePersonal(event.Payload); verr != "" {
		h.toDLQ(ctx, msg, verr)

		return h.commitOnDLQ
	}

	if err := h.events.InsertEvent(ctx, dto.KafkaEvent{
		MessageID: msgID,
		Topic:     msg.Topic,
		Key:       string(msg.Key),
		Partition: int(msg.Partition),
		Offset:    msg.Offset,
		Payload:   append([]byte(nil), msg.Value...),
	}); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert event: %s", err.Error()))

		return h.commitOnDLQ
	}

	// Бизнес-апдейт: upsert personal
	first := event.Payload.FirstName
	last := event.Payload.LastName
	bdate := event.Payload.BirthDate
	email := event.Payload.Contacts.Email
	phone := event.Payload.Contacts.Phone

	// приводим к DTO (pointer-поля)
	p := dto.EmployeeProfile{
		EmployeeID: event.EmployeeID,
		UpdatedAt:  "",
	}
	if first != "" {
		p.FirstName = &first
	}
	if last != "" {
		p.LastName = &last
	}
	if bdate != "" {
		p.BirthDate = &bdate
	}
	if email != "" {
		p.Email = &email
	}
	if phone != "" {
		p.Phone = &phone
	}

	if err := h.profiles.UpsertPersonal(ctx, p); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("profiles.UpsertPersonal db error upsert personal: %v", err))

		return h.commitOnDLQ
	}

	return true
}
