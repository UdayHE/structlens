import { CircleDashed, DatabaseZap, FileJson2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { useAppStore } from "@/store/app-store";

export function StatusBar() {
  const { activeSection, analysisError, schemaMetadata, uploadedFile } = useAppStore();

  const badgeVariant =
    uploadedFile?.status === "parsed"
      ? "success"
      : uploadedFile?.status === "parsing"
        ? "warning"
        : uploadedFile?.status === "invalid"
          ? "destructive"
      : "muted";

  return (
    <footer className="flex flex-wrap items-center justify-between gap-3 border-t border-white/6 px-5 py-3 text-xs text-muted-foreground">
      <div className="flex items-center gap-4">
        <span className="flex items-center gap-2">
          <CircleDashed className="h-3.5 w-3.5" />
          Workspace: {activeSection}
        </span>
        <span className="flex items-center gap-2">
          <FileJson2 className="h-3.5 w-3.5" />
          {uploadedFile ? uploadedFile.name : "No file selected"}
        </span>
        {schemaMetadata ? (
          <span>
            {schemaMetadata.totalFields} fields • {schemaMetadata.tableCount} tables
          </span>
        ) : null}
      </div>

      <div className="flex items-center gap-3">
        <Badge variant={badgeVariant}>
          <DatabaseZap className="mr-2 h-3.5 w-3.5" />
          {analysisError ?? uploadedFile?.message ?? "Awaiting input"}
        </Badge>
      </div>
    </footer>
  );
}
