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
