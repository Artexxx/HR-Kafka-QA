package consumer

import (
	"fmt"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func NewHistoryRunner(
	bootstrap string,
	topic string,
	groupID string,
	events EventsRepository,
	history HistoryRepository,
	log zerolog.Logger,
) *Runner {
	h := &handler{
		kind:        kindHistory,
		events:      events,
		profiles:    nil,
		history:     history,
		log:         log.With().Str("consumer", "history").Logger(),
		commitOnDLQ: false,
	}

	return newRunner(bootstrap, groupID, topic, h, log)
}

func (h *handler) processHistory(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, env Envelope[HistoryPayload]) bool {
	ctx := sess.Context()

	msgID, err := uuid.Parse(env.MessageID)
	if err != nil || env.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing_required_field")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, msgID)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: db error exists: %s", err.Error()))
		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().Str("message_id", env.MessageID).Str("employee_id", env.EmployeeID).Msg("duplicate message, skip (idempotency)")
		return true
	}

	if verr := validateHistory(env.Payload); verr != "" {
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
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: db error insert history: %s", err.Error()))

		return h.commitOnDLQ
	}

	hDto := dto.EmploymentHistory{
		EmployeeID: env.EmployeeID,
		Company:    env.Payload.Company,
		Position:   nil,
		PeriodFrom: env.Payload.Period.From,
		PeriodTo:   env.Payload.Period.To,
		Stack:      append([]string(nil), env.Payload.Stack...),
	}
	if env.Payload.Position != "" {
		pos := env.Payload.Position
		hDto.Position = &pos
	}

	if err := h.history.Insert(ctx, hDto); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("history.Insert: db error insert history: %s", err.Error()))

		return h.commitOnDLQ
	}

	return true
}
