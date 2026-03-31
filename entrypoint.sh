#!/bin/sh
set -e

# Run migrations
until /nakama/nakama migrate up --database.address "$DATABASE_URL"; do
  echo "Waiting for DB..."
  sleep 2
done

echo "Migrations complete!"

# Start Nakama
exec /nakama/nakama --name nakama1 --database.address "$DATABASE_URL" --session.token_expiry_sec 3600 --session.refresh_token_expiry_sec 3600 --logger.level INFO --runtime.path /nakama/data/modules
