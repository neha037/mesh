---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

## Prerequisites

- **Docker** (24+) and **Docker Compose** (v2+)
- **Git**
- **Linux** with systemd (tested on Fedora)

## Installation

### 1. Clone the repository

```bash
git clone https://github.com/neha037/mesh.git
cd mesh
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and set secure passwords for `PG_PASSWORD` and `MINIO_PASSWORD`.

### 3. Run the installer

```bash
bash scripts/install.sh
```

This will:
- Add your user to the `docker` group (if needed)
- Create the `.env` file from the template (if missing)
- Install a **systemd user service** (`mesh.service`) that manages Docker containers
- Install a **desktop entry** for the system tray icon with autostart on login
- Enable user lingering so services start on boot

**Important:** If the installer added you to the docker group, you must **log out and log back in** before starting Mesh.

### 4. Start Mesh

```bash
systemctl --user start mesh
```

Or use the system tray icon (see [System Tray](system-tray)).

### 5. Verify

```bash
# Check service status
systemctl --user status mesh

# Test the API
curl http://localhost:8080/api/v1/nodes/recent
```

You should see an empty JSON array `[]` (no pages saved yet). The web dashboard is planned for Phase 4 -- for now, use the browser extension and API directly.

### 6. Install the browser extension

See [Browser Extension](browser-extension) for setup instructions.

## Next Steps

- [Save your first page](browser-extension) using the browser extension
- [Explore the API](api-reference) for programmatic access
- Check the [Troubleshooting](troubleshooting) guide if something isn't working
