import { ClipboardList } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";
import { type RecordGroup, type RecordInstance } from "@/types/app";

export function RecordsView() {
  const { records, analysisStage, analysisError } = useAppStore();

  if (analysisError) {
    return (
      <div className="flex min-h-[420px] items-center justify-center p-6 text-center text-sm text-destructive">
        {analysisError}
      </div>
    );
  }

  if (analysisStage === "idle") {
    return (
      <div className="flex min-h-[420px] flex-col items-center justify-center gap-3 p-6 text-center">
        <ClipboardList className="h-8 w-8 text-muted-foreground/40" />
        <p className="text-sm text-muted-foreground">
          Upload a file to explore its records.
        </p>
      </div>
    );
  }

  if (!records || records.length === 0) {
    return (
      <div className="flex min-h-[420px] flex-col items-center justify-center gap-3 p-6 text-center">
        <ClipboardList className="h-8 w-8 text-muted-foreground/40" />
        <p className="text-sm text-muted-foreground">
          No attribute-based records found. Records view works best with XML files.
        </p>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto px-1 py-2 space-y-5">
      {records.map((group) => (
        <RecordGroupSection key={group.typeName} group={group} />
      ))}
    </div>
  );
}

function RecordGroupSection({ group }: { group: RecordGroup }) {
  return (
    <div>
      <div className="flex items-center gap-2 mb-2 px-1">
        <span className="font-mono text-sm font-semibold text-foreground">
          {group.typeName}
        </span>
        <Badge variant="muted" className="text-[10px]">
          {group.instances.length}
        </Badge>
      </div>
      <div className="space-y-2">
        {group.instances.map((instance, idx) => (
          <RecordInstanceCard key={idx} instance={instance} />
        ))}
      </div>
    </div>
  );
}

function RecordInstanceCard({ instance }: { instance: RecordInstance }) {
  const entries = Object.entries(instance.attributes ?? {}).sort(([a], [b]) =>
    a.localeCompare(b),
  );

  // Split into name-like (ending with "Name") and the rest
  const nameEntries = entries.filter(
    ([k]) => k.toLowerCase().endsWith("name") && !k.toLowerCase().startsWith("parent"),
  );
  const parentEntries = entries.filter(([k]) =>
    k.toLowerCase().startsWith("parent"),
  );
  const otherEntries = entries.filter(
    ([k]) =>
      !k.toLowerCase().endsWith("name") && !k.toLowerCase().startsWith("parent"),
  );

  return (
    <div className="rounded-xl border border-white/8 bg-white/[0.025] p-3 space-y-2">
      {/* Headline: name attrs + parent */}
      {(nameEntries.length > 0 || parentEntries.length > 0) && (
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-xs">
          {nameEntries.map(([k, v]) => (
            <span key={k} className="font-medium text-foreground">
              {v || <span className="text-muted-foreground italic">empty</span>}
            </span>
          ))}
          {parentEntries.map(([k, v]) => (
            <span key={k} className="flex items-center gap-1 text-muted-foreground">
              <span className="text-accent/70">←</span>
              <span>{v || "—"}</span>
            </span>
          ))}
        </div>
      )}

      {/* Other attributes grid */}
      {otherEntries.length > 0 && (
        <div
          className={cn(
            "grid gap-x-3 gap-y-0.5 text-[11px]",
            otherEntries.length > 4
              ? "grid-cols-3"
              : otherEntries.length > 2
                ? "grid-cols-2"
                : "grid-cols-1",
          )}
        >
          {otherEntries.map(([k, v]) => (
            <div key={k} className="flex gap-1 min-w-0">
              <span className="text-muted-foreground shrink-0">{k}:</span>
              <span
                className={cn(
                  "truncate",
                  v === "" ? "text-muted-foreground/50 italic" : "text-foreground",
                )}
              >
                {v === "" ? "—" : v}
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Children */}
      {instance.children && instance.children.length > 0 && (
        <div className="pt-1 border-t border-white/6 flex flex-wrap gap-x-4 gap-y-1">
          {instance.children.map((cg) => (
            <div key={cg.name} className="text-[11px] flex items-center gap-1.5">
              <span className="text-accent/80 font-medium">{cg.name}</span>
              <span className="text-muted-foreground">×{cg.count}</span>
              {cg.keyValues && cg.keyValues.length > 0 && (
                <span className="text-muted-foreground truncate max-w-[200px]">
                  {cg.keyValues.slice(0, 5).join(", ")}
                  {cg.keyValues.length > 5 && "…"}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
