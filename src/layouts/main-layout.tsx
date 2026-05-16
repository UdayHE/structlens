import { Sidebar } from "@/components/sidebar";
import { StatusBar } from "@/components/status-bar";
import { TopNav } from "@/components/top-nav";
import { WorkspacePage } from "@/pages/workspace-page";

export function MainLayout() {
  return (
    <div className="app-grid min-h-screen p-4 text-foreground">
      <div className="glass-panel flex min-h-[calc(100vh-2rem)] flex-col overflow-hidden rounded-[32px] border border-white/8">
        <TopNav />
        <div className="flex min-h-0 flex-1">
          <Sidebar />
          <main className="min-h-0 flex-1 overflow-auto">
            <WorkspacePage />
          </main>
        </div>
        <StatusBar />
      </div>
    </div>
  );
}
