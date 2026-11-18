import { For } from "solid-js";
import { X } from "lucide-solid";

const SHORTCUTS = [
  { keys: "Ctrl+N", action: "New download" },
  { keys: "Ctrl+P", action: "Preferences" },
  { keys: "Ctrl+Q", action: "Quit" },
  { keys: "Space", action: "Pause / resume selected download" },
];

export function KeyboardShortcutsDialog(props: { onClose: () => void }) {
  return (
    <div
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4 animate-in fade-in duration-150"
      onClick={props.onClose}
    >
      <div
        class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-96 flex flex-col overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <span class="text-xs font-medium text-foreground">Keyboard Shortcuts</span>
          <button
            type="button"
            onClick={props.onClose}
            class="text-muted-foreground hover:text-destructive-foreground hover:bg-destructive text-[10px] h-4 w-5 flex items-center justify-center rounded-sm transition-colors focus:outline-hidden"
          >
            <X class="h-3 w-3" />
          </button>
        </div>
        <div class="p-4 bg-background space-y-2">
          <For each={SHORTCUTS}>
            {(item) => (
              <div class="flex items-center justify-between text-xs">
                <span class="text-foreground">{item.action}</span>
                <kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-muted border border-border rounded-sm text-muted-foreground">
                  {item.keys}
                </kbd>
              </div>
            )}
          </For>
        </div>
      </div>
    </div>
  );
}
