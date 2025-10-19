package consumer

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/IBM/sarama"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type EventsRepository interface {
	ExistsMessage(ctx context.Context, messageID uuid.UUID) (bool, error)
	InsertEvent(ctx context.Context, ev dto.KafkaEvent) error
	InsertDLQ(ctx context.Context, dlq dto.KafkaDLQ) error
}

type ProfileRepository interface {
	UpsertPersonal(ctx context.Context, p dto.EmployeeProfile) error
	UpsertPosition(ctx context.Context, p dto.EmployeeProfile) error
}

type HistoryRepository interface {
	Insert(ctx context.Context, h dto.EmploymentHistory) error
}

type Runner struct {
	brokers   []string
	groupID   string
	topic     string
	handler   *handler
	log       zerolog.Logger
	createCfg func() *sarama.Config
}

func newRunner(bootstrap, groupID, topic string, h *handler, log zerolog.Logger) *Runner {
	createCfg := func() *sarama.Config {
		cfg := sarama.NewConfig()
		cfg.Version = sarama.V3_3_2_0
		cfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
		cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
		cfg.Consumer.Return.Errors = true
		// Автокоммит управляется вызовами session.MarkMessage
		return cfg
	}
	return &Runner{
		brokers:   []string{bootstrap},
		groupID:   groupID,
		topic:     topic,
		handler:   h,
		log:       log.With().Str("topic", topic).Str("group", groupID).Logger(),
		createCfg: createCfg,
	}
}

func (r *Runner) Start(ctx context.Context) error {
	cfg := r.createCfg()

	consumerGroup, err := sarama.NewConsumerGroup(r.brokers, r.groupID, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = consumerGroup.Close() }()

	go func() {
		for err := range consumerGroup.Errors() {
			if err == nil || errors.Is(err, context.Canceled) || (strings.Contains(err.Error(), "context canceled")) {
				continue
			}

			r.log.Error().Err(err).Msg("consumer group error")
		}
	}()

	r.log.Info().Msg("consumer started")
	defer r.log.Info().Msg("consumer stopped")

	for {
		if ctx.Err() != nil {
			return nil
		}

		err := consumerGroup.Consume(ctx, []string{r.topic}, r.handler)

		if errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return nil
		}

		if err != nil {
			r.log.Error().Err(err).Msg("consume error")
			time.Sleep(500 * time.Millisecond)
		}
	}
}
