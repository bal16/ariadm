import { Show } from "solid-js";
import {
  Activity,
  Check,
  FolderOpen,
  Monitor,
  Moon,
  Pause,
  Play,
  Plus,
  Settings,
  Sun,
} from "lucide-solid";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuShortcut,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { useTheme } from "~/hooks/useTheme";

interface MenuBarProps {
  onNewTask: () => void;
  onOpenFolder: () => void;
  onPreferences: () => void;
  onPauseAll: () => void;
  onResumeAll: () => void;
  onToggleInAppMenu: () => void;
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
        <DropdownMenuContent class="w-48 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem
            onClick={props.onNewTask}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Plus class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">New Download</span>
            <DropdownMenuShortcut class="font-mono text-[10px]">Ctrl+N</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={props.onOpenFolder}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <FolderOpen class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Open Download Folder</span>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
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

      <DropdownMenu>
        <DropdownMenuTrigger
          as="button"
          class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-hidden"
        >
          Downloads
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-44 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem
            onClick={props.onPauseAll}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Pause class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Pause All</span>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={props.onResumeAll}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <Play class="h-3.5 w-3.5 text-muted-foreground" />
            <span class="flex-1">Resume All</span>
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
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={props.onToggleInAppMenu}
            class="text-xs font-medium cursor-pointer flex items-center space-x-2"
          >
            <span class="flex-1">Hide In-App Menu</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <DropdownMenu>
        <DropdownMenuTrigger
          as="button"
          class="px-2 py-1 text-muted-foreground hover:text-foreground hover:bg-muted/60 rounded-sm transition-colors focus:outline-hidden"
        >
          Help
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-52 bg-popover border border-border text-popover-foreground shadow-md">
          <DropdownMenuItem class="text-xs font-medium flex items-center justify-between">
            <span>New Download</span>
            <DropdownMenuShortcut class="font-mono text-[10px]">Ctrl+N</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuItem class="text-xs font-medium flex items-center justify-between">
            <span>Preferences</span>
            <DropdownMenuShortcut class="font-mono text-[10px]">Ctrl+P</DropdownMenuShortcut>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem class="text-xs font-medium">
            <span>About AriaDM</span>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}
