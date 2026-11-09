import { createSignal, onMount, onCleanup } from "solid-js";
import browser from "webextension-polyfill";
import { PORT } from "./config";

function App() {
  const [serviceEnabled, setServiceEnabled] = createSignal<boolean>(false);

  onMount(async () => {
    const data = await browser.storage.local.get({ serviceEnabled: true });
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
    <div class="w-64 bg-background text-foreground p-4 font-sans">
      <header class="flex items-center justify-between pb-4 border-b border-border">
        <h1 class="text-lg font-bold">Ariadm</h1>
        <div
          class={`flex items-center text-xs font-semibold px-2 py-1 rounded-full ${
            serviceEnabled()
              ? "bg-success text-success-foreground"
              : "bg-destructive text-destructive-foreground"
          }`}
        >
          <div
            class={`w-2 h-2 rounded-full mr-1.5 ${
              serviceEnabled() ? "bg-success-foreground" : "bg-destructive-foreground"
            }`}
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
            class={`relative inline-flex items-center h-6 rounded-full w-11 transition-colors ${
              serviceEnabled() ? "bg-primary" : "bg-muted"
            }`}
          >
            <span
              class={`inline-block w-4 h-4 transform bg-card rounded-full transition-transform ${
                serviceEnabled() ? "translate-x-6" : "translate-x-1"
              }`}
            />
          </button>
        </div>
      </main>

      <footer class="text-center text-xs text-muted-foreground pt-4 border-t border-border">
        Listening on port {PORT}
      </footer>
    </div>
  );
}

export default App;
