import browser from "webextension-polyfill";
import { downloadUrl } from "./config";
import { clearCachedPort, resolvePort } from "./discovery";

browser.runtime.onInstalled.addListener(() => {
  // Set the service to be enabled by default on installation.
  browser.storage.local.set({ serviceEnabled: true });
  console.log("Installed Successfully.");
});

async function markUnreachable(): Promise<void> {
  await browser.action.setBadgeBackgroundColor({ color: "#dc2626" });
  await browser.action.setBadgeText({ text: "!" });
}

async function markReachable(): Promise<void> {
  await browser.action.setBadgeText({ text: "" });
}

browser.downloads.onCreated.addListener(async (downloadItem) => {
  // Get the service status, defaulting to `true` if not set.
  const data = await browser.storage.local.get({ serviceEnabled: true });
  // Handle legacy `1` value as well as modern boolean.
  const serviceEnabled = data.serviceEnabled === true || data.serviceEnabled === 1;

  if (!serviceEnabled) {
    return;
  }

  // It's possible the download is initiated by the extension itself or is not a regular http/https download.
  if (!downloadItem.url || !downloadItem.id) {
    return;
  }

  try {
    // Locate the local service before touching the browser download, so an
    // unreachable app leaves the download to the browser instead of losing it.
    const port = await resolvePort();
    if (port === null) {
      await markUnreachable();
      return;
    }

    // Cancel and erase the download from the browser's history.
    await browser.downloads.cancel(downloadItem.id);
    await browser.downloads.erase({ id: downloadItem.id });

    const response = await fetch(downloadUrl(port), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ url: downloadItem.url }),
    });

    if (response.ok) {
      await markReachable();
      return;
    }

    // A cached port can go stale between resolution and delivery.
    await clearCachedPort();
    await markUnreachable();
  } catch (error) {
    console.error("Ariadm: Failed to intercept download:", error);
    await markUnreachable();
  }
});
