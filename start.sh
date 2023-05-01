#!/bin/sh

set -e

# run migrations
echo "Running migration..."
/app/migrate -path /app/migration -database "$DB_SOURCE" -verbose up

# run the main program
echo "Running the main program..."
exec "$@"
