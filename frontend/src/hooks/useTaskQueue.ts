import { createSignal, onCleanup, onMount, type Accessor } from "solid-js";
import { createStore, reconcile } from "solid-js/store";
import { Events } from "@wailsio/runtime";
import type { Task } from "~/../bindings/github.com/bal16/ariadm/internal/domain/task/models";
import {
  EngineState,
  ExtensionState,
  type Status as RuntimeStatus,
} from "~/../bindings/github.com/bal16/ariadm/internal/appstate/models";
import {
  DeleteTask,
  GetRuntimeStatus,
  GetTasks,
  PauseAllTasks,
  ResumeAllTasks,
  ToggleTaskPauseState,
} from "~/../bindings/github.com/bal16/ariadm/internal/ingress/wailsbridge/wailsbridge";

const initialRuntimeStatus: RuntimeStatus = {
  engine: EngineState.EngineStarting,
  engine_message: "",
  engine_port: 0,
  extension: ExtensionState.ExtensionUnavailable,
  extension_message: "",
  extension_port: 0,
};

export interface TaskQueue {
  tasks: Task[];
  runtimeStatus: Accessor<RuntimeStatus>;
  totalSpeed: Accessor<number>;
  togglePause: (id: string) => Promise<void>;
  pauseAll: () => Promise<void>;
  resumeAll: () => Promise<void>;
  deleteTask: (id: string, deleteFiles: boolean) => Promise<void>;
}

/**
 * Owns the task list: initial fetch, live event subscriptions, and the
 * mutations exposed to the UI. Event listeners are registered and cleaned up
 * with the owning component.
 */
export function useTaskQueue(): TaskQueue {
  const [tasks, setTasks] = createStore<Task[]>([]);
  const [runtimeStatus, setRuntimeStatus] = createSignal<RuntimeStatus>(initialRuntimeStatus);

  onMount(() => {
    let active = true;

    const subscribe = () => {
      if (!active) return;

      Events.On("tasks:update", (event) => {
        const incomingTasks = event.data as Task[] | null;
        if (!incomingTasks) {
          setTasks([]);
          return;
        }
        setTasks(reconcile(incomingTasks, { key: "id" }));
      });

      Events.On("runtime:status", (event) => {
        const status = event.data as RuntimeStatus | null;
        if (status) setRuntimeStatus(status);
      });
    };

    const loadInitialState = async () => {
      // The webview is recreated when the app returns from background mode, so
      // the status must be fetched rather than waiting for the next event.
      try {
        setRuntimeStatus(await GetRuntimeStatus());
      } catch (err) {
        console.error("Failed to load runtime status:", err);
      }

      try {
        const initialTasks = await GetTasks();
        if (initialTasks) {
          setTasks(initialTasks.filter((task): task is Task => task !== null));
        }
      } catch (err) {
        console.error("Failed to load initial task records:", err);
      }

      subscribe();
    };

    void loadInitialState();

    onCleanup(() => {
      active = false;
      Events.Off("tasks:update", "runtime:status");
    });
  });

  const totalSpeed = () =>
    tasks.reduce((total, task) => total + (task.status === "active" ? task.speed : 0), 0);

  const togglePause = async (id: string) => {
    try {
      await ToggleTaskPauseState(id);
    } catch (err) {
      console.error("Failed to transition task status:", err);
    }
  };

  const pauseAll = async () => {
    try {
      await PauseAllTasks();
    } catch (err) {
      console.error("Failed to pause all tasks:", err);
    }
  };

  const resumeAll = async () => {
    try {
      await ResumeAllTasks();
    } catch (err) {
      console.error("Failed to resume all tasks:", err);
    }
  };

  const deleteTask = async (id: string, deleteFiles: boolean) => {
    await DeleteTask(id, deleteFiles);
    setTasks((current) => current.filter((task) => task.id !== id));
  };

  return { tasks, runtimeStatus, totalSpeed, togglePause, pauseAll, resumeAll, deleteTask };
}
