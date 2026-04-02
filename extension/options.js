document.addEventListener("DOMContentLoaded", () => {
  const input = document.getElementById("api-url");
  const saveBtn = document.getElementById("save-btn");
  const savedMsg = document.getElementById("saved-msg");

  // Load current setting.
  chrome.storage.sync.get({ apiUrl: "http://localhost:8080" }, (items) => {
    input.value = items.apiUrl;
  });

  saveBtn.addEventListener("click", () => {
    const url = input.value.replace(/\/+$/, ""); // strip trailing slashes
    try {
      new URL(url);
    } catch {
      alert("Invalid URL. Please enter a valid URL (e.g., http://localhost:8080).");
      return;
    }
    chrome.storage.sync.set({ apiUrl: url }, () => {
      savedMsg.style.display = "inline";
      setTimeout(() => {
        savedMsg.style.display = "none";
      }, 2000);
    });
  });
});
