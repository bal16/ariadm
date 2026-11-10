import { onCleanup, onMount } from "solid-js";
import { Events } from "@wailsio/runtime";

/**
 * Invokes the callback when the desktop shell asks the frontend to close,
 * for example after the window close button is pressed.
 */
export function useAppCloseRequest(onRequestClose: () => void) {
  onMount(() => {
    Events.On("app:request-close", onRequestClose);
    onCleanup(() => Events.Off("app:request-close"));
  });
}
