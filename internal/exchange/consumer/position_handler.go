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
func (h *handler) processPosition(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, messageId uuid.UUID, position PositionPayload) bool {
	ctx := sess.Context()

	if messageId == uuid.Nil {
		h.toDLQ(ctx, msg, "missing required field message_id")
		return h.commitOnDLQ
	}

	if position.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_id")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, messageId)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: %v", err))

		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().Str("message_id", messageId.String()).Str("employee_id", position.EmployeeID).Msg("duplicate message, skip (idempotency)")
		return true
	}

	if verr := validatePosition(position); verr != "" {
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
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert position: %v", err))

		return h.commitOnDLQ
	}

	var (
		title = position.Title
		dept  = position.Department
		grade = position.Grade
		eff   = position.EffectiveFrom
	)

	payload := dto.EmployeeProfile{
		EmployeeID: position.EmployeeID,
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
