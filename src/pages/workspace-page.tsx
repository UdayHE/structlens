import { AnalysisWorkspace } from "@/features/workspace/analysis-workspace";
import { EmptyPanel } from "@/components/empty-panel";
import { UploadCard } from "@/features/upload/upload-card";
import { useAppStore } from "@/store/app-store";

export function WorkspacePage() {
  const { activeSection } = useAppStore();

  if (activeSection === "files") {
    return <UploadCard />;
  }

  if (activeSection === "tree") {
    return <AnalysisWorkspace />;
  }

  if (activeSection === "sql") {
    return <AnalysisWorkspace />;
  }

  return <EmptyPanel section={activeSection} />;
}
