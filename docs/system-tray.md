---
layout: default
title: System Tray
nav_order: 4
---

# System Tray

Mesh includes a system tray icon for controlling services from your desktop.

## Tray Icon Menu

Right-click the Mesh icon in your system tray to access:

| Action | Description |
|--------|-------------|
| **Open Dashboard** | Opens the API at `http://localhost:8080` (web UI coming in Phase 4) |
| **Start Services** | Starts the Mesh systemd service |
| **Stop Services** | Stops the Mesh systemd service |
| **Restart Services** | Restarts all services (useful after config changes) |
| **Quit** | Closes the tray icon (services keep running) |

Left-clicking the tray icon opens the dashboard directly.

## Autostart

After running `scripts/install.sh`, the tray icon starts automatically when you log in to your desktop session. The underlying Mesh service (`mesh.service`) also starts automatically on boot.

## Manual Service Control

You can also control the Mesh service directly from the terminal:

```bash
# Start services
systemctl --user start mesh

# Stop services
systemctl --user stop mesh

# Restart services
systemctl --user restart mesh

# Check status
systemctl --user status mesh

# View logs
journalctl --user -u mesh -f
```

## Wayland Compatibility

The tray icon uses AppIndicator3, which works on both X11 and Wayland. On GNOME Wayland, make sure the AppIndicator extension is installed and enabled:

```bash
sudo dnf install gnome-shell-extension-appindicator
gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com
```

Log out and back in after enabling the extension.
