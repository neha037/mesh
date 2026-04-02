// Listen for badge update messages from the popup.
chrome.runtime.onMessage.addListener((msg) => {
  if (msg.type === "badge") {
    chrome.action.setBadgeText({ text: msg.text });
    chrome.action.setBadgeBackgroundColor({ color: msg.color });

    // Clear badge after 3 seconds.
    setTimeout(() => {
      chrome.action.setBadgeText({ text: "" });
    }, 3000);
  }
});
