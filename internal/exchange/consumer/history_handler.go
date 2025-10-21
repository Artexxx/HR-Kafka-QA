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

func (h *handler) processHistory(sess sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage, messageId uuid.UUID, history HistoryPayload) bool {

	ctx := sess.Context()

	if messageId == uuid.Nil {
		h.toDLQ(ctx, msg, "missing required field message_id")
		return h.commitOnDLQ
	}

	if history.EmployeeID == "" {
		h.toDLQ(ctx, msg, "missing required field employee_id")
		return h.commitOnDLQ
	}

	exists, err := h.events.ExistsMessage(ctx, messageId)
	if err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("events.ExistsMessage: db error exists: %s", err.Error()))
		return h.commitOnDLQ
	}
	if exists {
		h.log.Info().Str("message_id", messageId.String()).Str("employee_id", history.EmployeeID).Msg("duplicate message, skip (idempotency)")
		return true
	}

	if verr := validateHistory(history); verr != "" {
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
		h.toDLQ(ctx, msg, fmt.Sprintf("events.InsertEvent: %s", err.Error()))
		return h.commitOnDLQ
	}

	hDto := dto.EmploymentHistory{
		EmployeeID: history.EmployeeID,
		Company:    &history.Company,
		Position:   nil,
		PeriodFrom: &history.Period.From,
		PeriodTo:   &history.Period.To,
		Stack:      append([]string(nil), history.Stack...),
	}
	if history.Position != "" {
		pos := history.Position
		hDto.Position = &pos
	}

	if err := h.history.Insert(ctx, hDto); err != nil {
		h.toDLQ(ctx, msg, fmt.Sprintf("history.Insert: %s", err.Error()))

		return h.commitOnDLQ
	}

	return true
}
