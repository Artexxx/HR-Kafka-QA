# ╔════════════════════════════════════════════════════════════════════════╗
#                 ⚙️ Database Migrations Targets
# ╚════════════════════════════════════════════════════════════════════════╝
include .env

pg_conn := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:5432/$(POSTGRES_DB)?sslmode=disable
pg_conn_dev := "docker://postgres/15"

# Путь, в котором хранятся все файлы схем.
SCHEMA_PATH := $(PWD)/schemas
# Путь, в котором хранятся все файлы версионных миграций.
MIGRATION_PATH := $(PWD)/migrations

ATLAS := atlas

atlas-install:
	@go install ariga.io/atlas/cmd/atlas@latest

atlas-help:
	@$(ATLAS) $(topic) help

# Генерирует хеши для схем и миграций.
# - Проверяет целостность и актуальность файлов схем и миграций.
atlas-hash:
	@$(ATLAS) migrate hash --dir=file://schemas
	@$(ATLAS) migrate hash --dir=file://migrations

# Сравнивает локальные схемы с целевой базой данных.
# - Выводит SQL-дифф между локальной схемой и целевой базой.
# - Мы исключаем таблицу atlas_schema_revisions, которую Atlas использует для хранения ревизий.
atlas-diff: atlas-hash
	@$(ATLAS) schema diff \
		--from $(pg_conn) \
		--to file://schemas \
		--dev-url $(pg_conn_dev) \
		--exclude 'atlas_schema_revisions' > diff.sql
	@cat diff.sql
	@rm diff.sql


# Создаёт новый файл миграции с заданным именем.
# Где смотреть:
# - Новый файл миграции появится в каталоге `$(MIGRATION_PATH)`.
atlas-new: atlas-hash
ifndef name
	$(error 'name is required')
else
	@$(ATLAS) migrate new $(name)\
		--dir file://migrations
endif


# Создаёт миграцию на основе различий между схемами.
# - Требует указания имени через переменную `name`.
# Где смотреть:
# - Новый файл миграции появится в каталоге `$(MIGRATION_PATH)`.
atlas-migrate-diff: atlas-hash
	@$(ATLAS) migrate diff $(name)\
		--dir "file://migrations" \
		--to "file://schemas" \
		--dev-url $(pg_conn_dev)

# Применяет изменения к базе данных напрямую (сравнивая схему).
# - Использует метод "сравнение схем".
# - Изменения применяются автоматически без подтверждения.
# Где смотреть:
# - Результаты будут видны в целевой базе данных.
atlas-diff-apply: atlas-hash
	@$(ATLAS) schema apply \
		--url $(pg_conn) \
		--to file://schemas \
		--dev-url $(pg_conn_dev) \
		--exclude 'atlas_schema_revisions' \
		--auto-approve

# Пробный запуск применения изменений (без сохранения).
# - Показывает, какие изменения будут применены.
# Где смотреть:
# - Результаты будут выведены в консоль без изменения базы данных.
atlas-diff-apply-dry-run: atlas-hash
	@$(ATLAS) schema apply \
		--url $(pg_conn) \
		--to file://schemas \
		--dev-url $(pg_conn_dev) \
		--exclude 'atlas_schema_revisions' \
		--dry-run

# Применяет миграции из каталога `migrations`.
# - Устанавливает последовательные изменения в целевую базу.
# Где смотреть:
# - Результаты будут видны в базе данных.
atlas-migrate: atlas-hash
	@$(ATLAS) migrate apply \
		--dir "file://migrations" \
		--url $(pg_conn)

# Создаёт SQL-файл schema.sql на основе миграций.
# - Использует миграции в `migrations/`.
# - Генерирует `schema.sql` в `schemas/`.
# Где смотреть:
# - Новый `schema.sql` появится в `schemas/`.
atlas-migrate-to-schema: atlas-hash
	@$(ATLAS) schema inspect \
		--url "file://migrations" \
		--dev-url $(pg_conn_dev) \
		--format "{{ sql . }}" > $(SCHEMA_PATH)/schema.sql

