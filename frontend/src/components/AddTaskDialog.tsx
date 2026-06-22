import { createSignal, Show } from "solid-js";
import { Link, X, Download, TriangleAlert } from "lucide-solid";
import { TriggerNewDownload } from "~/../wailsjs/go/wailsbridge/WailsBridge";

export function AddTaskDialog(props: { onClose: () => void }) {
  const [url, setUrl] = createSignal("");
  const [isSubmitting, setIsSubmitting] = createSignal(false);
  const [errorMsg, setErrorMsg] = createSignal("");

  const handleSubmit = async (e: Event) => {
    e.preventDefault();
    if (!url().trim()) {
      setErrorMsg("URL string identifier cannot be empty.");
      return;
    }

    setIsSubmitting(true);
    setErrorMsg("");

    try {
      // 👈 Call your actual Go backend trigger method
      await TriggerNewDownload(url().trim());
      props.onClose();
    } catch (err) {
      setErrorMsg(err instanceof Error ? err.message : String(err));
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div class="absolute top-20 left-1/2 -translate-x-1/2 z-50 p-1 animate-in fade-in zoom-in-95 duration-150">
      <div class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-[480px] flex flex-col overflow-hidden">
        {/* Title Bar */}
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <div class="flex items-center space-x-1.5 text-xs font-medium text-foreground">
            <Link class="h-3.5 w-3.5 text-primary" />
            <span>Add New Download Task</span>
          </div>
          <button
            type="button"
            onClick={props.onClose}
            class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-none"
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
              Source Resource URL (HTTP / HTTPS)
            </label>
            <input
              type="text"
              autofocus
              placeholder="Paste download path link address here..."
              class="w-full px-3 py-1.5 text-xs border border-input bg-background text-foreground rounded-sm font-mono focus:outline-none focus:ring-1 focus:ring-ring"
              value={url()}
              onInput={(e) => setUrl(e.currentTarget.value)}
              disabled={isSubmitting()}
            />
          </div>

          <div class="text-[10px] text-muted-foreground font-mono bg-muted/30 border border-border p-2 rounded-sm select-none">
            ℹ️ Links are validated instantly against backend structural regex
            patterns before routing down to the active aria2 execution daemon
            stream.
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
              class="px-3 py-1 text-xs bg-primary text-primary-foreground font-semibold rounded-sm hover:bg-primary/90 disabled:opacity-50 shadow-sm flex items-center space-x-1 transition-colors"
            >
              <Download class="h-3 w-3" />
              <span>{isSubmitting() ? "Validating..." : "Launch"}</span>
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
