package producer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type HRProducer struct {
	sp             sarama.SyncProducer
	topicPersonal  string
	topicPositions string
	topicHistory   string
	source         string
	log            zerolog.Logger
}

type Config struct {
	TopicPersonal  string
	TopicPositions string
	TopicHistory   string
	Source         string
}

func NewHRProducer(sp sarama.SyncProducer, cfg Config, log zerolog.Logger) *HRProducer {
	return &HRProducer{
		sp:             sp,
		topicPersonal:  cfg.TopicPersonal,
		topicPositions: cfg.TopicPositions,
		topicHistory:   cfg.TopicHistory,
		source:         cfg.Source,
		log:            log.With().Str("component", "HRProducer").Logger(),
	}
}

func (p *HRProducer) Close() error {
	if p == nil || p.sp == nil {
		return nil
	}
	return p.sp.Close()
}

func (p *HRProducer) ProducePersonal(ctx context.Context, messageID uuid.UUID, profile dto.EmployeeProfile) error {
	var payload PersonalPayload

	payload.EmployeeID = profile.EmployeeID
	if profile.FirstName != nil {
		payload.FirstName = *profile.FirstName
	}
	if profile.LastName != nil {
		payload.LastName = *profile.LastName
	}
	if profile.BirthDate != nil {
		payload.BirthDate = *profile.BirthDate
	}
	if profile.Email != nil {
		payload.Contacts.Email = *profile.Email
	}
	if profile.Phone != nil {
		payload.Contacts.Phone = *profile.Phone
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal personal payload: %w", err)
	}

	return p.send(ctx, p.topicPersonal, messageID.String(), body, map[string]string{
		"event-kind":   "personal",
		"source":       p.source,
		"content-type": "application/json",
	})
}

func (p *HRProducer) ProducePosition(ctx context.Context, messageID uuid.UUID, profile dto.EmployeeProfile) error {
	var payload PositionPayload

	payload.EmployeeID = profile.EmployeeID
	if profile.Title != nil {
		payload.Title = *profile.Title
	}
	if profile.Department != nil {
		payload.Department = *profile.Department
	}
	if profile.Grade != nil {
		payload.Grade = *profile.Grade
	}
	if profile.EffectiveFrom != nil {
		payload.EffectiveFrom = *profile.EffectiveFrom
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal position payload: %w", err)
	}

	return p.send(ctx, p.topicPositions, messageID.String(), body, map[string]string{
		"event-kind":   "position",
		"source":       p.source,
		"content-type": "application/json",
	})
}

func (p *HRProducer) ProduceHistory(ctx context.Context, messageID uuid.UUID, history dto.EmploymentHistory) error {
	var body HistoryPayload

	body.EmployeeID = history.EmployeeID
	if history.Company != nil {
		body.Company = *history.Company
	}
	if history.Position != nil {
		body.Position = *history.Position
	}
	if history.PeriodFrom != nil {
		body.Period.From = *history.PeriodFrom
	}
	if history.PeriodTo != nil {
		body.Period.To = *history.PeriodTo
	}
	body.Stack = append(body.Stack, history.Stack...)

	message, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}

	return p.send(ctx, p.topicHistory, messageID.String(), message, map[string]string{
		"event-kind": "history",
		"source":     p.source,
	})
}

func (p *HRProducer) send(_ context.Context, topic, key string, value []byte, headers map[string]string) error {
	if p == nil || p.sp == nil {
		return errors.New("sync producer is not initialized")
	}

	var hs []sarama.RecordHeader
	for k, v := range headers {
		hs = append(hs, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
	}

	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: hs,
	}

	part, off, err := p.sp.SendMessage(msg)
	if err != nil {
		p.log.Error().
			Err(err).
			Str("topic", topic).
			Str("key", key).
			Int("headers_count", len(headers)).
			Int("bytes", len(value)).
			Msg("failed to send kafka message")
		return fmt.Errorf("send kafka message: %w", err)
	}

	p.log.Info().
		Str("topic", topic).
		Str("key", key).
		Int32("partition", part).
		Int64("offset", off).
		Int("bytes", len(value)).
		Msg("kafka message sent")
	return nil
}
