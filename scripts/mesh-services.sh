#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_DIR="$PROJECT_DIR/deploy"
ENV_FILE="$PROJECT_DIR/.env"

compose() {
    docker-compose -f "$COMPOSE_DIR/docker-compose.yml" --env-file "$ENV_FILE" "$@"
}

cleanup() {
    compose down 2>&1 | logger -t mesh || true
}
trap cleanup EXIT SIGTERM SIGINT

# Pre-flight: verify Docker is accessible.
if ! docker info &>/dev/null; then
    MSG="Mesh cannot start: Docker is not accessible. "
    if ! groups | grep -qw docker; then
        MSG+="User '$USER' is not in the 'docker' group. Run: sudo usermod -aG docker \$USER  then log out and back in."
    elif ! systemctl is-active --quiet docker 2>/dev/null; then
        MSG+="Docker daemon is not running. Run: sudo systemctl start docker"
    else
        MSG+="Unknown Docker error. Check: docker info"
    fi
    logger -t mesh "ERROR: $MSG"
    notify-send -u critical "Mesh" "$MSG" 2>/dev/null || true
    exit 2
fi

compose up -d --build postgres minio api 2>&1 | logger -t mesh

# Keep process alive and stream logs to systemd journal.
compose logs -f
