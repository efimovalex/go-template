#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE replaceme_test;
    CREATE DATABASE replaceme;
	GRANT ALL PRIVILEGES ON DATABASE replaceme_test TO replaceme;
	GRANT ALL PRIVILEGES ON DATABASE replaceme TO replaceme;

EOSQL

for f in /docker-entrypoint-initdb.d/*.sql; do
	psql -U replaceme replaceme_test < $f
	psql -U replaceme replaceme < $f
done
