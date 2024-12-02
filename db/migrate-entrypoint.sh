#!/bin/sh
set -ex  # Add -x for debugging

DB_USER=$(cat /run/secrets/postgres_user)
DB_PASS=$(cat /run/secrets/postgres_pswd)
DB_NAME=$(cat /run/secrets/postgres_db)

CONNECTION_STRING="postgres://${DB_USER}:${DB_PASS}@db:5432/${DB_NAME}?sslmode=disable"
/usr/local/bin/migrate \
    -path=/db/migrations \
    -database="${CONNECTION_STRING}" \
    up
