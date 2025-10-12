package producer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"

	"github.com/IBM/sarama"

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
	Source         string // например: "qa-kafka-lab"
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

func (p *HRProducer) ProducePersonal(ctx context.Context, messageID string, profile dto.EmployeeProfile) error {
	if messageID == "" || profile.EmployeeID == "" {
		return errors.New("missing message_id or employee_id")
	}

	var pl PersonalPayload
	pl.MessageID = messageID
	pl.EmployeeID = profile.EmployeeID
	if profile.FirstName != nil {
		pl.FirstName = *profile.FirstName
	}
	if profile.LastName != nil {
		pl.LastName = *profile.LastName
	}
	if profile.BirthDate != nil {
		pl.BirthDate = *profile.BirthDate
	}
	if profile.Email != nil {
		pl.Contacts.Email = *profile.Email
	}
	if profile.Phone != nil {
		pl.Contacts.Phone = *profile.Phone
	}

	env := Envelope[PersonalPayload]{
		Kind:       "personal",
		MessageID:  pl.MessageID,
		EmployeeID: pl.EmployeeID,
		Payload:    pl,
		Timestamp:  time.Now().UTC(),
		Source:     p.source,
	}

	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal personal envelope: %w", err)
	}

	return p.send(ctx, p.topicPersonal, pl.EmployeeID, body, map[string]string{
		"event-kind":   "personal",
		"message-id":   pl.MessageID,
		"employee-id":  pl.EmployeeID,
		"source":       p.source,
		"content-type": "application/json",
	})
}

func (p *HRProducer) ProducePosition(ctx context.Context, messageID string, profile dto.EmployeeProfile) error {
	if messageID == "" || profile.EmployeeID == "" {
		return errors.New("missing message_id or employee_id")
	}

	var pl PositionPayload
	pl.MessageID = messageID
	pl.EmployeeID = profile.EmployeeID
	if profile.Title != nil {
		pl.Title = *profile.Title
	}
	if profile.Department != nil {
		pl.Department = *profile.Department
	}
	if profile.Grade != nil {
		pl.Grade = *profile.Grade
	}
	if profile.EffectiveFrom != nil {
		pl.EffectiveFrom = *profile.EffectiveFrom
	}

	env := Envelope[PositionPayload]{
		Kind:       "position",
		MessageID:  pl.MessageID,
		EmployeeID: pl.EmployeeID,
		Payload:    pl,
		Timestamp:  time.Now().UTC(),
		Source:     p.source,
	}

	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal position envelope: %w", err)
	}

	return p.send(ctx, p.topicPositions, pl.EmployeeID, body, map[string]string{
		"event-kind":   "position",
		"message-id":   pl.MessageID,
		"employee-id":  pl.EmployeeID,
		"source":       p.source,
		"content-type": "application/json",
	})
}

func (p *HRProducer) ProduceHistory(ctx context.Context, messageID string, h dto.EmploymentHistory) error {
	if messageID == "" || h.EmployeeID == "" {
		return errors.New("missing message_id or employee_id")
	}
	if h.PeriodTo < h.PeriodFrom {
		return errors.New("invalid_period: period.to < period.from")
	}

	var pl HistoryPayload
	pl.MessageID = messageID
	pl.EmployeeID = h.EmployeeID
	pl.Company = h.Company
	if h.Position != nil {
		pl.Position = *h.Position
	}
	pl.Period.From = h.PeriodFrom
	pl.Period.To = h.PeriodTo
	pl.Stack = append(pl.Stack, h.Stack...)

	env := Envelope[HistoryPayload]{
		Kind:       "history",
		MessageID:  pl.MessageID,
		EmployeeID: pl.EmployeeID,
		Payload:    pl,
		Timestamp:  time.Now().UTC(),
		Source:     p.source,
	}

	body, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal history envelope: %w", err)
	}

	return p.send(ctx, p.topicHistory, pl.EmployeeID, body, map[string]string{
		"event-kind":   "history",
		"message-id":   pl.MessageID,
		"employee-id":  pl.EmployeeID,
		"source":       p.source,
		"content-type": "application/json",
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
