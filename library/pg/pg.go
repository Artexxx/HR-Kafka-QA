package pg

import (
	"context"
	"sync"

	"HR-Kafka-QA/library/yamlenv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type PostgresConfig struct {
	Conn *yamlenv.Env[string] `yaml:"conn"`
}
type postgres struct {
	db *pgxpool.Pool
}

var (
	pgInstance *postgres
	pgOnce     sync.Once
)

func NewPG(
	ctx context.Context,
	conn string,
	log zerolog.Logger,
) (*postgres, error) {
	var err error
	pgOnce.Do(func() {
		cfg, parseErr := pgxpool.ParseConfig(conn)
		if parseErr != nil {
			err = errors.Wrap(parseErr, "unable to parse connection config")
			log.Error().Err(err).Send()
			return
		}

		cfg.ConnConfig.Tracer = &myQueryTracer{
			log: log,
		}

		db, poolErr := pgxpool.NewWithConfig(ctx, cfg)
		if poolErr != nil {
			err = errors.Wrap(poolErr, "unable to create connection pool")
			log.Error().Err(err).Send()
			return
		}

		err = checkDBConnection(ctx, db, log)
		if err != nil {
			err = errors.Wrap(err, "database connection check failed")
			log.Error().Err(err).Send()
			return
		}

		pgInstance = &postgres{db}
	})

	if err != nil {
		return nil, err
	}

	return pgInstance, nil
}

func checkDBConnection(
	ctx context.Context,
	db *pgxpool.Pool,
	log zerolog.Logger,
) error {
	_, err := db.Acquire(ctx)
	if err != nil {
		log.Error().Err(err).Msg("unable to acquire a connection from the pool")
		return err
	}
	return nil
}

func (pg *postgres) Close() {
	if pg == nil || pg.db == nil {
		return
	}
	pg.db.Close()
}

func (pg *postgres) Pool() *pgxpool.Pool {
	return pg.db
}

type myQueryTracer struct {
	log zerolog.Logger
}

func (tracer *myQueryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	tracer.log.Info().Str("sql", data.SQL).Interface("args", data.Args).Send()
	return ctx
}

func (tracer *myQueryTracer) TraceQueryEnd(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryEndData,
) {
	if data.Err != nil {
		tracer.log.Error().Err(data.Err).Interface("args", data.CommandTag).Send()
	}
}
