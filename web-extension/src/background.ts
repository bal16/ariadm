import browser from "webextension-polyfill";
import { PORT } from "./config";

console.log("Background service worker berjalan!");

browser.runtime.onInstalled.addListener(() => {
  console.log("Installed Successfully.");
});

browser.downloads.onCreated.addListener((downloadItem) => {
  browser.downloads
    .cancel(downloadItem.id)
    .then(() => {
      browser.downloads.erase({ id: downloadItem.id });

      fetch(`http://localhost:${PORT}/download`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          url: downloadItem.url,
        }),
      });
    })
    .catch((error) => {
      console.error(`Failed to interrupt download: ${error}`);
    });
});
