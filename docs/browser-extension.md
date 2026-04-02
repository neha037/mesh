---
layout: default
title: Browser Extension
nav_order: 3
---

# Browser Extension

The Mesh browser extension lets you save any web page to your knowledge base with a single click.

## Installing the Extension

1. Open Chrome and navigate to `chrome://extensions`
2. Enable **Developer mode** (toggle in the top-right corner)
3. Click **Load unpacked**
4. Select the `extension/` folder inside your Mesh repository

The Mesh icon will appear in your browser toolbar.

## Saving a Page

Click the Mesh icon on any web page. The page is **saved automatically** -- no second click needed.

The popup will show:
- **"Saving..."** while the request is in progress
- **"Saved!"** (green) when the page is saved for the first time
- **"Updated!"** (amber) when re-saving a page you've already saved -- the title and content are updated, but no duplicate is created

The extension extracts the page's URL, title, and main text content (from `<article>`, `<main>`, or `<body>`).

## Viewing Recent Saves

The popup shows your 10 most recent saves below the status message. Each entry shows the title and the date it was saved.

## View All Saved Pages

Click **"View All Saved Pages"** at the bottom of the popup. This opens a full-page view in a new tab with:

- A table of all your saved pages (title, source domain, date)
- A **Delete** button on each row to remove pages you no longer want
- A **Load More** button at the bottom to load additional pages

## Deleting a Page

In the "View All Saved Pages" view, click the **Delete** button next to any page. The page is permanently removed from your knowledge base.

## Settings

Click **"Settings"** at the bottom of the popup to configure the API URL. The default is `http://localhost:8080`. You only need to change this if you're running the Mesh API on a different host or port.

## Tips

- You can click the extension icon on the same page multiple times without creating duplicates -- it will update the existing entry
- The extension works on most web pages but cannot save content from `chrome://` pages or other browser-internal URLs
- If the API is not running, the extension will show an error message -- start the Mesh service and try again
