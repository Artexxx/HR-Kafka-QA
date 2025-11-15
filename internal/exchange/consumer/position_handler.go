package consumer

import (
	"errors"
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
		commitOnDLQ: true,
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

	if _, err := h.profiles.GetProfile(ctx, position.EmployeeID); err != nil {
		if errors.Is(err, dto.ErrNotFound) {
			h.toDLQ(ctx, msg, fmt.Sprintf("employee_id=%s not found: create employee profile first", position.EmployeeID))
		}

		if !errors.Is(err, dto.ErrNotFound) {
			h.toDLQ(ctx, msg, fmt.Sprintf("profiles.GetProfile: db error get profile: %v", err))
		}

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

	payload := dto.EmployeeProfile{
		EmployeeID:    position.EmployeeID,
		Title:         &position.Title,
		Department:    &position.Department,
		Grade:         &position.Grade,
		EffectiveFrom: &position.EffectiveFrom,
	}

	if err := h.profiles.UpsertPosition(ctx, payload); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("profiles.UpsertPosition: db error upsert position: %v", err))

		return h.commitOnDLQ
	}

	return true
}
