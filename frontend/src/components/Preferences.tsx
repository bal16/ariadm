// frontend/src/components/Preferences.tsx
import { createResource, createSignal, Show, For } from "solid-js";
import {
  GetApplicationConfig,
  SaveApplicationConfig,
} from "~/../wailsjs/go/wailsbridge/WailsBridge";
import { Download, Zap, Settings, X, Save } from "lucide-solid";

type TabType = "download" | "speed" | "general";

export function Preferences(props: { onClose: () => void }) {
  const [activeTab, setActiveTab] = createSignal<TabType>("download");
  const [appConfig, { mutate }] = createResource(GetApplicationConfig);
  const [isSaving, setIsSaving] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal("");

  // Storing component references directly in the structural map array
  const tabs = [
    { id: "download" as TabType, label: "Downloads", icon: Download },
    { id: "speed" as TabType, label: "Throttling", icon: Zap },
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
    <div class="absolute top-16 left-1/2 -translate-x-1/2 z-50 p-1 animate-in fade-in zoom-in-95 duration-150">
      <div class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-[560px] h-[360px] flex flex-col overflow-hidden">
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
              class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-none"
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
                        ? "bg-primary text-primary-foreground border-primary shadow-sm"
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
                  <div class="p-2 bg-destructive/10 border border-destructive/20 text-destructive text-[11px] font-mono rounded">
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
                        Default Download Directory
                      </label>
                      <input
                        type="text"
                        class="w-full px-2 py-1 text-xs border border-input bg-background text-foreground rounded font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                        value={appConfig()?.default_download_path || ""}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            default_download_path: e.currentTarget.value,
                          })
                        }
                      />
                    </div>
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Max Concurrent Tasks
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="10"
                        class="w-20 px-2 py-1 text-xs border border-input bg-background text-foreground rounded font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                        value={appConfig()?.max_concurrent_tasks || 3}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            max_concurrent_tasks:
                              parseInt(e.currentTarget.value) || 1,
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
                    <div class="space-y-1">
                      <label class="block text-[11px] font-medium text-muted-foreground">
                        Global Speed Limit (Bytes/s, 0 = Max Speed)
                      </label>
                      <input
                        type="number"
                        class="w-36 px-2 py-1 text-xs border border-input bg-background text-foreground rounded font-mono focus:outline-none focus:ring-1 focus:ring-ring"
                        value={appConfig()?.speed_limit || 0}
                        onInput={(e) =>
                          mutate({
                            ...appConfig()!,
                            speed_limit: parseInt(e.currentTarget.value) || 0,
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
                        class="rounded border-input bg-background text-primary focus:ring-ring h-3.5 w-3.5"
                        checked={appConfig()?.minimize_to_tray || false}
                        onChange={(e) =>
                          mutate({
                            ...appConfig()!,
                            minimize_to_tray: e.currentTarget.checked,
                          })
                        }
                      />
                      <span class="text-xs font-medium text-foreground">
                        Minimize to system tray on window close
                      </span>
                    </label>
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
                  class="px-3 py-1 text-xs bg-primary text-primary-foreground font-semibold rounded-sm hover:bg-primary/90 disabled:opacity-50 shadow-sm flex items-center space-x-1 transition-colors"
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
