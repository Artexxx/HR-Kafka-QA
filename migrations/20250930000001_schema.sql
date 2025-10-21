CREATE TABLE IF NOT EXISTS kafka_events (
                                                id          BIGSERIAL PRIMARY KEY,
                                                message_id  UUID UNIQUE,
                                                topic       TEXT NOT NULL,
                                                partition   INT,
                                                "offset"      BIGINT,
                                                payload     JSONB NOT NULL,
                                                received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS kafka_dlq (
                                             id          BIGSERIAL PRIMARY KEY,
                                             topic       TEXT NOT NULL,
                                             msg_key     TEXT,
                                             payload     JSONB NOT NULL,
                                             error       TEXT NOT NULL,
                                             received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS employee_profile (
                                                    employee_id    TEXT PRIMARY KEY,
                                                    first_name     TEXT,
                                                    last_name      TEXT,
                                                    birth_date     DATE,
                                                    email          TEXT,
                                                    phone          TEXT,
                                                    title          TEXT,
                                                    department     TEXT,
                                                    grade          TEXT,
                                                    effective_from DATE,
                                                    updated_at     TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS employment_history (
                                                      id          BIGSERIAL PRIMARY KEY,
                                                      employee_id TEXT NOT NULL,
                                                      company     TEXT NOT NULL,
                                                      position    TEXT,
                                                      period_from DATE NOT NULL,
                                                      period_to   DATE NOT NULL,
                                                      stack       TEXT[] NOT NULL DEFAULT '{}',
                                                      created_at  TIMESTAMPTZ DEFAULT now()
);

-- Helpful indexes
CREATE INDEX IF NOT EXISTS idx_kafka_events_topic_received_at ON kafka_events (topic, received_at DESC);
CREATE INDEX IF NOT EXISTS idx_kafka_dlq_topic_received_at    ON kafka_dlq (topic, received_at DESC);
CREATE INDEX IF NOT EXISTS idx_history_employee_id            ON employment_history (employee_id);
