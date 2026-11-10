import { createSignal, Show } from "solid-js";
import { Events } from "@wailsio/runtime";
import { Preferences } from "~/components/Preferences";
import { AddTaskDialog } from "~/components/AddTaskDialog";
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

export default function App() {
  const [showPrefs, setShowPrefs] = createSignal(false);
  const [showAddTask, setShowAddTask] = createSignal(false);
  const [showQuitDialog, setShowQuitDialog] = createSignal(false);

  const { tasks, engineStatus, totalSpeed, togglePause, pauseAll, resumeAll, deleteTask } =
    useTaskQueue();
  const deletion = useTaskDeletion(deleteTask);

  useAppCloseRequest(() => setShowQuitDialog(true));
  useKeyboardShortcuts({
    onNewTask: () => setShowAddTask((open) => !open),
    onPreferences: () => setShowPrefs((open) => !open),
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
      <MenuBar onNewTask={() => setShowAddTask(true)} onPreferences={() => setShowPrefs(true)} />

      <Toolbar
        onNewTask={() => setShowAddTask(true)}
        onPauseAll={() => void pauseAll()}
        onResumeAll={() => void resumeAll()}
      />

      <div class="flex-1 p-2 bg-background/50 overflow-hidden">
        <DownloadList tasks={tasks} onTogglePause={togglePause} onDelete={deletion.request} />
      </div>

      <StatusBar engineStatus={engineStatus()} downloadSpeed={totalSpeed()} />

      <Show when={showPrefs()}>
        <Preferences onClose={() => setShowPrefs(false)} />
      </Show>

      <Show when={showAddTask()}>
        <AddTaskDialog onClose={() => setShowAddTask(false)} />
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
