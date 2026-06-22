import { createSignal, onMount, onCleanup, Show } from "solid-js";
import {
  Plus,
  Pause,
  Play,
  Search,
  Settings,
  Activity,
  Download,
} from "lucide-solid";
import { Preferences } from "~/components/Preferences";
import { AddTaskDialog } from "~/components/AddTaskDialog";
import { DownloadList } from "~/components/DownloadList";
import { Button } from "~/components/ui/button";
import { task } from "~/../wailsjs/go/models";
import { EventsOn, EventsOff } from "~/../wailsjs/runtime/runtime";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuShortcut,
} from "~/components/ui/dropdown-menu";
import {
  ToggleTaskPauseState,
  GetTasks,
  DeleteTask,
} from "~/../wailsjs/go/wailsbridge/WailsBridge";

export default function App() {
  const [showPrefs, setShowPrefs] = createSignal(false);
  const [showAddTask, setShowAddTask] = createSignal(false);
  const [engineStatus, setEngineStatus] = createSignal("Connecting");
  const [tasks, setTasks] = createSignal<task.Task[]>([]);

  // Calculate live cumulative download speeds from active tasks
  const totalDownloadSpeed = () => {
    return tasks().reduce(
      (acc, t) => acc + (t.status === "active" ? t.speed : 0),
      0,
    );
  };

  // Convert raw speed numbers into scannable text updates
  const formatGlobalSpeed = (bytesPerSec: number) => {
    if (bytesPerSec === 0) return "0.00 B/s";
    const k = 1024;
    const sizes = ["B/s", "KB/s", "MB/s", "GB/s"];
    const i = Math.floor(Math.log(bytesPerSec) / Math.log(k));
    return (
      parseFloat((bytesPerSec / Math.pow(k, i)).toFixed(2)) + " " + sizes[i]
    );
  };

  onMount(async () => {
    // Pull phase: Fetch the task data immediately on startup
    try {
      const initialTasks = await GetTasks();
      if (initialTasks) {
        setTasks(initialTasks.map((t) => task.Task.createFrom(t)));
        setEngineStatus("Running");
      }
    } catch (err) {
      console.error("Failed to load initial task records:", err);
      setEngineStatus("Disconnected");
    }

    // Push phase: Keep listening for the background ticker loop updates
    EventsOn("tasks:update", (incomingData: any[]) => {
      if (!incomingData) {
        setTasks([]);
        return;
      }
      const parsedTasks = incomingData.map((t) => task.Task.createFrom(t));
      setTasks(parsedTasks);
    });

    EventsOn("engine:status", (status: string) => {
      setEngineStatus(status === "running" ? "Running" : "Disconnected");
    });
  });

  onCleanup(() => {
    // Clean up event stream listeners to prevent memory issues during live reload
    EventsOff("tasks:update");
    EventsOff("engine:status");
  });

  // Call your Go service layer rather than modifying mock arrays locally
  const handleTogglePause = async (id: string) => {
    try {
      await ToggleTaskPauseState(id);
    } catch (err) {
      console.error("Failed to transition task status:", err);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await DeleteTask(id);
      // Optimistic removal from local state after backend confirms
      setTasks(tasks().filter((t) => t.id !== id));
    } catch (err) {
      console.error("Failed to delete task:", err);
    }
  };

  return (
    <div class="h-screen w-screen flex flex-col overflow-hidden relative select-none bg-background text-foreground font-sans antialiased">
      {/* 1. Desktop Window Frame Menu Layout */}
      <div class="flex items-center px-2 py-0.5 bg-muted/40 border-b border-border text-xs space-x-1 z-40">
        <span class="font-bold text-primary select-none font-mono px-2 mr-2 flex items-center space-x-1">
          <Activity class="h-3.5 w-3.5 text-primary" />
          <span>AriaDM</span>
        </span>

        <DropdownMenu>
          <DropdownMenuTrigger
            as="button"
            class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-none"
          >
            File
          </DropdownMenuTrigger>
          <DropdownMenuContent class="w-44 bg-popover border border-border text-popover-foreground shadow-md">
            <DropdownMenuItem
              onClick={() => setShowAddTask(true)}
              class="text-xs font-medium cursor-pointer flex items-center space-x-2"
            >
              <Plus class="h-3.5 w-3.5 text-muted-foreground" />
              <span class="flex-1">New Task</span>
              <DropdownMenuShortcut class="font-mono text-[10px]">
                Ctrl+N
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <DropdownMenu>
          <DropdownMenuTrigger
            as="button"
            class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-none"
          >
            Tools
          </DropdownMenuTrigger>
          <DropdownMenuContent class="w-48 bg-popover border border-border text-popover-foreground shadow-md">
            <DropdownMenuItem
              onClick={() => setShowPrefs(true)}
              class="text-xs font-medium cursor-pointer flex items-center space-x-2"
            >
              <Settings class="h-3.5 w-3.5 text-muted-foreground" />
              <span class="flex-1">Preferences</span>
              <DropdownMenuShortcut class="font-mono text-[10px]">
                Ctrl+P
              </DropdownMenuShortcut>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>

        <span class="text-muted-foreground/40 px-2 text-[11px] hover:text-foreground cursor-pointer transition-colors">
          Help
        </span>
      </div>

      {/* 2. Top-Level Main Action Toolbar */}
      <div class="flex items-center justify-between p-2 border-b border-border bg-muted/10 z-30">
        <div class="flex items-center space-x-1.5">
          <Button
            onClick={() => setShowAddTask(true)}
            size="sm"
            class="h-7 px-2.5 text-xs bg-primary text-primary-foreground font-semibold rounded shadow-sm hover:opacity-90 flex items-center space-x-1"
          >
            <Plus class="h-3.5 w-3.5" />
            <span>New Task</span>
          </Button>
          <Button
            size="sm"
            variant="outline"
            class="h-7 px-2.5 text-xs border border-border bg-background text-foreground rounded hover:bg-muted flex items-center space-x-1"
          >
            <Pause class="h-3.5 w-3.5 text-muted-foreground" />
            <span>Pause All</span>
          </Button>
          <Button
            size="sm"
            variant="outline"
            class="h-7 px-2.5 text-xs border border-border bg-background text-foreground rounded hover:bg-muted flex items-center space-x-1"
          >
            <Play class="h-3.5 w-3.5 text-muted-foreground" />
            <span>Resume All</span>
          </Button>
        </div>
        <div class="relative">
          <Search class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/80" />
          <input
            type="text"
            placeholder="Search tasks..."
            class="pl-7 pr-2 py-1 h-7 text-xs border border-input bg-background text-foreground rounded font-mono w-48 focus:outline-none focus:ring-1 focus:ring-ring"
          />
        </div>
      </div>

      {/* 3. Central Application Queue Workspace Area */}
      <div class="flex-1 p-2 bg-background/50 overflow-hidden">
        <DownloadList
          tasks={tasks()}
          onTogglePause={handleTogglePause}
          onDelete={handleDelete}
        />
      </div>

      {/* 4. Desktop System Status Footer Pin */}
      <div class="bg-muted/60 text-muted-foreground border-t border-border px-3 py-1 text-xs flex items-center justify-between font-mono select-none z-30">
        <div class="flex items-center space-x-4">
          <div class="flex items-center space-x-1.5">
            <span
              class={`h-2 w-2 rounded-full ${engineStatus() === "Running" ? "bg-emerald-500 animate-pulse" : "bg-destructive"}`}
            ></span>
            <span class="text-foreground font-medium">
              Engine: {engineStatus()}
            </span>
          </div>
          <span class="text-muted-foreground/40">│</span>
          <span>Session: 127.0.0.1:6800</span>
        </div>
        <div class="flex items-center space-x-4">
          <span class="text-foreground flex items-center space-x-1">
            <Download class="h-3 w-3 rotate-180 text-muted-foreground/70" />
            {/* 👇 Live reactive download speed tracker accumulation */}
            <span>⬇️ {formatGlobalSpeed(totalDownloadSpeed())}</span>
          </span>
          <span class="text-foreground flex items-center space-x-1">
            <Download class="h-3 w-3 text-muted-foreground/70" />
            <span>⬆️ 0.00 B/s</span>
          </span>
        </div>
      </div>

      {/* Modal Layers */}
      <Show when={showPrefs()}>
        <Preferences onClose={() => setShowPrefs(false)} />
      </Show>

      <Show when={showAddTask()}>
        <AddTaskDialog onClose={() => setShowAddTask(false)} />
      </Show>
    </div>
  );
}
