#!/bin/sh

set -e

echo "Run DB migrations"
echo $DB_URI

/app/migrate -path /app/migrations -database "$DB_URI" -verbose up

echo "Start the app"
exec "$@"