#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER replaceme WITH PASSWORD 'replaceme';
	CREATE DATABASE replaceme_test;
	GRANT ALL PRIVILEGES ON DATABASE replaceme_test TO replaceme;
    CREATE DATABASE replaceme;
	GRANT ALL PRIVILEGES ON DATABASE replaceme TO replaceme;
	GRANT ALL PRIVILEGES ON DATABASE replaceme TO root;
	GRANT ALL PRIVILEGES ON DATABASE replaceme_test TO root;

EOSQL

for f in /docker-entrypoint-initdb.d/*.sql; do
	psql -U replaceme replaceme_test < $f
	psql -U replaceme replaceme < $f
done
