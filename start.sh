#!/bin/sh

set -e

# run migrations
echo "Running migrations..."
/app/migrate -path /app/migrations -database "$DB_SOURCE" -verbose up

# run the main program
echo "Running the main program..."
exec "$@"
