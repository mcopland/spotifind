#!/usr/bin/env bash
set -euo pipefail

FILE="${1:-}"
if [[ -z "$FILE" ]]; then
  echo "usage: make db-restore FILE=backups/spotifind-<timestamp>.dump CONFIRM=1" >&2
  exit 1
fi
if [[ ! -s "$FILE" ]]; then
  echo "error: file not found or empty: $FILE" >&2
  exit 1
fi
if [[ "${CONFIRM:-}" != "1" ]]; then
  echo "error: this will destroy all current data -- rerun with CONFIRM=1 to proceed" >&2
  exit 1
fi

if ! docker compose ps db --status running --quiet 2>/dev/null | grep -q .; then
  echo "error: db container is not running (try: make up)" >&2
  exit 1
fi

docker compose exec -T db psql -U spotifind -d spotifind -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
docker compose exec -T db pg_restore -U spotifind -d spotifind --no-owner --no-acl < "$FILE"
echo "restore complete from $FILE"
