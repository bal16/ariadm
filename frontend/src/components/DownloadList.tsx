import { createSignal, For, Show } from "solid-js";
import { Play, Pause, Trash2, File, CircleCheck, Download, CircleAlert, Clock } from "lucide-solid";
import type { Task } from "~/../bindings/github.com/bal16/ariadm/internal/domain/task/models";
import { formatBytes, formatEta } from "~/lib/format";

interface DownloadListProps {
  tasks: Task[];
  onTogglePause: (id: string) => void;
  onDelete: (id: string) => void;
  onNewTask?: () => void;
  searchQuery?: string;
}

type FilterKey = "all" | "downloading" | "completed" | "paused" | "failed";

const FILTERS: { key: FilterKey; label: string }[] = [
  { key: "all", label: "All" },
  { key: "downloading", label: "Downloading" },
  { key: "paused", label: "Paused" },
  { key: "completed", label: "Completed" },
  { key: "failed", label: "Failed" },
];

function statusLabel(status: string): string {
  switch (status) {
    case "active":
      return "Downloading";
    case "waiting":
      return "Queued";
    case "paused":
      return "Paused";
    case "complete":
      return "Completed";
    case "error":
      return "Failed";
    default:
      return status.charAt(0).toUpperCase() + status.slice(1);
  }
}

function statusClass(status: string): string {
  switch (status) {
    case "active":
    case "waiting":
      return "text-blue-500";
    case "paused":
      return "text-amber-500";
    case "complete":
      return "text-emerald-500";
    case "error":
      return "text-destructive";
    default:
      return "text-muted-foreground";
  }
}

function matchesFilter(task: Task, filter: FilterKey): boolean {
  switch (filter) {
    case "all":
      return true;
    case "downloading":
      return task.status === "active" || task.status === "waiting";
    case "paused":
      return task.status === "paused";
    case "completed":
      return task.status === "complete";
    case "failed":
      return task.status === "error";
    default:
      return true;
  }
}

