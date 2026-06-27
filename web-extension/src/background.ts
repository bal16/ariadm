import browser from "webextension-polyfill";
import { PORT } from "./config";

browser.runtime.onInstalled.addListener(() => {
  // Set the service to be enabled by default on installation.
  browser.storage.local.set({ serviceEnabled: true });
  console.log("Installed Successfully.");
});

browser.downloads.onCreated.addListener(async (downloadItem) => {
  // Get the service status, defaulting to `true` if not set.
  const data = await browser.storage.local.get({ serviceEnabled: true });
  // Handle legacy `1` value as well as modern boolean.
  const serviceEnabled =
    data.serviceEnabled === true || data.serviceEnabled === 1;

  if (!serviceEnabled) {
    return;
  }

  // It's possible the download is initiated by the extension itself or is not a regular http/https download.
  if (!downloadItem.url || !downloadItem.id) {
    return;
  }

  try {
    // Cancel and erase the download from the browser's history.
    await browser.downloads.cancel(downloadItem.id);
    await browser.downloads.erase({ id: downloadItem.id });

    // Send the download URL to the local service.
    await fetch(`http://localhost:${PORT}/download`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: downloadItem.url }),
    });
  } catch (error) {
    console.error("Ariadm: Failed to intercept download:", error);
  }
});
