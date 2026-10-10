#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_SCRIPT="$SCRIPT_DIR/frontend/docker-build.sh"
BACKEND_SCRIPT="$SCRIPT_DIR/backend/docker-build.sh"

action="${1:-build}"

case "$action" in
  down|build|up)
    "$FRONTEND_SCRIPT" "$action"
    "$BACKEND_SCRIPT" "$action"
    ;;
  *)
    echo "Usage: $(basename "$0") [down|build|up]" >&2
    exit 1
    ;;
esac
