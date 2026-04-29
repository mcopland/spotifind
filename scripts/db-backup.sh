#!/usr/bin/env bash
set -euo pipefail

if ! docker compose ps db --status running --quiet 2>/dev/null | grep -q .; then
  echo "error: db container is not running (try: make up)" >&2
  exit 1
fi

mkdir -p backups
OUTFILE="backups/spotifind-$(date +%Y%m%d-%H%M%S).dump"
docker compose exec -T db pg_dump -U spotifind -d spotifind --format=custom > "$OUTFILE"
echo "backup written to $OUTFILE"
