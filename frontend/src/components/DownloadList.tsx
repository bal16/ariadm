import { For, Show } from "solid-js";
import {
  Play,
  Pause,
  Trash2,
  File,
  CircleCheck,
  Download,
  CircleAlert,
} from "lucide-solid";
import { task } from "~/../wailsjs/go/models";

// Temporary mock helper to simulate data before our live streaming loop in Phase 7 is active
const formatBytes = (bytes: number) => {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB", "TB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

interface DownloadListProps {
  tasks: task.Task[];
  onTogglePause: (id: string) => void;
  onDelete: (id: string) => void;
}

export function DownloadList(props: DownloadListProps) {
  return (
    <div class="w-full flex flex-col h-full bg-background border border-border rounded-md overflow-hidden">
      {/* Table Main Container Frame */}
      <div class="flex-1 overflow-auto text-xs">
        <table class="w-full border-collapse text-left select-none">
          {/* Table Sticky Headers */}
          <thead class="sticky top-0 bg-muted/60 backdrop-blur-sm text-muted-foreground border-b border-border font-medium select-none z-10">
            <tr>
              <th class="p-2 w-7 text-center"></th>
              <th class="p-2 font-medium">File Name</th>
              <th class="p-2 font-medium w-24">Size</th>
              <th class="p-2 font-medium w-24">Done</th>
              <th class="p-2 font-medium w-40">Progress</th>
              <th class="p-2 font-medium w-24">Status</th>
              <th class="p-2 font-medium w-24">Speed</th>
              <th class="p-2 font-medium w-24">Actions</th>
            </tr>
          </thead>

          {/* Table Data Matrix Rows */}
          <tbody class="divide-y divide-border/60 bg-background font-mono text-[11px]">
            <Show
              when={props.tasks && props.tasks.length > 0}
              fallback={
                <tr>
                  <td
                    colspan="8"
                    class="text-center py-20 text-muted-foreground font-sans text-xs"
                  >
                    <Download class="h-8 w-8 mx-auto text-muted-foreground/30 mb-2" />
                    No active transfers found in the local session queue.
                  </td>
                </tr>
              }
            >
              <For each={props.tasks}>
                {(item) => {
                  // Calculate raw structural ratios safely
                  const pct =
                    item.total_length > 0
                      ? Math.min(
                          100,
                          parseFloat(
                            (
                              (item.completed_length / item.total_length) *
                              100
                            ).toFixed(1),
                          ),
                        )
                      : 0;

                  return (
                    <tr class="hover:bg-muted/30 transition-colors group">
                      {/* Operational Status Icon Column */}
                      <td class="p-2 text-center align-middle">
                        <Show when={item.status === "active"}>
                          <Download class="h-3.5 w-3.5 text-blue-500 animate-pulse" />
                        </Show>
                        <Show when={item.status === "paused"}>
                          <Pause class="h-3.5 w-3.5 text-amber-500" />
                        </Show>
                        <Show when={item.status === "completed"}>
                          <CircleCheck class="h-3.5 w-3.5 text-emerald-500" />
                        </Show>
                        <Show when={item.status === "error"}>
                          <CircleAlert class="h-3.5 w-3.5 text-destructive" />
                        </Show>
                      </td>

                      {/* File Label Descriptor */}
                      <td class="p-2 font-sans text-xs text-foreground truncate max-w-xs font-medium flex items-center space-x-1.5">
                        <File class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                        <span class="truncate">
                          {item.file_name || "Unknown Resource Payload"}
                        </span>
                      </td>

                      {/* Byte Calculation Identifiers */}
                      <td class="p-2 text-muted-foreground">
                        {formatBytes(item.total_length)}
                      </td>
                      <td class="p-2 text-muted-foreground">
                        {formatBytes(item.completed_length)}
                      </td>

                      {/* Mini Native UI Progress Tracking Bar */}
                      <td class="p-2 align-middle">
                        <div class="flex items-center space-x-2">
                          <div class="flex-1 bg-muted rounded-full h-2 overflow-hidden border border-border/40 relative">
                            <div
                              class={`h-full transition-all duration-300 ${
                                item.status === "completed"
                                  ? "bg-emerald-500"
                                  : item.status === "error"
                                    ? "bg-destructive"
                                    : item.status === "paused"
                                      ? "bg-amber-500/70"
                                      : "bg-primary"
                              }`}
                              style={{ width: `${pct}%` }}
                            ></div>
                          </div>
                          <span class="w-10 text-right font-semibold text-foreground text-[10px]">
                            {pct}%
                          </span>
                        </div>
                      </td>

                      {/* Human-Readable Status Tags */}
                      <td class="p-2 uppercase text-[10px] tracking-wider font-semibold">
                        <span
                          class={
                            item.status === "active"
                              ? "text-blue-500"
                              : item.status === "paused"
                                ? "text-amber-500"
                                : item.status === "completed"
                                  ? "text-emerald-500"
                                  : "text-destructive"
                          }
                        >
                          {item.status}
                        </span>
                      </td>

                      {/* Bandwidth Transmission Array metrics */}
                      <td class="p-2 text-muted-foreground font-medium">
                        {item.status === "active"
                          ? `${formatBytes(item.speed)}/s`
                          : "0 B/s"}
                      </td>

                      {/* Action Command Toolbar Operations */}
                      <td class="p-2 align-middle">
                        <div class="flex items-center space-x-1 opacity-80 group-hover:opacity-100 transition-opacity">
                          <Show
                            when={
                              item.status === "active" ||
                              item.status === "paused"
                            }
                          >
                            <button
                              type="button"
                              onClick={() => props.onTogglePause(item.id)}
                              title={
                                item.status === "active"
                                  ? "Pause Task"
                                  : "Resume Task"
                              }
                              class="p-1 rounded border border-border bg-background hover:bg-muted text-foreground transition-all active:scale-95"
                            >
                              <Show
                                when={item.status === "active"}
                                fallback={
                                  <Play class="h-3 w-3 text-emerald-500 fill-emerald-500/20" />
                                }
                              >
                                <Pause class="h-3 w-3 text-amber-500 fill-amber-500/20" />
                              </Show>
                            </button>
                          </Show>
                          <button
                            type="button"
                            onClick={() => props.onDelete(item.id)}
                            title="Remove Task"
                            class="p-1 rounded border border-border bg-background hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-all active:scale-95"
                          >
                            <Trash2 class="h-3 w-3" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                }}
              </For>
            </Show>
          </tbody>
        </table>
      </div>
    </div>
  );
}
