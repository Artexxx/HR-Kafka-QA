package consumer

import (
	"fmt"

	"github.com/Artexxx/github.com/Artexxx/HR-Kafka-QA/internal/dto"

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
func (h *handler) processPosition(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, env Envelope[PositionPayload]) bool {
	ctx := sess.Context()

	msgID, err := uuid.Parse(env.MessageID)
	if err != nil || env.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_iD")

		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, msgID)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: db error exists: %v", err))

		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().Str("message_id", env.MessageID).Str("employee_id", env.EmployeeID).Msg("duplicate message, skip (idempotency)")
		return true
	}

	if verr := validatePosition(env.Payload); verr != "" {
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
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert event: %v", err))

		return h.commitOnDLQ
	}

	title := env.Payload.Title
	dept := env.Payload.Department
	grade := env.Payload.Grade
	eff := env.Payload.EffectiveFrom

	p := dto.EmployeeProfile{
		EmployeeID: env.EmployeeID,
		UpdatedAt:  "",
	}
	if title != "" {
		p.Title = &title
	}
	if dept != "" {
		p.Department = &dept
	}
	if grade != "" {
		p.Grade = &grade
	}
	if eff != "" {
		p.EffectiveFrom = &eff
	}

	if err := h.profiles.UpsertPosition(ctx, p); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("profiles.UpsertPosition: db error upsert position: %v", err))

		return h.commitOnDLQ
	}

	return true
}
