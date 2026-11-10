import { Download } from "lucide-solid";
import { formatSpeed } from "~/lib/format";
import {
  EngineState,
  ExtensionState,
  type Status as RuntimeStatus,
} from "~/../bindings/github.com/bal16/ariadm/internal/appstate/models";

interface StatusBarProps {
  runtimeStatus: RuntimeStatus;
  downloadSpeed: number;
}

function engineLabel(state: EngineState): string {
  switch (state) {
    case EngineState.EngineStarting:
      return "Connecting";
    case EngineState.EngineRunning:
      return "Running";
    case EngineState.EngineDisconnected:
      return "Disconnected";
    case EngineState.EnginePortConflict:
      return "Port conflict";
    case EngineState.EngineFailed:
      return "Engine error";
    case EngineState.EngineStopped:
      return "Stopped";
    default:
      return "Unknown";
  }
}

function engineDotClass(state: EngineState): string {
  switch (state) {
    case EngineState.EngineRunning:
      return "bg-emerald-500 animate-pulse";
    case EngineState.EngineStarting:
      return "bg-amber-500 animate-pulse";
    default:
      return "bg-destructive";
  }
}

export function StatusBar(props: StatusBarProps) {
  const engineRunning = () => props.runtimeStatus.engine === EngineState.EngineRunning;
  const extensionListening = () =>
    props.runtimeStatus.extension === ExtensionState.ExtensionListening;

  return (
    <div class="bg-muted/60 text-muted-foreground border-t border-border px-3 py-1 text-xs flex items-center justify-between font-mono select-none z-30">
      <div class="flex items-center space-x-4">
        <div class="flex items-center space-x-1.5" title={props.runtimeStatus.engine_message}>
          <span class={`h-2 w-2 rounded-full ${engineDotClass(props.runtimeStatus.engine)}`}></span>
          <span class="text-foreground font-medium">
            Engine: {engineLabel(props.runtimeStatus.engine)}
          </span>
          {engineRunning() && (
            <span class="text-muted-foreground/70">
              127.0.0.1:{props.runtimeStatus.engine_port}
            </span>
          )}
        </div>
        <span class="text-muted-foreground/40">│</span>
        <span title={props.runtimeStatus.extension_message}>
          {extensionListening()
            ? `Extension: Listening :${props.runtimeStatus.extension_port}`
            : "Extension: Unavailable"}
        </span>
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
