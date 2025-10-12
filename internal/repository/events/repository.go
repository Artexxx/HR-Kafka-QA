package events

import (
	"context"
	"errors"

	"github.com/Artexxx/github.com/Artexxx/HR-Kafka-QA/internal/dto"

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
	q := `
SELECT 1
FROM kafka_events
WHERE message_id = $1::uuid
LIMIT 1;
`
	row := r.pool.QueryRow(ctx, q, messageID)

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

func (r *Repository) InsertEvent(ctx context.Context, ev dto.KafkaEvent) error {
	q := `
INSERT INTO kafka_events
	(message_id, topic, msg_key, partition, "offset", payload, received_at)
VALUES
	($1::uuid, $2, $3, $4, $5, $6::jsonb, NOW());
`
	_, err := r.pool.Exec(ctx, q,
		ev.MessageID, ev.Topic, ev.Key, ev.Partition, ev.Offset, string(ev.Payload),
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) InsertDLQ(ctx context.Context, dlq dto.KafkaDLQ) error {
	q := `
INSERT INTO kafka_dlq
	(topic, msg_key, payload, error, received_at)
VALUES
	($1, $2, $3::jsonb, $4, NOW());
`
	_, err := r.pool.Exec(ctx, q, dlq.Topic, dlq.Key, string(dlq.Payload), dlq.Error)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ListEvents(ctx context.Context, limit, offset int) ([]dto.KafkaEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := `
SELECT id, message_id, topic, msg_key, partition, "offset", payload, to_char(received_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
FROM kafka_events
ORDER BY id DESC
LIMIT $1 OFFSET $2
`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.KafkaEvent
	for rows.Next() {
		var (
			e       dto.KafkaEvent
			payload []byte
		)

		err = rows.Scan(
			&e.ID, &e.MessageID, &e.Topic, &e.Key, &e.Partition, &e.Offset, &payload, &e.ReceivedAt,
		)
		if err != nil {
			return nil, err
		}

		e.Payload = payload
		out = append(out, e)
	}

	return out, rows.Err()
}

func (r *Repository) ListDLQ(ctx context.Context, limit, offset int) ([]dto.KafkaDLQ, error) {
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	q := `
SELECT id, topic, msg_key, payload, error, to_char(received_at, 'YYYY-MM-DD"T"HH24:MI:SSOF')
FROM kafka_dlq
ORDER BY id DESC
LIMIT $1 OFFSET $2;
`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dto.KafkaDLQ
	for rows.Next() {
		var (
			d       dto.KafkaDLQ
			payload []byte
		)

		err = rows.Scan(&d.ID, &d.Topic, &d.Key, &payload, &d.Error, &d.ReceivedAt)
		if err != nil {
			return nil, err
		}

		d.Payload = payload
		out = append(out, d)
	}

	return out, rows.Err()
}

func (r *Repository) ResetAll(ctx context.Context) error {
	q := `
TRUNCATE kafka_events RESTART IDENTITY CASCADE;
TRUNCATE kafka_dlq RESTART IDENTITY CASCADE;
TRUNCATE employment_history RESTART IDENTITY CASCADE;
TRUNCATE employee_profile RESTART IDENTITY CASCADE;
`
	_, err := r.pool.Exec(ctx, q)
	if err != nil {
		return err
	}

	return nil
}
