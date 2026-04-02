---
layout: default
title: Troubleshooting
nav_order: 6
---

# Troubleshooting

## Docker is not accessible

**Symptom:** "Mesh cannot start: Docker is not accessible" or services fail to start.

**Fix:**

1. Ensure Docker is installed and running:
   ```bash
   sudo systemctl start docker
   sudo systemctl enable docker
   ```

2. Add your user to the docker group:
   ```bash
   sudo usermod -aG docker $USER
   ```

3. **Log out and log back in** for the group change to take effect.

4. Verify:
   ```bash
   docker info
   ```

**Note for SSSD/FreeIPA users:** If your account is managed by a centralized identity system, the docker group may not be picked up by your login session even after logging out. The Mesh systemd service uses `sg docker` to work around this automatically.

---

## System tray icon not showing

**Symptom:** No Mesh icon appears in the system tray after starting the tray script.

**Cause:** You are running GNOME on Wayland. The tray icon uses AppIndicator3 which requires an extension.

**Fix:**

```bash
sudo dnf install gnome-shell-extension-appindicator
gnome-extensions enable appindicatorsupport@rgcjonas.gmail.com
```

Log out and back in. The tray icon should now appear.

---

## Mesh service keeps failing

**Symptom:** `systemctl --user status mesh` shows `failed`.

**Fix:**

1. Check the logs:
   ```bash
   journalctl --user -u mesh -n 20 --no-pager
   ```

2. Common causes:
   - **Docker not accessible** -- see above
   - **Port already in use** -- another service is using port 8080, 5432, or 9000
   - **`.env` not configured** -- ensure `PG_PASSWORD` and `MINIO_PASSWORD` are set

3. After fixing the issue, reset and restart:
   ```bash
   systemctl --user reset-failed mesh
   systemctl --user start mesh
   ```

---

## Browser extension shows error

**Symptom:** Clicking the extension icon shows "Error: Failed to fetch" or similar.

**Causes and fixes:**

| Error | Cause | Fix |
|-------|-------|-----|
| `Failed to fetch` | API server is not running | Start the Mesh service: `systemctl --user start mesh` |
| `HTTP 400` | Page URL or title is empty | Try on a different page (some pages block content extraction) |
| `HTTP 500` | Database error | Check API logs: `journalctl --user -u mesh -f` |

---

## Extension can't save certain pages

**Symptom:** Extension shows an error on specific pages.

**Cause:** Chrome restricts extensions from running on certain pages:
- `chrome://` URLs (settings, extensions, etc.)
- `chrome-extension://` URLs
- Chrome Web Store pages

This is a Chrome security restriction and cannot be bypassed.

---

## Port conflicts

**Symptom:** Services fail to start because ports are already in use.

**Default ports:**

| Service | Port |
|---------|------|
| API | 8080 |
| PostgreSQL | 5432 |
| MinIO API | 9000 |
| MinIO Console | 9001 |

**Fix:** Stop the conflicting service, or change the port in `deploy/docker-compose.yml`.

```bash
# Find what's using a port
sudo ss -tlnp | grep 8080
```
