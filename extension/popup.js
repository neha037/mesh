const DEFAULT_API = "http://localhost:8080";

document.addEventListener("DOMContentLoaded", async () => {
  const status = document.getElementById("status");
  const recentList = document.getElementById("recent-list");

  const apiUrl = await getApiUrl();

  // Auto-save the current page immediately on popup open.
  setStatus(status, "Saving...", "saving");

  try {
    const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });

    const [result] = await chrome.scripting.executeScript({
      target: { tabId: tab.id },
      func: () => {
        const el = document.querySelector("article") || document.querySelector("main") || document.body;
        return el.innerText;
      },
    });

    const content = result.result || "";

    const resp = await fetch(`${apiUrl}/api/v1/ingest/raw`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        url: tab.url,
        title: tab.title,
        content: content,
      }),
    });

    if (!resp.ok) {
      const body = await resp.json().catch(() => ({}));
      throw new Error(body.error || `HTTP ${resp.status}`);
    }

    const data = await resp.json();
    if (data.updated) {
      setStatus(status, "Updated!", "updated");
    } else {
      setStatus(status, "Saved!", "success");
    }
    chrome.runtime.sendMessage({ type: "badge", text: "OK", color: "#16a34a" });
  } catch (err) {
    setStatus(status, `Error: ${err.message}`, "error");
    chrome.runtime.sendMessage({ type: "badge", text: "ERR", color: "#dc2626" });
  }

  // Load recent saves.
  loadRecent(apiUrl, recentList);
});

async function getApiUrl() {
  return new Promise((resolve) => {
    chrome.storage.sync.get({ apiUrl: DEFAULT_API }, (items) => {
      resolve(items.apiUrl);
    });
  });
}

async function loadRecent(apiUrl, listEl) {
  try {
    const resp = await fetch(`${apiUrl}/api/v1/nodes/recent`);
    if (!resp.ok) return;

    const nodes = await resp.json();
    listEl.innerHTML = "";

    nodes.slice(0, 10).forEach((node) => {
      const li = document.createElement("li");
      li.title = node.source_url || "";
      li.innerHTML = `${escapeHtml(node.title)} <span class="time">${formatTime(node.created_at)}</span>`;
      listEl.appendChild(li);
    });
  } catch {
    // API might not be running yet — silently ignore.
  }
}

function setStatus(el, text, className) {
  el.textContent = text;
  el.className = `status ${className}`;
}

function formatTime(iso) {
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}
