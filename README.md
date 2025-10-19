# 🧩 Техническое задание: Учебный тренажёр Kafka для QA

## 1. Цель проекта

Создать учебный микросервисный стенд, имитирующий бизнес-процесс управления HR-профилями сотрудников.
Задача стенда — позволить начинающим QA инженерам на практике протестировать процесс обмена сообщениями через Kafka:

* отправку событий (продюсер),
* обработку сообщений тремя независимыми консьюмерами,
* запись и просмотр результатов в БД и UI,
* тестирование типичных сценариев: порядок доставки, дубликаты, DLQ, лаги и т. д.

---

## 2. Общая архитектура

### Компоненты:

| Компонент                 | Назначение                                            | Технологии                            |
|---------------------------|-------------------------------------------------------|---------------------------------------|
| Kafka (1 брокер)          | Брокер сообщений (Kafka KRaft Mode, без ZooKeeper)    | bitnami/kafka:latest                  |
| Kafka UI                  | Веб-панель для просмотра топиков и сообщений          | provectuslabs/kafka-ui                |
| Producer API              | REST API для отправки событий (FastAPI)               | Python 3.12, FastAPI, Confluent-Kafka |
| Consumer A                | Обрабатывает события о должностях (hr.positions)      | Python 3.12, Confluent-Kafka          |
| Consumer B                | Обрабатывает личные данные (hr.personal)              | Python 3.12, Confluent-Kafka          |
| Consumer C                | Обрабатывает историю работы (hr.history)              | Python 3.12, Confluent-Kafka          |
| PostgreSQL (отдельная БД) | Хранение событий, профилей сотрудников, истории и DLQ | PostgreSQL 15+                        |

Все сервисы разворачиваются в отдельном docker-compose стеке с собственной сетью, чтобы не конфликтовать с другими
проектами на VDS.

---

## 3. Функциональные требования

### 3.1. Producer API

* Поднять сервис на FastAPI (порт 7011).
* Реализовать три POST-эндпоинта:

    * /produce/position → топик hr.positions
    * /produce/personal → топик hr.personal
    * /produce/history → топик hr.history
* Каждый эндпоинт принимает JSON-тело (payload) с обязательными полями:

    * employee_id — строка
    * message_id — UUID (для идемпотентности)
    * Остальные поля — см. ниже.
* Producer кодирует payload в JSON и публикует в Kafka с ключом employee_id.
* Настройки Kafka берутся из переменных окружения:

KAFKA_BOOTSTRAP=qa-kafka:9092
PRODUCER_CLIENT_ID=qa-producer

#### Примеры payload:

1) /produce/position

{
"message_id": "uuid-1",
"employee_id": "e-1024",
"title": "QA Engineer",
"department": "Quality Assurance",
"grade": "Middle",
"effective_from": "2025-10-01"
}

2) /produce/personal

{
"message_id": "uuid-2",
"employee_id": "e-1024",
"first_name": "Anna",
"last_name": "Ivanova",
"birth_date": "1994-06-12",
"contacts": { "email": "anna@example.com", "phone": "+33..." }
}

3) /produce/history

{
"message_id": "uuid-3",
"employee_id": "e-1024",
"company": "Acme Corp",
"position": "QA",
"period": { "from": "2022-07-01", "to": "2025-09-30" },
"stack": ["Python", "Pytest", "Postman"]
}

---

### 3.2. Consumer-сервисы

Каждый консьюмер — отдельный контейнер с независимым group_id.
Конфигурация общая:

KAFKA_BOOTSTRAP=qa-kafka:9092
TOPIC=hr.positions|hr.personal|hr.history
GROUP_ID=consumer_positions|consumer_personal|consumer_history
PG_DSN=postgresql://kafka_lab_user:password@db_host:5432/qa_kafka_lab

#### Логика консьюмера:

