import browser from "webextension-polyfill";
import { PORT } from "./config";

console.log("Background service worker berjalan!");

browser.runtime.onInstalled.addListener(() => {
  console.log("Ekstensi berhasil diinstal.");
});

browser.downloads.onCreated.addListener((downloadItem) => {
  browser.downloads
    .cancel(downloadItem.id)
    .then(() => {
      console.log(`Download dibatalkan: ${downloadItem.filename}`);
      console.log({ url: downloadItem.url });
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
      console.error(`Gagal membatalkan download: ${error}`);
    });
});
