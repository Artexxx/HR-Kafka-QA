#!/usr/bin/env bash
set -euo pipefail

export POSTGRRES_DSN=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSL_MODE}

until pg_isready -d ${POSTGRRES_DSN} >/dev/null 2>&1; do
  sleep 1
done

./atlas migrate hash \
  --dir="file:///app/migrations"

./atlas migrate apply \
  --dir="file:///app/migrations" \
  --url=${POSTGRRES_DSN}

echo "Migrations successfully applied."