1. Подключается к Kafka, слушает свой топик.
2. При получении сообщения:

    * Валидирует JSON.
    * Проверяет наличие message_id в таблице lab.kafka_events → если уже было — игнорирует (идемпотентность).
    * Записывает «сырое» событие в таблицу lab.kafka_events (topic, key, payload, offset, received_at).
    * В зависимости от типа топика обновляет/вставляет данные:

* positions → обновляет должность и отдел в lab.employee_profile
    * personal → обновляет личные данные в lab.employee_profile
    * history → добавляет запись в lab.employment_history

3. В случае ошибки при разборе JSON — пишет запись в lab.kafka_dlq.

---

## 4. Структура БД

База должна быть отдельной, не пересекающейся с другими проектами на сервере.

База данных: qa_kafka_lab
Пользователь: kafka_lab_user
Схема: lab

SQL-инициализация (выполнить через миграции или при старте контейнера):

CREATE SCHEMA IF NOT EXISTS lab;

CREATE TABLE lab.kafka_events (
id BIGSERIAL PRIMARY KEY,
message_id UUID UNIQUE,
topic TEXT NOT NULL,
msg_key TEXT,
partition INT,
offset BIGINT,
payload JSONB NOT NULL,
received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE lab.kafka_dlq (
id BIGSERIAL PRIMARY KEY,
topic TEXT NOT NULL,
msg_key TEXT,
payload JSONB NOT NULL,
error TEXT NOT NULL,
received_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE lab.employee_profile (
employee_id TEXT PRIMARY KEY,
first_name TEXT,
last_name TEXT,
birth_date DATE,
email TEXT,
phone TEXT,
title TEXT,
department TEXT,
grade TEXT,
effective_from DATE,
updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE lab.employment_history (
id BIGSERIAL PRIMARY KEY,
employee_id TEXT NOT NULL,
company TEXT NOT NULL,
position TEXT,
period_from DATE,
period_to DATE,
stack TEXT[],
created_at TIMESTAMPTZ DEFAULT now()
);

---

## 5. Инфраструктура и деплой

### Docker-Compose структура

Создать директорию /opt/qa-kafka-lab/ с файлами:

/opt/qa-kafka-lab/
├─ docker-compose.yml
├─ producer/
│ ├─ Dockerfile
│ ├─ requirements.txt
│ └─ app.py
├─ consumer-positions/
│ ├─ Dockerfile
│ └─ app.py
├─ consumer-personal/
│ ├─ Dockerfile
│ └─ app.py
├─ consumer-history/
│ ├─ Dockerfile
│ └─ app.py
└─ init_db/
└─ init.sql

### Docker Compose

* Все контейнеры объединяются в сеть qa-kafka-net.
* Kafka не должен иметь ports: наружу (только внутри сети).
* Kafka UI доступен на http://localhost:7008 (или через Nginx /simulators/kafka-ui/).
* Producer доступен на http://localhost:7011 (или /simulators/kafka-producer/).

Порты 7008 и 7011 согласовать, чтобы не конфликтовали с уже занятыми.

---

## 6. Нефункциональные требования

| Категория          | Требование                                                                                                    |
|--------------------|---------------------------------------------------------------------------------------------------------------|
| Безопасность       | Kafka не открывается наружу; только UI и Producer проксируются через Nginx.                                   |
| Изоляция           | Проект в отдельной директории /opt/qa-kafka-lab/, отдельная docker-сеть, отдельная база qa_kafka_lab.         |
| Отказоустойчивость | При падении консьюмера данные не теряются — при перезапуске он дочитывает оффсеты.                            |
| Идемпотентность    | Повторная отправка того же message_id не должна дублировать записи в employee_profile или employment_history. |
| Логирование        | Все сервисы выводят логи в stdout (достаточно docker logs -f), без чувствительных данных.                     |
| Установка          | Однокомандный деплой: docker compose up -d --build                                                            |
| Совместимость      | Совместимо с Docker >= 24 и PostgreSQL >= 15                                                                  |
| Ресурсы            | Проект лёгкий: ≤ 1 ГБ RAM, ≤ 1 CPU на VDS.                                                                    |

---

## 7. Дополнительные требования

1. В README.md описать:

* структуру проекта,
    * переменные окружения,
    * команды запуска/остановки,
    * примеры curl запросов для тестирования.

2. Сделать простой health-endpoint /health в Producer и Consumer-ах.
3. Подготовить SQL-скрипт init_db/init.sql для первичного создания схемы и таблиц.
4. Поддержать ручную перезагрузку данных: docker compose down -v полностью очищает хранилище Kafka и БД (для QA-практики
   «с нуля»).

---

## 8. Примеры тестовых сценариев для QA (должны быть реализуемы на этом стенде)

| № | Сценарий                                                                                                                       | Цель проверки                                       |
|---|--------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| 1 | Отправить событие /produce/personal, убедиться, что consumer-personal получил сообщение и запись появилась в employee_profile. | Проверка базового обмена.                           |
| 2 | Отправить 3 события подряд для одного сотрудника — проверить порядок в kafka_events.                                           | Демонстрация партиционирования и порядка сообщений. |
| 3 | Отправить событие с тем же message_id — проверить, что дубликат не обработан повторно.                                         | Идемпотентность.                                    |
| 4 | Отправить некорректный JSON — проверить появление в kafka_dlq.                                                                 | Обработка ошибок.                                   |
| 5 | Остановить consumer-history, отправить события, потом запустить — убедиться, что он дочитал отставшие.                         | Поведение при оффлайне.                             |

---

## 9. Критерии приёмки

✅ Все контейнеры поднимаются без ошибок одной командой.
✅ Kafka UI доступен (и видны топики hr.positions, hr.personal, hr.history).
✅ Сообщения, отправленные продюсером, попадают в соответствующие топики и таблицы БД.
✅ Повторные message_id не дублируют записи.
✅ Ошибочные JSON-события фиксируются в kafka_dlq.
✅ Логи читаемы через docker compose logs -f.
✅ При docker compose down -v вся инфраструктура чистится.

---

## 10. Будущее расширение (не в рамках первой версии)

* DLQ-консьюмер (перемещение обратно в рабочие топики).
* Общий "Events Dashboard" (простая веб-страница для просмотра последних событий).
* Поддержка Redpanda как альтернативы Kafka.
* Генератор случайных событий для нагрузки (10 000 сообщений).

Отдельно требования по payload:

(https://chatgpt.com/s/t_68eb545255b08191aa02310fd499f54e)

# 📦 Требования к структуре сообщений (payload)

## Общие правила для всех сообщений

Каждое сообщение, отправляемое через /produce/*, должно содержать:

| Поле        | Тип           | Обязательно | Описание                                                                                 | Пример                                 |
|-------------|---------------|-------------|------------------------------------------------------------------------------------------|----------------------------------------|
| message_id  | string (UUID) | ✅           | Уникальный идентификатор события. Используется для идемпотентности.                      | "2a5b0df2-3d2f-4a6a-b5a9-3cbdf509b1f5" |
| employee_id | string        | ✅           | Уникальный идентификатор сотрудника, по которому объединяются данные из разных сервисов. | "e-1024"                               |

> ⚠️ Если message_id или employee_id отсутствует или имеет неверный тип — сообщение считается невалидным и должно быть
> отправлено в DLQ с ошибкой missing_required_field или invalid_type.

---

## 1️⃣ /produce/personal → Топик hr.personal

### Назначение:

Передача личных данных сотрудника (обновление профиля).

### Поля:

| Поле       | Тип                 | Обязательно                                 | Описание                             | Пример                                                    |
|------------|---------------------|---------------------------------------------|--------------------------------------|-----------------------------------------------------------|
| first_name | string              | ✅                                           | Имя сотрудника                       | "Anna"                                                    |
| last_name  | string              | ✅                                           | Фамилия сотрудника                   | "Ivanova"                                                 |
| birth_date | string (YYYY-MM-DD) | ✅                                           | Дата рождения в формате ISO          | "1994-06-12"                                              |
| contacts   | object              | ⚙️ (обязательно само поле, но не все ключи) | Контактная информация (email, phone) | { "email": "anna@example.com", "phone": "+33-123-45-67" } |

#### Валидационные правила:

| № | Условие                                      | Результат                |
|---|----------------------------------------------|--------------------------|
| 1 | first_name или last_name пусты ("" или null) | ❌ invalid_value          |
| 2 | birth_date не соответствует YYYY-MM-DD       | ❌ invalid_date_format    |
| 3 | contacts отсутствует полностью               | ❌ missing_required_field |
| 4 | contacts.email невалиден (нет @)             | ❌ invalid_email          |
| 5 | contacts.phone не строка                     | ❌ invalid_type           |
| 6 | Все обязательные поля есть и валидны         | ✅ OK                     |

---

## 2️⃣ /produce/position → Топик hr.positions

### Назначение:

Обновление информации о должности и подразделении сотрудника.

### Поля:

| Поле           | Тип                 | Обязательно | Описание                         | Пример              |
|----------------|---------------------|-------------|----------------------------------|---------------------|
| title          | string              | ✅           | Название должности               | "QA Engineer"       |
| department     | string              | ✅           | Подразделение                    | "Quality Assurance" |
| grade          | string (enum)       | ⚙️          | Уровень грейда сотрудника        | "Middle"            |
| effective_from | string (YYYY-MM-DD) | ✅           | Дата вступления изменений в силу | "2025-10-01"        |

#### Валидационные правила:

| № | Условие                                                                 | Результат                |
|---|-------------------------------------------------------------------------|--------------------------|
| 1 | title отсутствует или пуст                                              | ❌ missing_required_field |
| 2 | department отсутствует или не строка                                    | ❌ invalid_type           |
| 3 | grade не входит в список ["Junior", "Middle", "Senior", "Lead", "Head"] | ❌ invalid_enum_value     |
| 4 | effective_from не соответствует YYYY-MM-DD                              | ❌ invalid_date_format    |
| 5 | Все обязательные поля валидны                                           | ✅ OK                     |

---

## 3️⃣ /produce/history → Топик hr.history

### Назначение:

Добавление информации о предыдущих местах работы сотрудника.

### Поля:

| Поле     | Тип              | Обязательно | Описание                             | Пример                                       |
|----------|------------------|-------------|--------------------------------------|----------------------------------------------|
| company  | string           | ✅           | Название компании                    | "Acme Corp"                                  |
| position | string           | ✅           | Должность на предыдущем месте работы | "QA Engineer"                                |
| period   | object           | ✅           | Объект с периодом работы             | { "from": "2022-07-01", "to": "2025-09-30" } |
| stack    | array of strings | ⚙️          | Технологический стек                 | ["Python", "Pytest", "Postman"]              |

#### Валидационные правила:

| № | Условие                                                | Результат                |
|---|--------------------------------------------------------|--------------------------|
| 1 | company отсутствует или пуст                           | ❌ missing_required_field |
| 2 | period.from или period.to отсутствует                  | ❌ missing_required_field |
| 3 | period.from или period.to не YYYY-MM-DD                | ❌ invalid_date_format    |
| 4 | period.to < period.from (дата окончания раньше начала) | ❌ invalid_period         |
| 5 | stack не массив строк                                  | ❌ invalid_type           |
| 6 | Все обязательные поля валидны                          | ✅ OK                     |

---

## 🔥 Поведение при ошибках

| Сценарий                                                                                                     | Действие                                                                                                  |
|--------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------|
| Ошибка валидации на стороне продюсера (например, неверный JSON)                                              | Возвратить HTTP 400 с сообщением об ошибке (FastAPI автоматически валидирует).                            |
| Ошибка при обработке на стороне консьюмера (например, нарушен формат даты или отсутствует обязательное поле) | Сообщение не коммитить, записать в таблицу lab.kafka_dlq с указанием причины (error) и исходного payload. |
| Повторная отправка того же message_id                                                                        | Сообщение пропускается, но логируется (для QA наблюдения).                                                |

---

## 🧠 Примеры валидных и невалидных сообщений

### ✅ Валидное /produce/personal

{
"message_id": "fd3f9edb-37a0-41ff-9a1e-7a3c9a5c820a",
"employee_id": "e-1024",
"first_name": "Anna",
"last_name": "Ivanova",
"birth_date": "1994-06-12",
"contacts": { "email": "anna@example.com", "phone": "+33-123-45-67" }
}

### ❌ Невалидное /produce/personal (плохая дата)

Alexey, [12.10.2025 9:59]
{
"message_id": "fd3f9edb-37a0-41ff-9a1e-7a3c9a5c820a",
"employee_id": "e-1024",
"first_name": "Anna",
"last_name": "Ivanova",
"birth_date": "94-06-12"
}

→ Ошибка: invalid_date_format

---

### ✅ Валидное /produce/position

{
"message_id": "9f1a7c6e-4402-40cf-b8db-55cb08fba4c1",
"employee_id": "e-1024",
"title": "QA Engineer",
"department": "Testing",
"grade": "Middle",
"effective_from": "2025-10-01"
}

### ❌ Невалидное /produce/position (неверный грейд)

{
"message_id": "9f1a7c6e-4402-40cf-b8db-55cb08fba4c1",
"employee_id": "e-1024",
"title": "QA Engineer",
"department": "Testing",
"grade": "Expert"
}

→ Ошибка: invalid_enum_value

---

### ✅ Валидное /produce/history

{
"message_id": "b3e9e30a-4b8f-4c11-9f7b-0b1c41f25842",
"employee_id": "e-1024",
"company": "Acme Corp",
"position": "QA",
"period": { "from": "2022-07-01", "to": "2025-09-30" },
"stack": ["Python", "Pytest"]
}

### ❌ Невалидное /produce/history (период некорректен)

{
"message_id": "b3e9e30a-4b8f-4c11-9f7b-0b1c41f25842",
"employee_id": "e-1024",
"company": "Acme Corp",
"position": "QA",
"period": { "from": "2025-09-30", "to": "2022-07-01" }
}

→ Ошибка: invalid_period

---

## 💾 Логика записи в DLQ

Если сообщение не проходит валидацию на уровне консьюмера:

* записывается в таблицу lab.kafka_dlq:

INSERT INTO lab.kafka_dlq (topic, msg_key, payload, error, received_at)
VALUES ('hr.history', 'e-1024', '{"..."}', 'invalid_period', now());

Here's a complete `README.md` you can drop into the repo. Оно под ваш текущий стек (Go + fasthttp + Sarama + Postgres),
включает payload-требования, запуск, миграции, curl/HTTP, Postman, QA-сценарии и троблшутинг.

---

# QA Kafka Lab — HR Profiles Trainer

Учебный стенд для QA по Kafka: публикация событий продюсером, обработка консьюмерами, запись в Postgres и ручки для
чтения/проверки. Проект собран на **Go (fasthttp + zerolog)**, **Kafka (Sarama)** и **PostgreSQL**.

## Содержание

* [Архитектура](#архитектура)
* [Требования к payload](#требования-к-payload)
* [Структура репозитория](#структура-репозитория)
* [Переменные окружения](#переменные-окружения)
* [Быстрый старт](#быстрый-старт)
* [Миграции (goose)](#миграции-goose)
* [API (HTTP)](#api-http)
* [Curl примеры](#curl-примеры)
* [Postman коллекция](#postman-коллекция)
* [QA сценарии](#qa-сценарии)
* [Троблшутинг](#троблшутинг)
* [Лицензия](#лицензия)

---

## Архитектура

Компоненты текущей версии:

* **HTTP API** (порт конфигурируемый): CRUD по БД, плюс продюсер-ручки (`/producer/*`) — отправляют события в Kafka.
* **Kafka Producer**: обёртка над `sarama.SyncProducer`, шлёт в топики `hr.personal`, `hr.positions`, `hr.history`.
* **(дальше)** Консьюмер(ы) — отдельные процессы/сервисы (проект подготовлен под подключение; спецификация ниже).
* **PostgreSQL**: таблицы `employee_profile`, `employment_history`, `kafka_events`, `kafka_dlq`.

Graceful shutdown: SIGINT/SIGTERM → корректная остановка сервера и освобождение ресурсов.

---

## Требования к payload

Кратко (полные таблицы правил приведены в ТЗ и повторены ниже).

Общие обязательные поля для всех `/producer/*`:

* `message_id` — UUID (string). Идемпотентность.
* `employee_id` — string. Ключ маршрутизации и агрегации.

### `/producer/personal` → `hr.personal`

```json
{
  "message_id": "fd3f9edb-37a0-41ff-9a1e-7a3c9a5c820a",
  "employee_id": "e-1024",
  "first_name": "Anna",
  "last_name": "Ivanova",
  "birth_date": "1994-06-12",
  "contacts": {
    "email": "anna@example.com",
    "phone": "+33-123-45-67"
  }
}
```

Ошибки: `invalid_value`, `invalid_date_format`, `missing_required_field`, `invalid_email`, `invalid_type`.

### `/producer/position` → `hr.positions`

```json
{
  "message_id": "9f1a7c6e-4402-40cf-b8db-55cb08fba4c1",
  "employee_id": "e-1024",
  "title": "QA Engineer",
  "department": "Testing",
  "grade": "Middle",
  "effective_from": "2025-10-01"
}
```

Ошибки: `missing_required_field`, `invalid_type`, `invalid_enum_value`, `invalid_date_format`.

### `/producer/history` → `hr.history`

```json
{
  "message_id": "b3e9e30a-4b8f-4c11-9f7b-0b1c41f25842",
  "employee_id": "e-1024",
  "company": "Acme Corp",
  "position": "QA",
  "period": {
    "from": "2022-07-01",
    "to": "2025-09-30"
  },
  "stack": [
    "Python",
    "Pytest"
  ]
}
```

Ошибки: `missing_required_field`, `invalid_date_format`, `invalid_period`, `invalid_type`.

Если сообщение невалидно у консьюмера — **не коммитится**, записывается в `kafka_dlq` (payload + причина), потом можно
разбирать.

---

## Структура репозитория

```
.
├─ cmd/
│  └─ qa-kafka-api/
│     └─ main.go                # запуск сервиса (HTTP + продюсер)
├─ internal/
│  ├─ api/                      # fasthttp + handlers, middleware
│  ├─ config/                   # конфиг-файлы и структуры
│  ├─ dto/                      # DTO сущности (Profile, History, Events, DLQ)
│  ├─ exchange/
│  │  └─ producer/              # HRProducer (sarama SyncProducer)
│  └─ repository/
│     ├─ events/                # kafka_events / kafka_dlq
│     ├─ history/               # employment_history
│     └─ profile/               # employee_profile
├─ library/
│  ├─ pg/                       # pgxpool init + tracer
│  └─ yamlreader/               # загрузка YAML-конфига
├─ migrations/                  # goose миграции (SQL)
├─ config/
│  └─ application-local.yaml    # пример конфига
├─ docker-compose.yml           # (если используете docker-оркестрацию локально)
├─ Makefile
└─ README.md
```

---

## Переменные окружения

В `config/application-local.yaml` (или через ENV):

```yaml
userAPI:
  port: 7011

postgres:
  conn: "postgres://user:pass@localhost:5432/qa_kafka_lab?sslmode=disable"

kafka:
  bootstrap: "localhost:9092"
  producerClientID: "qa-producer"
  topics:
    personal: "hr.personal"
    positions: "hr.positions"
    history: "hr.history"
```

ENV override (опционально):

```
CONFIG_PATH=./config/application-local.yaml
```

---

## Быстрый старт

### Локально (без контейнеров)

1. Запустите PostgreSQL и выполните миграции:

```bash
make goose-up
```

2. Запустите Kafka (любой способ: локально / docker).

3. Соберите и запустите сервис:

```bash
make run
# или
go run ./cmd/qa-kafka-api
```

Логи покажут активный порт и bootstrap Kafka. Health: `GET /health` (если добавлен, см. ниже API).

### Docker Compose (при необходимости)

Добавьте `docker-compose.yml` для Kafka/Postgres/Kafka UI (не включён сюда, если нужен — заведите свой, чтобы порты не
конфликтовали). Затем:

```bash
docker compose up -d --build
docker compose logs -f
```

---

## Миграции (goose)

Убедитесь, что `goose` установлен:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Команды:

```bash
make goose-status
make goose-up
make goose-down
```

`Makefile` должен содержать переменные подключения к вашей БД. Пример:

```makefile
GOOSE_DRIVER=postgres
GOOSE_DBSTRING=postgres://user:pass@localhost:5432/qa_kafka_lab?sslmode=disable

goose-up:
	goose -dir ./migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) up

goose-down:
	goose -dir ./migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) down

goose-status:
	goose -dir ./migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) status

run:
	go run ./cmd/qa-kafka-api
```

---

## API (HTTP)

База URL: `http://localhost:<port>` (по умолчанию `7011`, если так задано в конфиге).

### Producer

* `POST /producer/personal`
* `POST /producer/position`
* `POST /producer/history`

Тело — как в [Требования к payload](#требования-к-payload).

### Profiles (CRUD)

* `POST /profiles` — создать профиль (минимум `employee_id`; остальные поля опциональны)
* `PUT /profiles/{employee_id}` — частичное обновление
* `DELETE /profiles/{employee_id}`
* `GET /profiles?limit=50&offset=0`
* `GET /profiles/{employee_id}`

### History (CRUD)

* `POST /history`
* `PUT /history/{id}`
* `DELETE /history/{id}`
* `GET /history/{employee_id}?limit=50&offset=0`

### Events / DLQ (Read)

* `GET /events?limit=50&offset=0&pretty=1`
* `GET /dlq?limit=50&offset=0&pretty=1`

Параметр `pretty=1` красиво форматирует вложенный `payload` (если он JSON).

---

## Curl примеры

```bash
# PERSONAL
curl -sS -X POST http://localhost:7011/producer/personal \
 -H 'Content-Type: application/json' \
 -d '{
  "message_id":"fd3f9edb-37a0-41ff-9a1e-7a3c9a5c820a",
  "employee_id":"e-1024",
  "first_name":"Anna",
  "last_name":"Ivanova",
  "birth_date":"1994-06-12",
  "contacts":{"email":"anna@example.com","phone":"+33-123-45-67"}
 }' | jq

# POSITION
curl -sS -X POST http://localhost:7011/producer/position \
 -H 'Content-Type: application/json' \
 -d '{
  "message_id":"9f1a7c6e-4402-40cf-b8db-55cb08fba4c1",
  "employee_id":"e-1024",
  "title":"QA Engineer",
  "department":"Testing",
  "grade":"Middle",
  "effective_from":"2025-10-01"
 }' | jq

# HISTORY
curl -sS -X POST http://localhost:7011/producer/history \
 -H 'Content-Type: application/json' \
 -d '{
  "message_id":"b3e9e30a-4b8f-4c11-9f7b-0b1c41f25842",
  "employee_id":"e-1024",
  "company":"Acme Corp",
  "position":"QA",
  "period":{"from":"2022-07-01","to":"2025-09-30"},
  "stack":["Python","Pytest"]
 }' | jq

# Просмотр событий (payload красиво)
curl -sS "http://localhost:7011/events?pretty=1&limit=20" | jq
```

---

## Postman коллекция

Готовая коллекция с примерами эндпоинтов и тел запросов (включая негативные кейсы) — *
*`postman/QA-Kafka-Lab.postman_collection.json`** в репозитории. Импортируйте в Postman и меняйте `{{baseUrl}}` на
`http://localhost:7011`.

---

## QA сценарии

1. **Базовый обмен**: отправить `/producer/personal` → убедиться в записи в `employee_profile` и появлении события в
   `kafka_events`.
2. **Порядок**: отправить 3 события подряд по одному `employee_id` → проверить порядок в `kafka_events` (
   partition/offset).
3. **Идемпотентность**: отправить событие с тем же `message_id` → убедиться, что второй раз не изменяет запись/не
   дублирует.
4. **Ошибки**: отправить заведомо невалидный JSON → запись в `kafka_dlq` с причиной (например, `invalid_date_format`).
5. **Отставание**: остановить консьюмер истории, отправить пачку `history`, затем поднять консьюмер → он дочитывает и
   пишет в БД.

---

## Троблшутинг

* **Приложение «зависает» при остановке**
  В `main.go` используется упрощённое завершение: по SIGINT/SIGTERM — отменяется контекст, `errgroup.Wait()` и
  корректный `Shutdown()` HTTP-сервера. Проверьте, что нет зависших горутин и что Kafka/DB доступны (иначе соединения
  могут держать закрытие чуть дольше).

* **payload отображается base64 в /events**
  В API включена опция `?pretty=1`, а в DTO `payload` хранится как `json.RawMessage` (рekomenduем). Тогда в ответе будет
  красиво отформатированный JSON вместо base64.

* **Kafka недоступна**
  Проверьте `KAFKA_BOOTSTRAP` и что брокер слушает внутри вашей сети/на нужном адресе. Для локального single-node Kafka
  в Docker используйте корректные `listeners/advertised.listeners`.

* **Миграции не применяются**
  Проверьте строку подключения в `Makefile`/ENV и наличие `goose`. Запустите `make goose-status` и проверьте логи.

* **Дубликаты событий**
  Дубликаты «шума» в Kafka нормальны. Идемпотентность обеспечивается проверкой `message_id` в `kafka_events` (и/или на
  уровне бизнес-таблиц).

---

## Лицензия

MIT. Делайте с проектом что хотите — форкайте, дополняйте и ломайте (в учебных целях 😄).

---

### Приложение: Полные правила валидации payload

(Сводные таблицы из ТЗ, чтобы всё было в одном месте.)

**Общее:** `message_id` (UUID, string) и `employee_id` (string) — **обязательны**. Ошибки: `missing_required_field`,
`invalid_type`.

**/producer/personal:**

* first_name, last_name — обязательны, непустые → иначе `invalid_value`.
* birth_date — `YYYY-MM-DD` → иначе `invalid_date_format`.
* contacts — объект обязателен; внутри email (при наличии) должен содержать `@` → иначе `invalid_email`.

**/producer/position:**

* title — обязателен → `missing_required_field`.
* department — string → иначе `invalid_type`.
* grade — enum: `Junior|Middle|Senior|Lead|Head` → иначе `invalid_enum_value`.
* effective_from — `YYYY-MM-DD` → иначе `invalid_date_format`.

**/producer/history:**

* company — обязателен.
* period.{from,to} — обязательны и `YYYY-MM-DD`.
* `to < from` → `invalid_period`.
* stack — массив строк → иначе `invalid_type`.

**DLQ запись (пример):**

```sql
INSERT INTO kafka_dlq (topic, msg_key, payload, error, received_at)
VALUES ('hr.history', 'e-1024', '{"..."}'::jsonb, 'invalid_period', now());
```
