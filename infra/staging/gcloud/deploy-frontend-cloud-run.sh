#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/frontend-cloud-run/.env}"
WEB_ENV_FILE="$SCRIPT_DIR/../../../frontend/web/.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Environment file tidak ditemukan: $ENV_FILE" >&2
  echo "Buat dan isi frontend-cloud-run/.env sebelum deploy." >&2
  exit 1
fi
if [[ -f "$WEB_ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$WEB_ENV_FILE"
  set +a
  unset BACKEND_BASE_URL
fi
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
exec python3 "$SCRIPT_DIR/frontend-cloud-run/deploy.py" "$@"
