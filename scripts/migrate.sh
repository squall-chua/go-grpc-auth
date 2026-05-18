#!/bin/bash
# Run MongoDB migrations using migrate-mongo
# Usage: scripts/migrate.sh [up|down|status]
# Requires: migrate-mongo installed globally (npm install -g migrate-mongo)

set -e

cd "$(dirname "$0")/.."

ACTION="${1:-up}"

case "$ACTION" in
  up|down|status)
    migrate-mongo "$ACTION"
    ;;
  *)
    echo "Usage: $0 [up|down|status]"
    exit 1
    ;;
esac
