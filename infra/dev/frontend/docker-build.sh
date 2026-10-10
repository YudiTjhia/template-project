#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
WEB_ENV_FILE="$SCRIPT_DIR/../../../frontend/web/.env"
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/.env}"

if [[ ! -f "$ENV_FILE" ]]; then
  echo "Environment file tidak ditemukan: $ENV_FILE" >&2
  echo "Buat dan isi file .env sebelum menjalankan script ini." >&2
  exit 1
fi

if [[ -f "$WEB_ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$WEB_ENV_FILE"
  set +a
fi
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # Environment-specific API origins override local web defaults.
  # shellcheck disable=SC1090
  source "$ENV_FILE"
  set +a
fi

case "${1:-up}" in
  up) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d ;;
  down) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" down ;;
  build) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" build ;;
  logs) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs -f ;;
  ps) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps ;;
  *) echo "Usage: $(basename "$0") [down|build|up|logs|ps]"; exit 1 ;;
esac
