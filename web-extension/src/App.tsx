import { createSignal, onMount, onCleanup } from "solid-js";
import browser from "webextension-polyfill";
import { PORT } from "./config";

function App() {
  const [serviceEnabled, setServiceEnabled] = createSignal<boolean>(false);

  onMount(async () => {
    // Set a default value of `true` if `serviceEnabled` is not found in storage.
    const data = await browser.storage.local.get({ serviceEnabled: true });
    // The check for `1` is kept for backward compatibility if older versions used it.
    const isEnabled = data.serviceEnabled === 1 || data.serviceEnabled === true;
    setServiceEnabled(isEnabled);

    browser.storage.onChanged.addListener(handleStorageChange);
  });

  onCleanup(() => {
    browser.storage.onChanged.removeListener(handleStorageChange);
  });

  const handleStorageChange = (
    changes: { [key: string]: browser.Storage.StorageChange },
    areaName: string,
  ) => {
    if (areaName === "local" && changes.serviceEnabled) {
      setServiceEnabled(!!changes.serviceEnabled.newValue);
    }
  };

  const handleToggle = async () => {
    const newStatus = !serviceEnabled();
    await browser.storage.local.set({ serviceEnabled: newStatus });
    setServiceEnabled(newStatus);
  };

  return (
    <>
      <div class="w-64 bg-gray-50 dark:bg-gray-900 text-gray-800 dark:text-gray-200 p-4 font-sans">
        <header class="flex items-center justify-between pb-4 border-b border-gray-200 dark:border-gray-700">
          <h1 class="text-lg font-bold">Ariadm</h1>
          <div
            class={`flex items-center text-xs font-semibold px-2 py-1 rounded-full ${
              serviceEnabled()
                ? "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
                : "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
            }`}
          >
            <div
              class={`w-2 h-2 rounded-full mr-1.5 ${serviceEnabled() ? "bg-green-500" : "bg-red-500"}`}
            ></div>
            {serviceEnabled() ? "Active" : "Inactive"}
          </div>
        </header>

        <main class="py-4">
          <div class="flex items-center justify-between">
            <label for="service-toggle" class="text-sm">
              Enable Service
            </label>
            <button
              id="service-toggle"
              onClick={handleToggle}
              class={`relative inline-flex items-center h-6 rounded-full w-11 transition-colors ${serviceEnabled() ? "bg-blue-600" : "bg-gray-300 dark:bg-gray-600"}`}
            >
              <span
                class={`inline-block w-4 h-4 transform bg-white rounded-full transition-transform ${serviceEnabled() ? "translate-x-6" : "translate-x-1"}`}
              />
            </button>
          </div>
        </main>

        <footer class="text-center text-xs text-gray-400 dark:text-gray-500 pt-4 border-t border-gray-200 dark:border-gray-700">
          Listening on port {PORT}
        </footer>
      </div>
    </>
  );
}

export default App;
