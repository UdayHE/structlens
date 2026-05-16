import {
  Braces,
  Database,
  Files,
  PanelsTopLeft,
  Settings2,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";
import { type SidebarSection } from "@/types/app";

const sections: Array<{
  id: SidebarSection;
  label: string;
  description: string;
  icon: typeof Files;
}> = [
  { id: "files", label: "Files", description: "Source imports", icon: Files },
  { id: "tree", label: "Tree View", description: "Shape explorer", icon: PanelsTopLeft },
  { id: "schema", label: "Schema View", description: "Inference output", icon: Braces },
  { id: "sql", label: "SQL View", description: "DDL preview", icon: Database },
  { id: "settings", label: "Settings", description: "Project options", icon: Settings2 },
];

export function Sidebar() {
  const { activeSection, isSidebarCollapsed, setActiveSection } = useAppStore();

  return (
    <aside
      className={cn(
        "border-r border-white/6 bg-sidebar/90 px-3 py-4 transition-all duration-300",
        isSidebarCollapsed ? "w-[92px]" : "w-[280px]",
      )}
    >
      <nav className="flex h-full flex-col gap-2">
        {sections.map(({ id, label, description, icon: Icon }) => {
          const isActive = activeSection === id;

          return (
            <button
              key={id}
              type="button"
              onClick={() => setActiveSection(id)}
              className={cn(
                "group flex items-center gap-3 rounded-2xl border px-3 py-3 text-left transition duration-200",
                isActive
                  ? "border-accent/35 bg-accent/10 shadow-[0_0_24px_rgba(111,231,255,0.09)]"
                  : "border-transparent bg-transparent hover:border-white/8 hover:bg-white/4",
              )}
            >
              <div
                className={cn(
                  "flex h-11 w-11 shrink-0 items-center justify-center rounded-2xl border",
                  isActive
                    ? "border-accent/30 bg-accent/12 text-accent"
                    : "border-white/8 bg-white/4 text-muted-foreground group-hover:text-foreground",
                )}
              >
                <Icon className="h-4.5 w-4.5" />
              </div>
              {!isSidebarCollapsed ? (
                <div className="min-w-0">
                  <p className="text-sm font-medium text-foreground">{label}</p>
                  <p className="truncate text-xs text-muted-foreground">{description}</p>
                </div>
              ) : null}
            </button>
          );
        })}
      </nav>
    </aside>
  );
}
