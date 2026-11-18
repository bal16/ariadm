import { X } from "lucide-solid";

export function AboutDialog(props: { onClose: () => void }) {
  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 animate-in fade-in duration-150"
      onClick={props.onClose}
    >
      <div
        class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-80 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <span class="text-xs font-medium text-foreground">About AriaDM</span>
          <button
            type="button"
            onClick={props.onClose}
            class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-hidden"
          >
            <X class="h-3 w-3" />
          </button>
        </div>
        <div class="p-4 bg-background space-y-2 text-xs text-foreground">
          <p class="font-semibold">AriaDM</p>
          <p class="text-muted-foreground">Desktop download manager for aria2.</p>
          <p class="text-muted-foreground">Built with Wails v3 and SolidJS.</p>
        </div>
      </div>
    </div>
  );
}
