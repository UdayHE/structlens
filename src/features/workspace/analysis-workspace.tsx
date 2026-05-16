import { useEffect, useRef, useState } from "react";
import { GripVertical, LayoutPanelTop, TableProperties } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { SchemaTreeView } from "@/features/tree/schema-tree-view";
import { SqlViewer } from "@/features/sql/sql-viewer";
import { cn } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";

const minPanePercent = 28;
const maxPanePercent = 72;
const storageKey = "structlens.workspace.split";

export function AnalysisWorkspace() {
  const { activeSection } = useAppStore();
  const [leftPanePercent, setLeftPanePercent] = useState(48);
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);
  const nextPanePercentRef = useRef(leftPanePercent);
  const frameRef = useRef<number | null>(null);

  useEffect(() => {
    const saved = window.localStorage.getItem(storageKey);
    if (!saved) {
      return;
    }
    const numericValue = Number(saved);
    if (!Number.isNaN(numericValue)) {
      setLeftPanePercent(Math.min(maxPanePercent, Math.max(minPanePercent, numericValue)));
    }
  }, []);

  useEffect(() => {
    window.localStorage.setItem(storageKey, String(leftPanePercent));
  }, [leftPanePercent]);

  const flushPaneSize = () => {
    frameRef.current = null;
    setLeftPanePercent(nextPanePercentRef.current);
  };

  const handlePointerMove = (event: PointerEvent) => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const bounds = container.getBoundingClientRect();
    const nextPercent = ((event.clientX - bounds.left) / bounds.width) * 100;
    nextPanePercentRef.current = Math.min(maxPanePercent, Math.max(minPanePercent, nextPercent));

    if (frameRef.current === null) {
      frameRef.current = window.requestAnimationFrame(flushPaneSize);
    }
  };

  const stopDragging = () => {
    setIsDragging(false);
    document.body.style.cursor = "";
    document.body.style.userSelect = "";
    window.removeEventListener("pointermove", handlePointerMove);
    window.removeEventListener("pointerup", stopDragging);
  };

  const startDragging = () => {
    setIsDragging(true);
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", stopDragging);
  };

  useEffect(
    () => () => {
      if (frameRef.current !== null) {
        window.cancelAnimationFrame(frameRef.current);
      }
      stopDragging();
    },
    [],
  );

  return (
    <div className="flex h-full flex-col gap-4 p-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="space-y-2">
          <Badge variant="accent" className="w-fit">
            <LayoutPanelTop className="mr-2 h-3.5 w-3.5" />
            Analysis Workspace
          </Badge>
          <p className="text-sm text-muted-foreground">
            Drag the divider to rebalance the tree explorer and SQL preview. Panel sizes persist.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={activeSection === "tree" ? "default" : "outline"}
            onClick={() => setLeftPanePercent(58)}
          >
            Tree Focus
          </Button>
          <Button
            variant={activeSection === "sql" ? "default" : "outline"}
            onClick={() => setLeftPanePercent(38)}
          >
            SQL Focus
          </Button>
        </div>
      </div>

      <div
        ref={containerRef}
        className="grid min-h-0 flex-1 gap-0 overflow-hidden rounded-[30px] border border-white/8 bg-black/10"
        style={{
          gridTemplateColumns: `${leftPanePercent}% 12px minmax(0, 1fr)`,
        }}
      >
        <div className="min-w-0 p-3">
          <SchemaTreeView isPrimary={activeSection === "tree"} />
        </div>

        <div
          role="separator"
          aria-orientation="vertical"
          aria-label="Resize workspace panels"
          onPointerDown={startDragging}
          className={cn(
            "group relative flex cursor-col-resize items-center justify-center bg-white/[0.02] transition-colors duration-150 hover:bg-accent/10",
            isDragging && "bg-accent/12",
          )}
        >
          <div className="absolute inset-y-0 left-1/2 w-px -translate-x-1/2 bg-white/8 group-hover:bg-accent/35" />
          <div className="relative z-10 rounded-full border border-white/10 bg-[#07131d] p-1 text-muted-foreground shadow-[0_10px_18px_rgba(0,0,0,0.22)] group-hover:border-accent/35 group-hover:text-accent">
            <GripVertical className="h-3.5 w-3.5" />
          </div>
        </div>

        <div className="min-w-0 p-3">
          <div className="flex h-full min-h-0 flex-col gap-3">
            <div className="flex items-center justify-end">
              <Badge variant="muted">
                <TableProperties className="mr-2 h-3.5 w-3.5" />
                Split-pane ready
              </Badge>
            </div>
            <div className="min-h-0 flex-1">
              <SqlViewer isPrimary={activeSection === "sql"} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
