const DEFAULT_API = "http://localhost:8080";
const PER_PAGE = 20;

let apiUrl = DEFAULT_API;

document.addEventListener("DOMContentLoaded", async () => {
  apiUrl = await getApiUrl();
  loadNodes();
});

async function getApiUrl() {
  return new Promise((resolve) => {
    chrome.storage.sync.get({ apiUrl: DEFAULT_API }, (items) => {
      resolve(items.apiUrl);
    });
  });
}

async function loadNodes(cursor) {
  const tbody = document.getElementById("pages-body");
  const status = document.getElementById("status");
  const pagination = document.getElementById("pagination");

  try {
    let url = `${apiUrl}/api/v1/nodes?per_page=${PER_PAGE}`;
    if (cursor) {
      url += `&cursor=${encodeURIComponent(cursor)}`;
    }

    const resp = await fetch(url);
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);

    const data = await resp.json();

    // Only clear on first load (no cursor).
    if (!cursor) {
      tbody.innerHTML = "";
      status.className = "status";
    }

    if (data.nodes.length === 0 && !cursor) {
      status.textContent = "No saved pages yet.";
      status.className = "status info";
      pagination.innerHTML = "";
      return;
    }

    data.nodes.forEach((node) => {
      const tr = document.createElement("tr");
      tr.dataset.id = node.id;

      const hostname = node.source_url ? new URL(node.source_url).hostname : "";

      tr.innerHTML = `
        <td class="title-cell" title="${escapeAttr(node.title)}">${escapeHtml(node.title)}</td>
        <td class="source-cell">${node.source_url ? `<a href="${escapeAttr(node.source_url)}" target="_blank">${escapeHtml(hostname)}</a>` : ""}</td>
        <td class="date-cell">${formatDate(node.created_at)}</td>
        <td><button class="delete-btn">Delete</button></td>
      `;

      tr.querySelector(".delete-btn").addEventListener("click", () => deleteNode(node.id, tr));
      tbody.appendChild(tr);
    });

    pagination.innerHTML = "";
    if (data.has_more) {
      const loadMore = document.createElement("button");
      loadMore.textContent = "Load More";
      loadMore.addEventListener("click", () => loadNodes(data.next_cursor));
      pagination.appendChild(loadMore);
    }
  } catch (err) {
    status.textContent = `Failed to load pages: ${err.message}`;
    status.className = "status error";
  }
}

async function deleteNode(id, row) {
  if (!confirm("Delete this page? This cannot be undone.")) return;
  const btn = row.querySelector(".delete-btn");
  btn.disabled = true;
  btn.textContent = "...";

  try {
    const resp = await fetch(`${apiUrl}/api/v1/nodes/${id}`, { method: "DELETE" });
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    row.remove();
  } catch (err) {
    btn.disabled = false;
    btn.textContent = "Delete";
    alert(`Failed to delete: ${err.message}`);
  }
}

function formatDate(iso) {
  const d = new Date(iso);
  return d.toLocaleDateString(undefined, { year: "numeric", month: "short", day: "numeric" });
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str;
  return div.innerHTML;
}

function escapeAttr(str) {
  return str.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
}
