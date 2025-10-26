package events

import (
	"context"
	"errors"
	"fmt"

	"github.com/Artexxx/HR-Kafka-QA/internal/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PgxPoolIface interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Repository struct {
	pool PgxPoolIface
}

func NewRepository(pool PgxPoolIface) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ExistsMessage(ctx context.Context, messageID uuid.UUID) (bool, error) {
	query := `
SELECT 1
FROM kafka_events
WHERE message_id = $1::uuid
LIMIT 1;
`
	row := r.pool.QueryRow(ctx, query, messageID)

	var x int
	err := row.Scan(&x)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *Repository) InsertEvent(ctx context.Context, event dto.KafkaEvent) error {
	query := `
INSERT INTO kafka_events
	(topic, message_id, partition, "offset", payload, received_at)
VALUES
	($1, $2::uuid, $3, $4, $5::jsonb, NOW());
`
	_, err := r.pool.Exec(ctx, query, event.Topic, event.MessageID, event.Partition, event.Offset, string(event.Payload))
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) InsertDLQ(ctx context.Context, dlq dto.KafkaDLQ) error {
	query := `
INSERT INTO kafka_dlq
	(topic, msg_key, payload, error, received_at)
VALUES
	($1, $2, $3::jsonb, $4, NOW());
`
	_, err := r.pool.Exec(ctx, query, dlq.Topic, dlq.Key, string(dlq.Payload), dlq.Error)
	if err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}

func (r *Repository) ListEvents(ctx context.Context) ([]dto.KafkaEvent, error) {
	query := `
SELECT id, topic, message_id, partition, "offset", payload, to_char(received_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
FROM kafka_events
ORDER BY id DESC
`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	var out []dto.KafkaEvent
	for rows.Next() {
		var (
			kafkaEvent dto.KafkaEvent
			payload    []byte
		)

		err = rows.Scan(&kafkaEvent.ID, &kafkaEvent.Topic, &kafkaEvent.MessageID, &kafkaEvent.Partition, &kafkaEvent.Offset, &payload, &kafkaEvent.ReceivedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		kafkaEvent.Payload = payload
		out = append(out, kafkaEvent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return out, nil
}

func (r *Repository) ListDLQ(ctx context.Context) ([]dto.KafkaDLQ, error) {
	query := `
select id, topic, msg_key, payload, error, to_char(received_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
from kafka_dlq
order by id desc
`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("pool.Query: %w", err)
	}
	defer rows.Close()

	var out []dto.KafkaDLQ
	for rows.Next() {
		var (
			kafkaDLQ dto.KafkaDLQ
			payload  []byte
		)

		err = rows.Scan(&kafkaDLQ.ID, &kafkaDLQ.Topic, &kafkaDLQ.Key, &payload, &kafkaDLQ.Error, &kafkaDLQ.ReceivedAt)
		if err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		kafkaDLQ.Payload = payload
		out = append(out, kafkaDLQ)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return out, nil
}

func (r *Repository) ResetAll(ctx context.Context) error {
	query := `
TRUNCATE kafka_events RESTART IDENTITY CASCADE;
TRUNCATE kafka_dlq RESTART IDENTITY CASCADE;
TRUNCATE employment_history RESTART IDENTITY CASCADE;
TRUNCATE employee_profile RESTART IDENTITY CASCADE;
`
	if _, err := r.pool.Exec(ctx, query); err != nil {
		return fmt.Errorf("pool.Exec: %w", err)
	}

	return nil
}