export function DownloadList(props: DownloadListProps) {
  const [activeFilter, setActiveFilter] = createSignal<FilterKey>("all");

  const filteredTasks = () => {
    const query = (props.searchQuery ?? "").trim().toLowerCase();
    return props.tasks.filter((task) => {
      if (!matchesFilter(task, activeFilter())) return false;
      if (!query) return true;
      return (
        (task.file_name ?? "").toLowerCase().includes(query) ||
        (task.url ?? "").toLowerCase().includes(query)
      );
    });
  };

  const etaFor = (task: Task) => {
    const remaining = Math.max(0, task.total_length - task.completed_length);
    if (remaining === 0 || task.speed <= 0) return 0;
    return remaining / task.speed;
  };

  return (
    <div class="w-full flex flex-col h-full bg-background border border-border rounded-md overflow-hidden">
      {/* Category tabs */}
      <div class="flex items-center gap-1 px-2 py-1.5 border-b border-border bg-muted/20">
        <For each={FILTERS}>
          {(filter) => (
            <button
              type="button"
              onClick={() => setActiveFilter(filter.key)}
              class={`px-2.5 py-1 text-xs rounded-sm font-medium transition-colors ${
                activeFilter() === filter.key
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:bg-muted/60 hover:text-foreground"
              }`}
            >
              {filter.label}
            </button>
          )}
        </For>
      </div>

      <div class="flex-1 overflow-auto text-xs">
        <table class="w-full border-collapse text-left select-none">
          <thead class="sticky top-0 bg-muted/60 backdrop-blur-xs text-muted-foreground border-b border-border font-medium select-none z-10">
            <tr>
              <th class="p-2 w-7 text-center"></th>
              <th class="p-2 font-medium">Name</th>
              <th class="p-2 font-medium w-24">Size</th>
              <th class="p-2 font-medium w-24">Downloaded</th>
              <th class="p-2 font-medium w-44">Progress</th>
              <th class="p-2 font-medium w-24">Status</th>
              <th class="p-2 font-medium w-20">Speed</th>
              <th class="p-2 font-medium w-20">Time left</th>
              <th class="p-2 font-medium w-24">Actions</th>
            </tr>
          </thead>

          <tbody class="divide-y divide-border/60 bg-background font-mono text-[11px]">
            <For
              each={filteredTasks()}
              fallback={
                <tr>
                  <td colspan="9" class="text-center py-16">
                    <Download class="h-8 w-8 mx-auto text-muted-foreground/30 mb-2" />
                    <p class="text-muted-foreground font-sans text-xs">
                      {props.tasks.length === 0
                        ? "No downloads yet."
                        : "No downloads match this filter."}
                    </p>
                    <Show when={props.tasks.length === 0 && props.onNewTask}>
                      <button
                        type="button"
                        onClick={() => props.onNewTask?.()}
                        class="mt-3 px-3 py-1 text-xs bg-primary text-primary-foreground rounded-sm font-semibold hover:opacity-90"
                      >
                        Add download
                      </button>
                    </Show>
                  </td>
                </tr>
              }
            >
              {(item) => {
                const pct = () =>
                  item.total_length > 0
                    ? Math.min(
                        100,
                        parseFloat(((item.completed_length / item.total_length) * 100).toFixed(1)),
                      )
                    : 0;

                return (
                  <tr class="hover:bg-muted/30 transition-colors group">
                    <td class="p-2 text-center align-middle">
                      <span class="flex items-center justify-center">
                        {item.status === "active" && (
                          <Download class="h-3.5 w-3.5 text-blue-500 animate-pulse" />
                        )}
                        {item.status === "waiting" && <Clock class="h-3.5 w-3.5 text-blue-500" />}
                        {item.status === "paused" && <Pause class="h-3.5 w-3.5 text-amber-500" />}
                        {item.status === "complete" && (
                          <CircleCheck class="h-3.5 w-3.5 text-emerald-500" />
                        )}
                        {item.status === "error" && (
                          <CircleAlert class="h-3.5 w-3.5 text-destructive" />
                        )}
                      </span>
                    </td>

                    <td class="p-2 font-sans text-xs text-foreground truncate max-w-xs font-medium flex items-center space-x-1.5">
                      <File class="h-3.5 w-3.5 text-muted-foreground shrink-0" />
                      <span class="truncate" title={item.file_name || item.url}>
                        {item.file_name || "Unknown file"}
                      </span>
                    </td>

                    <td class="p-2 text-muted-foreground">{formatBytes(item.total_length)}</td>
                    <td class="p-2 text-muted-foreground">{formatBytes(item.completed_length)}</td>

                    <td class="p-2 align-middle">
                      <div class="flex items-center space-x-2">
                        <div class="flex-1 bg-muted rounded-full h-2 overflow-hidden border border-border/40 relative">
                          <div
                            class={`h-full transition-all duration-300 ${
                              item.status === "complete"
                                ? "bg-emerald-500"
                                : item.status === "error"
                                  ? "bg-destructive"
                                  : item.status === "paused"
                                    ? "bg-amber-500/70"
                                    : "bg-primary"
                            }`}
                            style={{ width: `${pct()}%` }}
                          ></div>
                        </div>
                        <span class="w-10 text-right font-semibold text-foreground text-[10px]">
                          {pct()}%
                        </span>
                      </div>
                    </td>

                    <td class="p-2 uppercase text-[10px] tracking-wider font-semibold">
                      <span class={statusClass(item.status)}>{statusLabel(item.status)}</span>
                    </td>

                    <td class="p-2 text-muted-foreground font-medium">
                      {item.status === "active" ? `${formatBytes(item.speed)}/s` : "—"}
                    </td>

                    <td class="p-2 text-muted-foreground font-medium">
                      {item.status === "active" || item.status === "waiting"
                        ? formatEta(etaFor(item))
                        : "—"}
                    </td>

                    <td class="p-2 align-middle">
                      <div class="flex items-center space-x-1 opacity-80 group-hover:opacity-100 transition-opacity">
                        {(item.status === "active" ||
                          item.status === "paused" ||
                          item.status === "waiting") && (
                          <button
                            type="button"
                            onClick={() => props.onTogglePause(item.id)}
                            class="p-1 rounded-sm border border-border bg-background hover:bg-muted text-foreground transition-all active:scale-95"
                            title={item.status === "paused" ? "Resume" : "Pause"}
                          >
                            {item.status === "paused" ? (
                              <Play class="h-3 w-3 text-emerald-500 fill-emerald-500/20" />
                            ) : (
                              <Pause class="h-3 w-3 text-amber-500 fill-amber-500/20" />
                            )}
                          </button>
                        )}
                        <button
                          type="button"
                          onClick={() => props.onDelete(item.id)}
                          class="p-1 rounded-sm border border-border bg-background hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-all active:scale-95"
                          title="Remove"
                        >
                          <Trash2 class="h-3 w-3" />
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              }}
            </For>
          </tbody>
        </table>
      </div>
    </div>
  );
}
