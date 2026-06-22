import { CircleStop, MonitorOff, Minimize2 } from "lucide-solid";

export function QuitConfirmDialog(props: {
  onClose: () => void;
  onQuit: () => void;
  onBackground: () => void;
}) {
  return (
    <div class="absolute top-20 left-1/2 -translate-x-1/2 z-[60] p-1 animate-in fade-in zoom-in-95 duration-150">
      <div class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-[420px] flex flex-col overflow-hidden">
        {/* Title Bar */}
        <div class="flex items-center px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <div class="flex items-center space-x-1.5 text-xs font-medium text-foreground">
            <MonitorOff class="h-3.5 w-3.5 text-primary" />
            <span>Close Application</span>
          </div>
        </div>

        {/* Body */}
        <div class="p-4 bg-background space-y-4">
          <p class="text-sm text-foreground font-medium">
            How would you like to close AriaDM?
          </p>
          
          <div class="grid grid-cols-2 gap-2 mt-2">
            <button
              onClick={props.onBackground}
              class="flex flex-col items-center justify-center space-y-2 p-3 border border-border rounded-md bg-muted/20 hover:bg-muted/60 hover:border-primary/50 transition-colors group"
            >
              <Minimize2 class="h-6 w-6 text-muted-foreground group-hover:text-primary transition-colors" />
              <span class="text-xs font-semibold">Run in Background</span>
              <span class="text-[10px] text-muted-foreground text-center">
                Keep downloading in the background. Open the app shortcut again to show this window.
              </span>
            </button>
            <button
              onClick={props.onQuit}
              class="flex flex-col items-center justify-center space-y-2 p-3 border border-border rounded-md bg-muted/20 hover:bg-destructive/10 hover:border-destructive/50 transition-colors group"
            >
              <CircleStop class="h-6 w-6 text-muted-foreground group-hover:text-destructive transition-colors" />
              <span class="text-xs font-semibold">Quit Completely</span>
              <span class="text-[10px] text-muted-foreground text-center">
                Stop the active download engine and close the app entirely.
              </span>
            </button>
          </div>

          <div class="flex justify-end pt-2 border-t border-border mt-4">
            <button
              type="button"
              onClick={props.onClose}
              class="px-4 py-1 text-xs border border-border bg-muted/40 text-muted-foreground rounded-sm hover:bg-muted hover:text-foreground transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
