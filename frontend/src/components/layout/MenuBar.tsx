import { Show } from "solid-js";
import { Activity, Check, Monitor, Moon, Plus, Settings, Sun } from "lucide-solid";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { useTheme } from "~/hooks/useTheme";

interface MenuBarProps {
  onNewTask: () => void;
  onPreferences: () => void;
}

export function MenuBar(props: MenuBarProps) {
  const { themePref, setTheme } = useTheme();

  return (
    <div class="flex items-center px-2 py-0.5 bg-muted/40 border-b border-border text-xs space-x-1 z-40">
      <span class="font-bold text-primary select-none font-mono px-2 mr-2 flex items-center space-x-1">
        <Activity class="h-3.5 w-3.5 text-primary" />
        <span>AriaDM</span>
      </span>

      <DropdownMenu>
        <DropdownMenuTrigger
          as="button"
          class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-hidden"
        >
          File
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-44 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem
            onClick={props.onNewTask}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Plus class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">New Task</span>
            <DropdownMenuShortcut class="font-mono text-[10px]">Ctrl+N</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <DropdownMenu>
        <DropdownMenuTrigger
          as="button"
          class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-hidden"
        >
          View
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-44 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem
            onSelect={() => setTheme("light")}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Sun class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Light Theme</span>
            <Show when={themePref() === "light"}>
              <Check class="h-3.5 w-3.5 text-primary" />
            </Show>
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => setTheme("dark")}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Moon class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Dark Theme</span>
            <Show when={themePref() === "dark"}>
              <Check class="h-3.5 w-3.5 text-primary" />
            </Show>
          </DropdownMenuItem>
          <DropdownMenuItem
            onSelect={() => setTheme("system")}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Monitor class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">System Theme</span>
            <Show when={themePref() === "system"}>
              <Check class="h-3.5 w-3.5 text-primary" />
            </Show>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <DropdownMenu>
        <DropdownMenuTrigger
          as="button"
          class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-hidden"
        >
          Tools
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-48 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem
            onClick={props.onPreferences}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Settings class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Preferences</span>
            <DropdownMenuShortcut class="font-mono text-[10px]">Ctrl+P</DropdownMenuShortcut>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <span class="text-muted-foreground/40 px-2 text-[11px] hover:text-foreground cursor-pointer transition-colors">
        Help
      </span>
    </div>
  );
}
