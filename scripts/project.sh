#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker is not installed."
  exit 1
fi

if docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD=(docker compose)
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD=(docker-compose)
else
  echo "Docker Compose is not available."
  exit 1
fi

ACTION="${1:-up}"

case "$ACTION" in
  up)
    "${COMPOSE_CMD[@]}" up -d --build
    ;;
  down)
    "${COMPOSE_CMD[@]}" down
    ;;
  restart)
    "${COMPOSE_CMD[@]}" down
    "${COMPOSE_CMD[@]}" up -d --build
    ;;
  logs)
    "${COMPOSE_CMD[@]}" logs -f "${2:-}"
    ;;
  status)
    "${COMPOSE_CMD[@]}" ps
    ;;
  *)
    echo "Usage: ./scripts/project.sh [up|down|restart|logs|status]"
    exit 1
    ;;
esac
