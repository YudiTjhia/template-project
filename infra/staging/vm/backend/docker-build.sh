#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"
NETWORK_NAME=template_web_staging
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/.env}"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Environment file tidak ditemukan: $ENV_FILE" >&2
  echo "Buat dan isi file .env sebelum menjalankan script ini." >&2
  exit 1
fi
case "${1:-up}" in
  up)
    docker network inspect "$NETWORK_NAME" >/dev/null 2>&1 || docker network create "$NETWORK_NAME" >/dev/null
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build
    ;;
  down) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" down ;;
  restart) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --build --force-recreate ;;
  build) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" build ;;
  logs) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" logs -f ;;
  ps) docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps ;;
  *) echo "Usage: $(basename "$0") [up|down|restart|build|logs|ps]"; exit 1 ;;
esac
