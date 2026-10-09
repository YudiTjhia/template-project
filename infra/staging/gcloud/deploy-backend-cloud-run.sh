#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${ENV_FILE:-$SCRIPT_DIR/backend-cloud-run/.env}"
if [[ ! -f "$ENV_FILE" ]]; then
  echo "Environment file tidak ditemukan: $ENV_FILE" >&2
  echo "Buat dan isi backend-cloud-run/.env sebelum deploy." >&2
  exit 1
fi
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
exec python3 "$SCRIPT_DIR/backend-cloud-run/deploy.py" "$@"
