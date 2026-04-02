#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=== Mesh Installer ==="
echo ""

# ──────────────────────────────────────────────
# 0. Dependency checks
# ──────────────────────────────────────────────
MISSING=()

if ! command -v docker &>/dev/null; then
    MISSING+=("docker (install: sudo dnf install docker-ce)")
fi

if ! docker-compose version &>/dev/null 2>&1 && ! docker compose version &>/dev/null 2>&1; then
    MISSING+=("docker-compose (install: sudo dnf install docker-compose)")
fi

if [ ${#MISSING[@]} -gt 0 ]; then
    echo "ERROR: Required dependencies are missing:"
    for dep in "${MISSING[@]}"; do
        echo "  - $dep"
    done
    exit 1
fi

if ! command -v yad &>/dev/null; then
    echo "WARNING: 'yad' is not installed. The system tray icon will not work."
    echo "  Install with: sudo dnf install yad"
    echo "  (Mesh services will still start without it.)"
    echo ""
fi

# ──────────────────────────────────────────────
# 1. Docker group
# ──────────────────────────────────────────────
NEEDS_RELOGIN=false
if ! groups | grep -q docker; then
    echo "Adding $USER to docker group (requires sudo)..."
    sudo usermod -aG docker "$USER"
    NEEDS_RELOGIN=true
else
    echo "User already in docker group."
fi

# ──────────────────────────────────────────────
# 1b. Docker daemon check
# ──────────────────────────────────────────────
if ! systemctl is-active --quiet docker 2>/dev/null; then
    echo ""
    echo "WARNING: Docker daemon is not running."
    echo "  Start it with:  sudo systemctl start docker"
    echo "  Enable on boot: sudo systemctl enable docker"
    echo ""
fi

# ──────────────────────────────────────────────
# 2. .env file
# ──────────────────────────────────────────────
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "Creating .env from .env.example..."
    cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    echo "IMPORTANT: Edit $PROJECT_DIR/.env and set secure passwords before starting."
fi

# ──────────────────────────────────────────────
# 3. Make scripts executable
# ──────────────────────────────────────────────
chmod +x "$SCRIPT_DIR/mesh-tray.sh"
chmod +x "$SCRIPT_DIR/mesh-services.sh"

# ──────────────────────────────────────────────
# 4. Install systemd user service (headless)
# ──────────────────────────────────────────────
echo "Installing systemd user service..."
mkdir -p ~/.config/systemd/user

sed "s|MESH_PROJECT_DIR|$PROJECT_DIR|g" \
    "$SCRIPT_DIR/mesh.service" > ~/.config/systemd/user/mesh.service

systemctl --user daemon-reload
systemctl --user enable mesh

# ──────────────────────────────────────────────
# 5. Install desktop entry (app menu + autostart)
# ──────────────────────────────────────────────
echo "Installing desktop entry..."
mkdir -p ~/.local/share/applications
mkdir -p ~/.config/autostart

DESKTOP_CONTENT=$(sed "s|MESH_PROJECT_DIR|$PROJECT_DIR|g" "$SCRIPT_DIR/mesh.desktop")

echo "$DESKTOP_CONTENT" > ~/.local/share/applications/mesh.desktop
echo "$DESKTOP_CONTENT" > ~/.config/autostart/mesh.desktop

if command -v update-desktop-database &>/dev/null; then
    update-desktop-database ~/.local/share/applications 2>/dev/null || true
fi

# ──────────────────────────────────────────────
# 6. Enable lingering so user services start on boot
# ──────────────────────────────────────────────
echo "Enabling user lingering for boot startup..."
sudo loginctl enable-linger "$USER"

# ──────────────────────────────────────────────
# Done
# ──────────────────────────────────────────────
echo ""
echo "=== Installation complete ==="
echo ""
echo "To start Mesh services now:"
echo "  systemctl --user start mesh"
echo ""
echo "The tray icon will appear automatically at next graphical login."
echo "To start it now:  $SCRIPT_DIR/mesh-tray.sh &"
echo ""
if [ "$NEEDS_RELOGIN" = true ]; then
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║  ACTION REQUIRED: Log out and log back in before starting   ║"
    echo "║  Mesh. Docker group membership is not effective until then.  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""
    echo "After re-login, start Mesh with:  systemctl --user start mesh"
elif ! docker info &>/dev/null; then
    echo ""
    echo "WARNING: 'docker info' failed. Docker may not be accessible."
    echo "  Check that the Docker daemon is running and your user has permission."
fi
