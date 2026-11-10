import { onCleanup, onMount } from "solid-js";

export interface KeyboardShortcutHandlers {
  onNewTask: () => void;
  onPreferences: () => void;
}

/**
 * Binds the desktop shortcuts: Ctrl+N for a new task and Ctrl+P for
 * preferences. The listeners are scoped to the owning component.
 */
export function useKeyboardShortcuts(handlers: KeyboardShortcutHandlers) {
  onMount(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (!event.ctrlKey) return;

      const key = event.key.toLowerCase();
      if (key === "n") {
        event.preventDefault();
        handlers.onNewTask();
      } else if (key === "p") {
        event.preventDefault();
        handlers.onPreferences();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    onCleanup(() => window.removeEventListener("keydown", handleKeyDown));
  });
}
