#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER replaceme WITH PASSWORD 'replaceme';
	CREATE DATABASE replaceme_test;
	GRANT ALL PRIVILEGES ON DATABASE replaceme_test TO replaceme;
    CREATE DATABASE replaceme;
	GRANT ALL PRIVILEGES ON DATABASE replaceme TO replaceme;
EOSQL

psql -U replaceme replaceme_test < /docker-entrypoint-initdb.d/zinit-db.sql
psql -U replaceme replaceme < /docker-entrypoint-initdb.d/zinit-db.sql