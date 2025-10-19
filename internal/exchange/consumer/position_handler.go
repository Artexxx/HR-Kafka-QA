package consumer

import (
	"fmt"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func NewPositionsRunner(
	bootstrap string,
	topic string,
	groupID string,
	events EventsRepository,
	profiles ProfileRepository,
	log zerolog.Logger,
) *Runner {
	h := &handler{
		kind:        kindPositions,
		events:      events,
		profiles:    profiles,
		history:     nil,
		log:         log.With().Str("consumer", "positions").Logger(),
		commitOnDLQ: false,
	}

	return newRunner(bootstrap, groupID, topic, h, log)
}
func (h *handler) processPosition(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, event Envelope[PositionPayload]) bool {
	ctx := sess.Context()

	if event.MessageID != uuid.Nil {
		h.toDLQ(ctx, msg, "missing required field message_id")
		return h.commitOnDLQ
	}

	if event.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_id")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, event.MessageID)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: %v", err))

		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().Str("message_id", event.MessageID.String()).Str("employee_id", event.EmployeeID).Msg("duplicate message, skip (idempotency)")
		return true
	}

	if verr := validatePosition(event.Payload); verr != "" {
		h.toDLQ(ctx, msg, verr)
		return h.commitOnDLQ
	}

	if err := h.events.InsertEvent(ctx, dto.KafkaEvent{
		MessageID: event.MessageID,
		Topic:     msg.Topic,
		Key:       string(msg.Key),
		Partition: int(msg.Partition),
		Offset:    msg.Offset,
		Payload:   append([]byte(nil), msg.Value...),
	}); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert event: %v", err))

		return h.commitOnDLQ
	}

	var (
		title = event.Payload.Title
		dept  = event.Payload.Department
		grade = event.Payload.Grade
		eff   = event.Payload.EffectiveFrom
	)

	payload := dto.EmployeeProfile{
		EmployeeID: event.EmployeeID,
	}

	if title != "" {
		payload.Title = &title
	}
	if dept != "" {
		payload.Department = &dept
	}
	if grade != "" {
		payload.Grade = &grade
	}
	if eff != "" {
		payload.EffectiveFrom = &eff
	}

	if err := h.profiles.UpsertPosition(ctx, payload); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("profiles.UpsertPosition: db error upsert position: %v", err))

		return h.commitOnDLQ
	}

	return true
}
