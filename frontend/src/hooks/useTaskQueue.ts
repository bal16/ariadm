import { createSignal, onCleanup, onMount, type Accessor } from "solid-js";
import { createStore, reconcile } from "solid-js/store";
import { Events } from "@wailsio/runtime";
import type { Task } from "~/../bindings/github.com/bal16/ariadm/internal/domain/task/models";
import {
  DeleteTask,
  GetTasks,
  PauseAllTasks,
  ResumeAllTasks,
  ToggleTaskPauseState,
} from "~/../bindings/github.com/bal16/ariadm/internal/ingress/wailsbridge/wailsbridge";

export type EngineStatus = "Connecting" | "Running" | "Disconnected";

export interface TaskQueue {
  tasks: Task[];
  engineStatus: Accessor<EngineStatus>;
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
  const [engineStatus, setEngineStatus] = createSignal<EngineStatus>("Connecting");

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

      Events.On("engine:status", (event) => {
        const status = event.data as string;
        setEngineStatus(status === "running" ? "Running" : "Disconnected");
      });
    };

    const loadInitialTasks = async () => {
      try {
        const initialTasks = await GetTasks();
        if (initialTasks) {
          setTasks(initialTasks.filter((task): task is Task => task !== null));
          setEngineStatus("Running");
        }
      } catch (err) {
        console.error("Failed to load initial task records:", err);
        setEngineStatus("Disconnected");
      }

      subscribe();
    };

    void loadInitialTasks();

    onCleanup(() => {
      active = false;
      Events.Off("tasks:update", "engine:status");
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

  return { tasks, engineStatus, totalSpeed, togglePause, pauseAll, resumeAll, deleteTask };
}
