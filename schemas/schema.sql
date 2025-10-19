-- Add new schema named "public"
CREATE SCHEMA IF NOT EXISTS "public";
-- Set comment to schema: "public"
COMMENT ON SCHEMA "public" IS 'standard public schema';
-- Create "employee_profile" table
CREATE TABLE "public"."employee_profile" ("employee_id" text NOT NULL, "first_name" text NULL, "last_name" text NULL, "birth_date" date NULL, "email" text NULL, "phone" text NULL, "title" text NULL, "department" text NULL, "grade" text NULL, "effective_from" date NULL, "updated_at" timestamptz NULL DEFAULT now(), PRIMARY KEY ("employee_id"));
-- Create "employment_history" table
CREATE TABLE "public"."employment_history" ("id" bigserial NOT NULL, "employee_id" text NOT NULL, "company" text NOT NULL, "position" text NULL, "period_from" date NOT NULL, "period_to" date NOT NULL, "stack" text[] NOT NULL DEFAULT '{}', "created_at" timestamptz NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create index "idx_history_employee_id" to table: "employment_history"
CREATE INDEX "idx_history_employee_id" ON "public"."employment_history" ("employee_id");
-- Create "kafka_dlq" table
CREATE TABLE "public"."kafka_dlq" ("id" bigserial NOT NULL, "topic" text NOT NULL, "msg_key" text NULL, "payload" jsonb NOT NULL, "error" text NOT NULL, "received_at" timestamptz NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create index "idx_kafka_dlq_topic_received_at" to table: "kafka_dlq"
CREATE INDEX "idx_kafka_dlq_topic_received_at" ON "public"."kafka_dlq" ("topic", "received_at" DESC);
-- Create "kafka_events" table
CREATE TABLE "public"."kafka_events" ("id" bigserial NOT NULL, "message_id" uuid NULL, "topic" text NOT NULL, "msg_key" text NULL, "partition" integer NULL, "offset" bigint NULL, "payload" jsonb NOT NULL, "received_at" timestamptz NULL DEFAULT now(), PRIMARY KEY ("id"));
-- Create index "idx_kafka_events_topic_received_at" to table: "kafka_events"
CREATE INDEX "idx_kafka_events_topic_received_at" ON "public"."kafka_events" ("topic", "received_at" DESC);
-- Create index "kafka_events_message_id_key" to table: "kafka_events"
CREATE UNIQUE INDEX "kafka_events_message_id_key" ON "public"."kafka_events" ("message_id");

-- Создаём роль "только чтение"
CREATE ROLE qa_readonly LOGIN PASSWORD 'pg-ro-secret' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;

-- Запретим PUBLIC на схему, если нужно
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- Дадим доступ на чтение к существующим объектам
GRANT USAGE ON SCHEMA public TO qa_readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO qa_readonly;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA public TO qa_readonly;

-- И на будущие объекты по умолчанию
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO qa_readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON SEQUENCES TO qa_readonly;
