import { createSignal, type Accessor } from "solid-js";

export interface TaskDeletion {
  taskToDelete: Accessor<string | null>;
  isDeleting: Accessor<boolean>;
  request: (id: string) => void;
  cancel: () => void;
  confirm: (deleteFiles: boolean) => Promise<void>;
}

/**
 * Tracks the task selected for deletion and runs the confirmation flow.
 * The actual backend mutation is injected so the queue owns it.
 */
export function useTaskDeletion(
  deleteTask: (id: string, deleteFiles: boolean) => Promise<void>,
): TaskDeletion {
  const [taskToDelete, setTaskToDelete] = createSignal<string | null>(null);
  const [isDeleting, setIsDeleting] = createSignal(false);

  const request = (id: string) => setTaskToDelete(id);
  const cancel = () => setTaskToDelete(null);

  const confirm = async (deleteFiles: boolean) => {
    const id = taskToDelete();
    if (!id) return;

    setIsDeleting(true);
    try {
      await deleteTask(id, deleteFiles);
      setTaskToDelete(null);
    } catch (err) {
      console.error("Failed to delete task:", err);
    } finally {
      setIsDeleting(false);
    }
  };

  return { taskToDelete, isDeleting, request, cancel, confirm };
}
