import { createSignal, onMount, Show } from "solid-js";
import { X, Download, TriangleAlert, FolderOpen } from "lucide-solid";
import {
  GetApplicationConfig,
  SelectDownloadFolder,
  TriggerNewDownloadWithOptions,
} from "~/../bindings/github.com/bal16/ariadm/internal/ingress/wailsbridge/wailsbridge";

export function AddTaskDialog(props: { onClose: () => void }) {
  const [url, setUrl] = createSignal("");
  const [fileName, setFileName] = createSignal("");
  const [downloadPath, setDownloadPath] = createSignal("");
  const [connections, setConnections] = createSignal(1);
  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [isBrowsing, setIsBrowsing] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal("");
  let urlInput: HTMLInputElement | undefined;

  onMount(() => {
    urlInput?.focus();
    void GetApplicationConfig()
      .then((cfg) => {
        if (cfg && cfg.connections_per_download > 0) {
          setConnections(cfg.connections_per_download);
        }
      })
      .catch(() => {});
  });

  const handleBrowse = async () => {
    setIsBrowsing(true);
    setErrorMsg("");
    try {
      const path = await SelectDownloadFolder();
      if (path) setDownloadPath(path);
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setIsBrowsing(false);
    }
  };

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!url().trim()) {
      setErrorMsg("Please enter a download URL.");
      return;
    }

    setIsSubmitting(true);
    setErrorMsg("");

    try {
      await TriggerNewDownloadWithOptions(
        url().trim(),
        downloadPath().trim(),
        fileName().trim(),
        connections(),
      );
      props.onClose();
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 animate-in fade-in duration-150"
      onClick={props.onClose}
    >
      <div
        class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-130 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Title Bar */}
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <div class="flex items-center space-x-1.5 text-xs font-medium text-foreground">
            <Download class="h-3.5 w-3.5 text-primary" />
            <span>Add Download</span>
          </div>
          <button
            type="button"
            onClick={props.onClose}
            class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-hidden"
          >
            <X class="h-3 w-3" />
          </button>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} class="p-4 bg-background space-y-4">
          <Show when={errorMsg()}>
            <div class="p-2 bg-destructive/10 border border-destructive/20 text-destructive text-[11px] font-mono rounded-sm flex items-start space-x-2">
              <TriangleAlert class="h-3.5 w-3.5 shrink-0 mt-0.5" />
              <span>{errorMsg()}</span>
            </div>
          </Show>

          <div class="space-y-1.5">
            <label class="block text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
              Download URL
            </label>
            <input
              ref={(el) => {
                urlInput = el;
              }}
              type="text"
              placeholder="Paste download link here..."
              class="w-full px-3 py-1.5 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
              value={url()}
              onInput={(e) => setUrl(e.currentTarget.value)}
              disabled={isSubmitting()}
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div class="space-y-1.5">
              <label class="block text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                File name (optional)
              </label>
              <input
                type="text"
                placeholder="auto"
                class="w-full px-3 py-1.5 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                value={fileName()}
                onInput={(e) => setFileName(e.currentTarget.value)}
                disabled={isSubmitting()}
              />
            </div>
            <div class="space-y-1.5">
              <label class="block text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                Connections
              </label>
              <input
                type="number"
                min="1"
                max="16"
                class="w-full px-3 py-1.5 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                value={connections()}
                onInput={(e) => setConnections(parseInt(e.currentTarget.value) || 1)}
                disabled={isSubmitting()}
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <label class="block text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
              Save to folder (optional)
            </label>
            <div class="flex space-x-2">
              <input
                type="text"
                placeholder="Default download folder"
                class="flex-1 px-3 py-1.5 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-hidden focus:ring-1 focus:ring-ring"
                value={downloadPath()}
                onInput={(e) => setDownloadPath(e.currentTarget.value)}
                disabled={isSubmitting()}
              />
              <button
                type="button"
                onClick={() => void handleBrowse()}
                disabled={isSubmitting() || isBrowsing()}
                class="px-2.5 py-1 text-xs border border-input bg-muted/40 text-foreground rounded-sm hover:bg-muted disabled:opacity-50 flex items-center space-x-1"
              >
                <FolderOpen class="h-3.5 w-3.5" />
                <span>{isBrowsing() ? "..." : "Browse"}</span>
              </button>
            </div>
          </div>

          <div class="text-[10px] text-muted-foreground font-mono bg-muted/30 border border-border p-2 rounded-sm select-none">
            The link is validated and sent to the download engine.
          </div>

          <div class="flex justify-end space-x-1.5 pt-2 border-t border-border">
            <button
              type="button"
              onClick={props.onClose}
              disabled={isSubmitting()}
              class="px-2.5 py-1 text-xs border border-border bg-muted/40 text-muted-foreground rounded-sm hover:bg-muted hover:text-foreground disabled:opacity-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={isSubmitting()}
              class="px-3 py-1 text-xs bg-primary text-primary-foreground font-semibold rounded-sm hover:opacity-90 disabled:opacity-50 shadow-xs flex items-center space-x-1 transition-colors"
            >
              <Download class="h-3 w-3" />
              <span>{isSubmitting() ? "Adding..." : "Download"}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
