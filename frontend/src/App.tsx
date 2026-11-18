import { createEffect, createResource, createSignal, onCleanup, onMount, Show } from "solid-js";
import type { Task } from "~/../bindings/github.com/bal16/ariadm/internal/domain/task/models";
import { Events } from "@wailsio/runtime";
import { Preferences } from "~/components/Preferences";
import { AddTaskDialog } from "~/components/AddTaskDialog";
import { AboutDialog } from "~/components/AboutDialog";
import { KeyboardShortcutsDialog } from "~/components/KeyboardShortcutsDialog";
import { DownloadList } from "~/components/DownloadList";
import { DeleteConfirmDialog } from "~/components/DeleteConfirmDialog";
import { QuitConfirmDialog } from "~/components/QuitConfirmDialog";
import { MenuBar } from "~/components/layout/MenuBar";
import { Toolbar } from "~/components/layout/Toolbar";
import { StatusBar } from "~/components/layout/StatusBar";
import { useAppCloseRequest } from "~/hooks/useAppCloseRequest";
import { useKeyboardShortcuts } from "~/hooks/useKeyboardShortcuts";
import { useTaskDeletion } from "~/hooks/useTaskDeletion";
import { useTaskQueue } from "~/hooks/useTaskQueue";
import { useTheme } from "~/hooks/useTheme";
import {
  GetApplicationConfig,
  OpenPath,
  SaveApplicationConfig,
  SyncTheme,
} from "~/../bindings/github.com/bal16/ariadm/internal/ingress/wailsbridge/wailsbridge";

