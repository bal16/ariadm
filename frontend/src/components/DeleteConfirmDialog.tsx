import { TriangleAlert, Trash2 } from "lucide-solid";

export function DeleteConfirmDialog(props: {
  onClose: () => void;
  onConfirm: () => void;
  isDeleting: boolean;
}) {
  return (
    <div class="absolute top-20 left-1/2 -translate-x-1/2 z-50 p-1 animate-in fade-in zoom-in-95 duration-150">
      <div class="bg-card text-card-foreground border border-border rounded-md shadow-2xl w-[380px] flex flex-col overflow-hidden">
        {/* Title Bar */}
        <div class="flex items-center justify-between px-2.5 py-1.5 border-b border-border bg-muted/60 select-none">
          <div class="flex items-center space-x-1.5 text-xs font-medium text-destructive">
            <TriangleAlert class="h-3.5 w-3.5" />
            <span>Confirm Deletion</span>
          </div>
        </div>

        {/* Body */}
        <div class="p-4 bg-background space-y-3">
          <p class="text-sm text-foreground">
            Are you sure you want to delete this task?
          </p>
          <p class="text-[11px] text-muted-foreground">
            This action cannot be undone. The task will be removed from your download list and the active daemon.
          </p>

          <div class="flex justify-end space-x-1.5 pt-3 mt-4 border-t border-border">
            <button
              type="button"
              onClick={props.onClose}
              disabled={props.isDeleting}
              class="px-3 py-1 text-xs border border-border bg-muted/40 text-muted-foreground rounded-sm hover:bg-muted hover:text-foreground disabled:opacity-50 transition-colors"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={props.onConfirm}
              disabled={props.isDeleting}
              class="px-3 py-1 text-xs bg-destructive text-destructive-foreground font-semibold rounded-sm hover:bg-destructive/90 disabled:opacity-50 shadow-sm flex items-center space-x-1 transition-colors"
            >
              <Trash2 class="h-3 w-3" />
              <span>{props.isDeleting ? "Deleting..." : "Delete Task"}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
