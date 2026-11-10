import { Plus, Search } from "lucide-solid";
import { Button } from "~/components/ui/button";

interface ToolbarProps {
  onNewTask: () => void;
}

export function Toolbar(props: ToolbarProps) {
  return (
    <div class="flex items-center justify-between p-2 border-b border-border bg-muted/10 z-30">
      <div class="flex items-center space-x-1.5">
        <Button
          onClick={props.onNewTask}
          size="sm"
          class="h-7 px-2.5 text-xs bg-primary text-primary-foreground font-semibold rounded-sm shadow-xs hover:opacity-90 flex items-center space-x-1"
        >
          <Plus class="h-3.5 w-3.5" />
          <span>New Task</span>
        </Button>
      </div>
      <div class="relative">
        <Search class="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground/80" />
        <input
          type="text"
          placeholder="Search tasks..."
          class="pl-7 pr-2 py-1 h-7 text-xs border border-input bg-background text-foreground rounded-sm font-mono w-48 focus:outline-hidden focus:ring-1 focus:ring-ring"
        />
      </div>
    </div>
  );
}
