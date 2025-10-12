package consumer

import (
	"context"
	"encoding/json"
	"fmt"

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
	for msg := range claim.Messages() {
		switch h.kind {
		case kindPersonal:
			var env Envelope[PersonalPayload]
			if err := json.Unmarshal(msg.Value, &env); err != nil {
				h.toDLQ(sess.Context(), msg, fmt.Sprintf("invalid_json: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(msg, "")
				}
				continue
			}
			if ok := h.processPersonal(sess, msg, env); ok {
				sess.MarkMessage(msg, "")
			}
		case kindPositions:
			var env Envelope[PositionPayload]
			if err := json.Unmarshal(msg.Value, &env); err != nil {
				h.toDLQ(sess.Context(), msg, fmt.Sprintf("invalid_json: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(msg, "")
				}
				continue
			}
			if ok := h.processPosition(sess, msg, env); ok {
				sess.MarkMessage(msg, "")
			}
		case kindHistory:
			var env Envelope[HistoryPayload]
			if err := json.Unmarshal(msg.Value, &env); err != nil {
				h.toDLQ(sess.Context(), msg, fmt.Sprintf("invalid_json: %v", err))
				if h.commitOnDLQ {
					sess.MarkMessage(msg, "")
				}
				continue
			}
			if ok := h.processHistory(sess, msg, env); ok {
				sess.MarkMessage(msg, "")
			}
		default:
			h.log.Error().Str("kind", string(h.kind)).Msg("unknown consumer kind")
			sess.MarkMessage(msg, "")
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
