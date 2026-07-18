#!/bin/bash
# Creates one database per service (database-per-service pattern).
# Runs automatically on first container start via /docker-entrypoint-initdb.d/
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE orders_db;
    CREATE DATABASE catalog_db;
    CREATE DATABASE inventory_db;
    CREATE DATABASE payments_db;
EOSQL

echo "Databases created: orders_db, catalog_db, inventory_db, payments_db"
