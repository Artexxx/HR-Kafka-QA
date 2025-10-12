package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HR-Kafka-QA/internal/api"
	"HR-Kafka-QA/internal/config"
	"HR-Kafka-QA/internal/exchange/consumer"
	"HR-Kafka-QA/internal/exchange/producer"
	"HR-Kafka-QA/internal/repository/events"
	"HR-Kafka-QA/internal/repository/history"
	"HR-Kafka-QA/internal/repository/profile"
	"HR-Kafka-QA/library/pg"
	"HR-Kafka-QA/library/yamlreader"

	"github.com/IBM/sarama"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

func main() {
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	cfg := MustNewConfig(parseFlags())

	log.Info().Msgf("pg=%+v", cfg.Postgres.Conn.Value)
	log.Info().Msgf("kafka=%+v", cfg.Kafka.Bootstrap.Value)

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339

	pgClient, err := pg.NewPG(rootCtx, cfg.Postgres.Conn.Value, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres init failed")
	}
	defer pgClient.Close()

	eventsRepo := events.NewRepository(pgClient.Pool())
	profileRepo := profile.NewRepository(pgClient.Pool())
	historyRepo := history.NewRepository(pgClient.Pool())

	hrProducer, err := initHRProducer(cfg.Kafka)
	if err != nil {
		log.Fatal().Err(err).Msg("kafka producer init failed")
	}
	defer func() { _ = hrProducer.Close() }()

	apiService := api.NewService(api.ServiceDeps{
		Port:        cfg.UserAPI.Port.Value,
		Producer:    hrProducer,
		EventsRepo:  eventsRepo,
		ProfileRepo: profileRepo,
		HistoryRepo: historyRepo,
	})

	consumerPersonal := consumer.NewPersonalRunner(
		cfg.Kafka.Bootstrap.Value,
		cfg.Kafka.Topics.Personal.Value,
		"consumer_personal",
		eventsRepo,
		profileRepo,
		log.Logger,
	)
	consumerPositions := consumer.NewPositionsRunner(
		cfg.Kafka.Bootstrap.Value,
		cfg.Kafka.Topics.Positions.Value,
		"consumer_positions",
		eventsRepo,
		profileRepo,
		log.Logger,
	)
	consumerHistory := consumer.NewHistoryRunner(
		cfg.Kafka.Bootstrap.Value,
		cfg.Kafka.Topics.History.Value,
		"consumer_history",
		eventsRepo,
		historyRepo,
		log.Logger,
	)

	group, gctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		log.Info().Msg("запуск HTTP API")
		if err := apiService.Start(gctx); err != nil {

			log.Error().Err(err).Msg("HTTP API завершился с ошибкой")

			return err
		}

		log.Info().Msg("HTTP API остановлен")

		return nil
	})

	// Consumer personal
	group.Go(func() error {
		log.Info().Msg("запуск consumer_personal")

		if err := consumerPersonal.Start(gctx); err != nil {
			log.Error().Err(err).Msg("consumer_personal завершился с ошибкой")

			return err
		}

		log.Info().Msg("consumer_personal остановлен")

		return nil
	})

	// Consumer positions
	group.Go(func() error {
		log.Info().Msg("запуск consumer_positions")
		if err := consumerPositions.Start(gctx); err != nil {
			log.Error().Err(err).Msg("consumer_positions завершился с ошибкой")

			return err
		}

		log.Info().Msg("consumer_positions остановлен")

		return nil
	})

	// Consumer history
	group.Go(func() error {
		log.Info().Msg("запуск consumer_history")
		if err := consumerHistory.Start(gctx); err != nil {
			log.Error().Err(err).Msg("consumer_history завершился с ошибкой")

			return err
		}

		log.Info().Msg("consumer_history остановлен")

		return nil
	})

	// упрощённая остановка (без таймаута)
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = group.Wait()
	}()

	select {
	case <-rootCtx.Done():
		log.Info().Msg("signal received, graceful shutdown...")
		<-done
		log.Info().Msg("all services stopped")
	case <-done:
		log.Info().Msg("all services stopped")
	}
}

func initHRProducer(kafkaConfig config.KafkaConfig) (*producer.HRProducer, error) {
	sCfg := sarama.NewConfig()
	sCfg.Version = sarama.V3_3_2_0
	sCfg.Producer.Return.Successes = true
	sCfg.Producer.RequiredAcks = sarama.WaitForAll
	sCfg.Producer.Idempotent = true
	sCfg.Net.MaxOpenRequests = 1
	sCfg.Producer.Retry.Max = 5
	sCfg.Producer.Retry.Backoff = 200 * time.Millisecond

	sp, err := sarama.NewSyncProducer([]string{kafkaConfig.Bootstrap.Value}, sCfg)
	if err != nil {
		return nil, err
	}

	hrProd := producer.NewHRProducer(
		sp,
		producer.Config{
			TopicPersonal:  kafkaConfig.Topics.Personal.Value,
			TopicPositions: kafkaConfig.Topics.Positions.Value,
			TopicHistory:   kafkaConfig.Topics.History.Value,
			Source:         "qa-kafka-api",
		},
		log.Logger,
	)

	return hrProd, nil
}

func waitWithTimeout(done <-chan struct{}, timeout time.Duration) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return
	case <-timer.C:
		log.Warn().Dur("timeout", timeout).Msg("graceful shutdown")
	}
}

func MustNewConfig(path string) *config.Config {
	cfg, err := yamlreader.NewConfig[config.Config](path)

	if err != nil {
		log.Fatal().Str("path", path).Err(err).Msg("ошибка чтения конфигурации приложения")
		return nil
	}

	return cfg
}

func parseFlags() string {
	var configPath string

	flag.StringVar(&configPath, "config", "", "path to config file")
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	godotenv.Load(".env")

	if configPath == "" {
		configPath = "config/application-local.yaml"
	}
	return configPath
}
