import { Bell, PanelLeftClose, PanelLeftOpen, Search, Sparkles } from "lucide-react";
import { AppLogo } from "@/components/app-logo";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAppStore } from "@/store/app-store";

export function TopNav() {
  const { isSidebarCollapsed, toggleSidebar } = useAppStore();

  return (
    <header className="flex items-center justify-between gap-4 border-b border-white/6 px-5 py-4">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleSidebar}
          aria-label="Toggle sidebar"
        >
          {isSidebarCollapsed ? (
            <PanelLeftOpen className="h-4 w-4" />
          ) : (
            <PanelLeftClose className="h-4 w-4" />
          )}
        </Button>
        <AppLogo />
      </div>

      <div className="hidden min-w-0 flex-1 items-center justify-center lg:flex">
        <div className="flex w-full max-w-xl items-center gap-3 rounded-full border border-white/8 bg-white/4 px-4 py-3">
          <Search className="h-4 w-4 text-muted-foreground" />
          <span className="truncate text-sm text-muted-foreground">
            Inspect structures, infer schemas, and map relational tables.
          </span>
        </div>
      </div>

      <div className="flex items-center gap-3">
        <Badge variant="accent" className="hidden md:inline-flex">
          <Sparkles className="mr-2 h-3.5 w-3.5" />
          Ready for import
        </Badge>
        <Button variant="outline" size="icon" aria-label="Notifications">
          <Bell className="h-4 w-4" />
        </Button>
      </div>
    </header>
  );
}
