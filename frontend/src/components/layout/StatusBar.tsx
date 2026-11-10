import { Download } from "lucide-solid";
import { formatSpeed } from "~/lib/format";
import type { EngineStatus } from "~/hooks/useTaskQueue";

interface StatusBarProps {
  engineStatus: EngineStatus;
  downloadSpeed: number;
}

export function StatusBar(props: StatusBarProps) {
  return (
    <div class="bg-muted/60 text-muted-foreground border-t border-border px-3 py-1 text-xs flex items-center justify-between font-mono select-none z-30">
      <div class="flex items-center space-x-4">
        <div class="flex items-center space-x-1.5">
          <span
            class={`h-2 w-2 rounded-full ${props.engineStatus === "Running" ? "bg-emerald-500 animate-pulse" : "bg-destructive"}`}
          ></span>
          <span class="text-foreground font-medium">Engine: {props.engineStatus}</span>
        </div>
        <span class="text-muted-foreground/40">│</span>
        <span>Session: 127.0.0.1:6800</span>
      </div>
      <div class="flex items-center space-x-4">
        <span class="text-foreground flex items-center space-x-1">
          <Download class="h-3 w-3 text-muted-foreground/70" />
          <span>{formatSpeed(props.downloadSpeed)}</span>
        </span>
      </div>
    </div>
  );
}
