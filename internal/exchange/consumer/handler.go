package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/IBM/sarama"
	"github.com/rs/zerolog"
)

type kind string

const (
	kindPersonal  kind = "personal"
	kindPositions kind = "position"
	kindHistory   kind = "history"
)

type handler struct {
	kind        kind
	events      EventsRepository
	profiles    ProfileRepository
	history     HistoryRepository
	log         zerolog.Logger
	commitOnDLQ bool
}

func (h *handler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		messageID, err := messageIDFromKey(message)
		if err != nil {
			h.toDLQ(sess.Context(), message, fmt.Sprintf("error in message_id parse: %v", err))
			sess.MarkMessage(message, "")
		}

		switch h.kind {
		case kindPersonal:
			var event PersonalPayload
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.toDLQ(sess.Context(), message, fmt.Sprintf("json.Unmarshal: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(message, "")
				}
				continue
			}
			if ok := h.processPersonal(sess, message, messageID, event); ok {
				sess.MarkMessage(message, "")
			}
		case kindPositions:
			var event PositionPayload
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.toDLQ(sess.Context(), message, fmt.Sprintf("json.Unmarshal: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(message, "")
				}
				continue
			}
			if ok := h.processPosition(sess, message, messageID, event); ok {
				sess.MarkMessage(message, "")
			}
		case kindHistory:
			var event HistoryPayload
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.toDLQ(sess.Context(), message, fmt.Sprintf("json.Unmarshal: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(message, "")
				}
				continue
			}
			if ok := h.processHistory(sess, message, messageID, event); ok {
				sess.MarkMessage(message, "")
			}
		default:
			h.log.Error().Str("kind", string(h.kind)).Msg("unknown consumer kind")
			sess.MarkMessage(message, "")
		}
	}
	return nil
}

func (h *handler) toDLQ(ctx context.Context, msg *sarama.ConsumerMessage, reason string) {
	_ = h.events.InsertDLQ(ctx, dto.KafkaDLQ{
		Topic:   msg.Topic,
		Key:     string(msg.Key),
		Payload: append([]byte(nil), msg.Value...),
		Error:   reason,
	})

	h.log.Warn().
		Str("topic", msg.Topic).
		Int32("partition", msg.Partition).
		Int64("offset", msg.Offset).
		Str("reason", reason).
		Msg("message sent to DLQ")
}

func messageIDFromKey(msg *sarama.ConsumerMessage) (uuid.UUID, error) {
	if len(msg.Key) == 0 {
		return uuid.Nil, fmt.Errorf("missing required field message_id")
	}

	id, err := uuid.Parse(string(msg.Key))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid message_id in key: %w", err)
	}

	return id, nil
}
