#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
ICON="$PROJECT_DIR/extension/icons/icon128.png"

# --- Service health check ---
if systemctl --user is-failed mesh &>/dev/null; then
    FAIL_REASON=$(journalctl --user -u mesh -n 1 --no-pager -o cat 2>/dev/null || echo "unknown error")
    notify-send -u critical -i "$ICON" "Mesh" \
        "Services failed to start: $FAIL_REASON. Use tray menu to retry after fixing the issue." \
        2>/dev/null || true
fi

exec python3 "$SCRIPT_DIR/mesh-tray.py" "$ICON"