export default function App() {
  const [showPrefs, setShowPrefs] = createSignal(false);
  const [showAddTask, setShowAddTask] = createSignal(false);
  const [showQuitDialog, setShowQuitDialog] = createSignal(false);
  const [showShortcuts, setShowShortcuts] = createSignal(false);
  const [showAbout, setShowAbout] = createSignal(false);
  const [searchQuery, setSearchQuery] = createSignal("");
  const { themePref, setTheme } = useTheme();

  const [appConfig, { refetch: refetchAppConfig }] = createResource(GetApplicationConfig);
  const showInAppMenu = () =>
    appConfig()?.show_in_app_menu ??
    (typeof navigator !== "undefined" && /linux/i.test(navigator.userAgent));

  createEffect(() => {
    const cfg = appConfig();
    if (cfg?.theme_preference) {
      setTheme(cfg.theme_preference as "light" | "dark" | "system");
    }
  });

  const { tasks, runtimeStatus, totalSpeed, togglePause, pauseAll, resumeAll, deleteTask } =
    useTaskQueue();
  const deletion = useTaskDeletion(deleteTask);

  useAppCloseRequest(() => setShowQuitDialog(true));
  useKeyboardShortcuts({
    onNewTask: () => setShowAddTask((open) => !open),
    onPreferences: () => setShowPrefs((open) => !open),
  });

  const openFolder = async (task: Task) => {
    let folder = task.download_path;
    if (!folder) {
      try {
        const cfg = await GetApplicationConfig();
        folder = cfg?.default_download_path ?? "";
      } catch {
        folder = "";
      }
    }
    if (folder) void OpenPath(folder);
  };

  const toggleInAppMenu = async () => {
    const cfg = appConfig();
    if (!cfg) return;
    try {
      await SaveApplicationConfig({ ...cfg, show_in_app_menu: !cfg.show_in_app_menu });
      void refetchAppConfig();
    } catch {
      // ignore
    }
  };

  const notifiedCompletions = new Set<string>();

  createEffect(() => {
    for (const task of tasks) {
      if (task.status === "complete" && !notifiedCompletions.has(task.id)) {
        notifiedCompletions.add(task.id);
        if (typeof window !== "undefined" && "Notification" in window) {
          const show = () => {
            try {
              new Notification("Download complete", {
                body: task.file_name || "A download has finished.",
              });
            } catch {
              // Ignore notification errors in restricted webviews.
            }
          };
          if (Notification.permission === "granted") {
            show();
          } else if (Notification.permission !== "denied") {
            Notification.requestPermission()
              .then((permission) => {
                if (permission === "granted") show();
              })
              .catch(() => {});
          }
        }
      }
    }
  });

  onMount(() => {
    void SyncTheme(themePref());
    Events.On("menu:new-task", () => setShowAddTask(true));
    Events.On("menu:preferences", () => setShowPrefs(true));
    Events.On("menu:shortcuts", () => setShowShortcuts(true));
    Events.On("menu:about", () => setShowAbout(true));
    Events.On("theme:set", (e) => {
      const mode = (e?.data as string) ?? "system";
      if (mode === "light" || mode === "dark" || mode === "system") {
        setTheme(mode);
        const cfg = appConfig();
        if (cfg) void SaveApplicationConfig({ ...cfg, theme_preference: mode });
        void SyncTheme(mode);
      }
    });
    Events.On("config:updated", () => void refetchAppConfig());

    onCleanup(() => {
      Events.Off("menu:new-task", "menu:preferences", "config:updated");
    });
  });

  const quitCompletely = () => {
    void Events.Emit("app:force-quit");
  };

  const runInBackground = () => {
    setShowQuitDialog(false);
    void Events.Emit("app:background");
  };

  return (
    <div class="h-screen w-screen flex flex-col overflow-hidden relative select-none bg-background text-foreground font-sans antialiased">
      <Show when={showInAppMenu()}>
        <MenuBar
          onNewTask={() => setShowAddTask(true)}
          onOpenFolder={() => void openFolder({ download_path: "" } as Task)}
          onPreferences={() => setShowPrefs(true)}
          onPauseAll={() => void pauseAll()}
          onResumeAll={() => void resumeAll()}
          onToggleInAppMenu={() => void toggleInAppMenu()}
          onShowShortcuts={() => setShowShortcuts(true)}
          onShowAbout={() => setShowAbout(true)}
          themePref={themePref}
          onSetTheme={(theme) => {
            setTheme(theme);
            const cfg = appConfig();
            if (cfg) void SaveApplicationConfig({ ...cfg, theme_preference: theme });
            void SyncTheme(theme);
          }}
        />
      </Show>

      <Toolbar
        onNewTask={() => setShowAddTask(true)}
        onPauseAll={() => void pauseAll()}
        onResumeAll={() => void resumeAll()}
        searchQuery={searchQuery()}
        onSearchChange={setSearchQuery}
      />

      <div class="flex-1 p-2 bg-background/50 overflow-hidden">
        <DownloadList
          tasks={tasks}
          onTogglePause={togglePause}
          onDelete={deletion.request}
          onNewTask={() => setShowAddTask(true)}
          searchQuery={searchQuery()}
          onOpenFolder={openFolder}
        />
      </div>

      <StatusBar runtimeStatus={runtimeStatus()} downloadSpeed={totalSpeed()} />

      <Show when={showPrefs()}>
        <Preferences onClose={() => setShowPrefs(false)} />
      </Show>

      <Show when={showAddTask()}>
        <AddTaskDialog onClose={() => setShowAddTask(false)} />
      </Show>

      <Show when={showShortcuts()}>
        <KeyboardShortcutsDialog onClose={() => setShowShortcuts(false)} />
      </Show>

      <Show when={showAbout()}>
        <AboutDialog onClose={() => setShowAbout(false)} />
      </Show>

      <Show when={deletion.taskToDelete() !== null}>
        <DeleteConfirmDialog
          onClose={deletion.cancel}
          onConfirm={deletion.confirm}
          isDeleting={deletion.isDeleting()}
        />
      </Show>

      <Show when={showQuitDialog()}>
        <QuitConfirmDialog
          onClose={() => setShowQuitDialog(false)}
          onQuit={quitCompletely}
          onBackground={runInBackground}
        />
      </Show>
    </div>
  );
}
