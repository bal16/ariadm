import { createResource, createSignal, Show, For } from "solid-js";
import {
  GetApplicationConfig,
  OpenPath,
  SaveApplicationConfig,
  SelectDownloadFolder,
} from "~/../bindings/github.com/bal16/ariadm/internal/ingress/wailsbridge/wailsbridge";
import { Download, FolderOpen, Zap, Settings, X, Save } from "lucide-solid";

type TabType = "download" | "speed" | "general";

export function Preferences(props: { onClose: () => void }) {
  const [activeTab, setActiveTab] = createSignal<TabType>("download");
  const [appConfig, { mutate }] = createResource(GetApplicationConfig);
  const [isSaving, setIsSaving] = createSignal(false);
  const [isBrowsing, setIsBrowsing] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal("");

  const handleBrowseFolder = async () => {
    setIsBrowsing(true);
    setErrorMsg("");
    try {
      const path = await SelectDownloadFolder();
      if (path) {
        mutate({ ...appConfig()!, default_download_path: path });
      }
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setIsBrowsing(false);
    }
  };

  const handleOpenFolder = () => {
    const path = appConfig()?.default_download_path;
    if (path) void OpenPath(path);
  };

  // Storing component references directly in the structural map array
  const tabs = [
    { id: "download" as TabType, label: "Downloads", icon: Download },
    { id: "speed" as TabType, label: "Speed", icon: Zap },
    { id: "general" as TabType, label: "General", icon: Settings },
  ];

  const handleSave = async (e: Event) => {
    e.preventDefault();
    const currentConfig = appConfig();
    if (!currentConfig) return;

    setIsSaving(true);
    setErrorMsg("");

    try {
      await SaveApplicationConfig(currentConfig);
      props.onClose();
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 animate-in fade-in duration-150"
      onClick={props.onClose}
    >
      <div
        class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-140 h-90 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Window Title Bar */}
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <div class="flex items-center space-x-1.5 text-xs font-medium text-foreground">
            <Settings class="h-3.5 w-3.5 text-muted-foreground" />
            <span>Preferences</span>
          </div>
          <div class="flex items-center space-x-1">
            <button
              type="button"
              onClick={props.onClose}
              class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-hidden"
            >
              <X class="h-3 w-3" />
            </button>
          </div>
        </div>

        {/* Dialog Body Matrix */}
        <div class="flex flex-1 overflow-hidden bg-background">
          {/* Left Tab Sidepane Navigation */}
          <div class="w-36 border-r border-border bg-muted/20 p-1 flex flex-col space-y-0.5">
            <For each={tabs}>
              {(tab) => {
                const IconComponent = tab.icon; // Instantiating the reference variable for Solid's compiler
                return (
                  <button
                    type="button"
                    onClick={() => setActiveTab(tab.id)}
                    class={`flex items-center space-x-2 px-2.5 py-1.5 text-left text-xs rounded-sm font-medium select-none border transition-all ${
                      activeTab() === tab.id
                        ? "bg-primary text-primary-foreground border-primary shadow-xs"
                        : "text-muted-foreground border-transparent hover:bg-muted/60 hover:text-foreground"
                    }`}
                  >
                    <IconComponent class="h-3.5 w-3.5" />
                    <span>{tab.label}</span>
                  </button>
                );
              }}
            </For>
          </div>

          {/* Right Core Property Configurations */}
          <form
            onSubmit={handleSave}
            class="flex-1 flex flex-col justify-between p-3.5 overflow-y-auto"
          >
            <Show
              fallback={
                <div class="text-xs font-mono text-muted-foreground animate-pulse p-2">
                  Reading properties...
                </div>
              }
              when={!appConfig.loading}
            >
              <div class="space-y-3.5">
                <Show when={errorMsg()}>
                  <div class="p-2 bg-destructive/10 border border-destructive/20 text-destructive text-[11px] font-mono rounded-sm">
                    Error: {errorMsg()}
                  </div>
                </Show>

                {/* DOWNLOADS TAB */}
                <Show when={activeTab() === "download"}>
                  <div class="space-y-3 animate-in fade-in duration-100">
                    <h3 class="text-[11px] font-bold uppercase tracking-wider text-muted-foreground border-b border-border pb-1">
                      Queue Configuration
                    </h3>
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Default download folder
                      </label>
                      <div class="flex space-x-2">
                        <input
                          type="text"
                          class="flex-1 px-2 py-1 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                          value={appConfig()?.default_download_path || ""}
                          onInput={(e) =>
                            mutate({
                              ...appConfig()!,
                              default_download_path: e.currentTarget.value,
                            })
                          }
                        />
                        <button
                          type="button"
                          onClick={() => void handleBrowseFolder()}
                          disabled={isBrowsing()}
                          class="px-2.5 py-1 text-xs border border-input bg-muted/40 text-foreground rounded-sm hover:bg-muted disabled:opacity-50 flex items-center space-x-1"
                        >
                          <FolderOpen class="h-3.5 w-3.5" />
                          <span>{isBrowsing() ? "..." : "Browse"}</span>
                        </button>
                        <button
                          type="button"
                          onClick={handleOpenFolder}
                          class="px-2.5 py-1 text-xs border border-input bg-muted/40 text-foreground rounded-sm hover:bg-muted flex items-center space-x-1"
                        >
                          <FolderOpen class="h-3.5 w-3.5" />
                          <span>Open</span>
                        </button>
                      </div>
                    </div>
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Connections per download
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="16"
                        class="w-20 px-2 py-1 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                        value={appConfig()?.connections_per_download || 4}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            connections_per_download: parseInt(e.currentTarget.value) || 1,
                          })
                        }
                      />
                    </div>
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Max concurrent downloads
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="10"
                        class="w-20 px-2 py-1 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                        value={appConfig()?.max_concurrent_tasks || 3}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            max_concurrent_tasks: parseInt(e.currentTarget.value) || 1,
                          })
                        }
                      />
                    </div>
                  </div>
                </Show>

                {/* THROTTLING TAB */}
                <Show when={activeTab() === "speed"}>
                  <div class="space-y-3 animate-in fade-in duration-100">
                    <h3 class="text-[11px] font-bold uppercase tracking-wider text-muted-foreground border-b border-border pb-1">
                      Bandwidth Allocation
                    </h3>
                    <label class="flex items-center space-x-2 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        class="rounded-sm border-input bg-background text-primary focus:ring-ring h-3.5 w-3.5"
                        checked={(appConfig()?.speed_limit || 0) === 0}
                        onChange={(e) =>
                          mutate({
                            ...appConfig()!,
                            speed_limit: e.currentTarget.checked ? 0 : 1024 * 1024,
                          })
                        }
                      />
                      <span class="text-xs font-medium text-foreground">Unlimited speed</span>
                    </label>
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Speed limit (KB/s)
                      </label>
                      <input
                        type="number"
                        min="0"
                        disabled={(appConfig()?.speed_limit || 0) === 0}
                        class="w-36 px-2 py-1 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring disabled:opacity-50"
                        value={Math.round((appConfig()?.speed_limit || 0) / 1024)}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            speed_limit: (parseInt(e.currentTarget.value) || 0) * 1024,
                          })
                        }
                      />
                    </div>
                  </div>
                </Show>

                {/* GENERAL TAB */}
                <Show when={activeTab() === "general"}>
                  <div class="space-y-3 animate-in fade-in duration-100">
                    <h3 class="text-[11px] font-bold uppercase tracking-wider text-muted-foreground border-b border-border pb-1">
                      Desktop Options
                    </h3>
                    <label class="flex items-center space-x-2 mt-1 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        class="rounded-sm border-input bg-background text-primary focus:ring-ring h-3.5 w-3.5"
                        checked={appConfig()?.minimize_to_tray || false}
                        onChange={(e) =>
                          mutate({
                            ...appConfig()!,
                            minimize_to_tray: e.currentTarget.checked,
                          })
                        }
                      />
                      <span class="text-xs font-medium text-foreground">
                        Minimize to tray when closing the window
                      </span>
                    </label>

                    <label class="flex items-center space-x-2 mt-1 cursor-pointer select-none">
                      <input
                        type="checkbox"
                        class="rounded-sm border-input bg-background text-primary focus:ring-ring h-3.5 w-3.5"
                        checked={appConfig()?.show_in_app_menu ?? true}
                        onChange={(e) =>
                          mutate({
                            ...appConfig()!,
                            show_in_app_menu: e.currentTarget.checked,
                          })
                        }
                      />
                      <span class="text-xs font-medium text-foreground">Show in-app menu bar</span>
                    </label>
                    <p class="text-[10px] text-muted-foreground pl-6">
                      Native menu stays enabled for global menu / DBus.
                    </p>
                  </div>
                </Show>
              </div>

              {/* Window Footer Control Bar */}
              <div class="flex justify-end space-x-1.5 pt-2 border-t border-border bg-background">
                <button
                  type="button"
                  onClick={props.onClose}
                  class="px-2.5 py-1 text-xs border border-border bg-muted/40 text-muted-foreground rounded-sm hover:bg-muted hover:text-foreground transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={isSaving()}
                  class="px-3 py-1 text-xs bg-primary text-primary-foreground font-semibold rounded-sm hover:bg-primary/90 disabled:opacity-50 shadow-xs flex items-center space-x-1 transition-colors"
                >
                  <Save class="h-3 w-3" />
                  <span>{isSaving() ? "Applying..." : "Apply"}</span>
                </button>
              </div>
            </Show>
          </form>
        </div>
      </div>
    </div>
  );
}
