#!/bin/sh

set -e

# run migrations
echo "Running migration..."
source /app/app.env
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

# run the main program
echo "Running the main program..."
exec "$@"
