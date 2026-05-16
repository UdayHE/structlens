import { useEffect, useRef, useState } from "react";
import { open as openDialog } from "@tauri-apps/api/dialog";
import { appWindow } from "@tauri-apps/api/window";
import {
  AlertTriangle,
  FileCode2,
  FileJson2,
  FolderUp,
  Sparkles,
  UploadCloud,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { analyzeStructuredData } from "@/lib/structlens-ipc";
import { formatFileSize } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";

const supportedExtensions = new Set(["json", "xml"]);
const loadingPhases = [
  { message: "Reading source", stage: "reading" as const },
  { message: "Parsing...", stage: "parsing" as const },
  { message: "Inferring schema...", stage: "inferring" as const },
  { message: "Building tree...", stage: "building" as const },
];

function getExtension(fileName: string) {
  return fileName.split(".").pop()?.toLowerCase() ?? "";
}

function fileNameFromPath(filePath: string) {
  return filePath.split(/[\\/]/).pop() ?? filePath;
}

export function UploadCard() {
  const loadingTimerRef = useRef<number | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  const {
    analysisError,
    analysisStage,
    clearAnalysis,
    setActiveSection,
    setAnalysisError,
    setAnalysisResult,
    setAnalysisStage,
    setUploadedFile,
    updateUploadedFileStatus,
    uploadedFile,
  } = useAppStore();

  const stopLoadingTimer = () => {
    if (loadingTimerRef.current !== null) {
      window.clearInterval(loadingTimerRef.current);
      loadingTimerRef.current = null;
    }
  };

  useEffect(() => stopLoadingTimer, []);

  const toErrorMessage = (error: unknown) => {
    if (typeof error === "string") return error;
    if (error instanceof Error) return error.message;
    if (
      typeof error === "object" &&
      error !== null &&
      "message" in error &&
      typeof error.message === "string"
    ) {
      return error.message;
    }
    return "Unable to analyze the selected file.";
  };

  const commitFileByPath = async (filePath: string) => {
    const fileName = fileNameFromPath(filePath);
    const extension = getExtension(fileName);

    if (!supportedExtensions.has(extension)) {
      clearAnalysis();
      setUploadedFile({
        name: fileName,
        size: 0,
        extension: "json",
        status: "invalid",
        message: "Unsupported file type",
      });
      setAnalysisError("Unsupported file type. Supported formats: .json and .xml.");
      return;
    }

    clearAnalysis();
    setUploadedFile({
      name: fileName,
      size: 0,
      extension: extension as "json" | "xml",
      status: "parsing",
      message: loadingPhases[0].message,
    });
    setAnalysisStage(loadingPhases[0].stage);

    stopLoadingTimer();
    let phaseIndex = 0;
    loadingTimerRef.current = window.setInterval(() => {
      phaseIndex = Math.min(phaseIndex + 1, loadingPhases.length - 1);
      const phase = loadingPhases[phaseIndex];
      setAnalysisStage(phase.stage);
      updateUploadedFileStatus("parsing", phase.message);
    }, 500);

    try {
      const result = await analyzeStructuredData({ fileName, filePath });
      stopLoadingTimer();
      setAnalysisResult(result);
      updateUploadedFileStatus("parsed", "Schema ready");
      setActiveSection("tree");
    } catch (error) {
      stopLoadingTimer();
      const message = toErrorMessage(error);
      setAnalysisError(message);
      updateUploadedFileStatus("invalid", message);
    }
  };

  // Keep a ref so the Tauri event listener always calls the latest version.
  const commitFileByPathRef = useRef(commitFileByPath);
  commitFileByPathRef.current = commitFileByPath;

  // Use Tauri's native file-drop event — the only way to get real paths on macOS.
  useEffect(() => {
    let unlistenFn: (() => void) | null = null;

    appWindow
      .onFileDropEvent((event) => {
        if (event.payload.type === "hover") {
          setIsDragging(true);
        } else if (event.payload.type === "cancel") {
          setIsDragging(false);
        } else if (event.payload.type === "drop") {
          setIsDragging(false);
          const paths = event.payload.paths;
          if (paths.length > 0) {
            void commitFileByPathRef.current(paths[0]);
          }
        }
      })
      .then((fn) => {
        unlistenFn = fn;
      });

    return () => {
      unlistenFn?.();
    };
  }, []);

  const handleBrowse = () => {
    void openDialog({
      multiple: false,
      filters: [{ name: "Structured Data", extensions: ["json", "xml"] }],
    }).then((selected) => {
      if (selected && typeof selected === "string") {
        void commitFileByPath(selected);
      }
    });
  };

  const statusVariant =
    uploadedFile?.status === "parsed"
      ? "success"
      : uploadedFile?.status === "parsing"
        ? "warning"
        : uploadedFile?.status === "invalid"
          ? "destructive"
          : "accent";

  return (
    <div className="flex h-full flex-col items-center justify-center p-6">
      <Card className="relative w-full max-w-3xl overflow-hidden">
        <div className="pointer-events-none absolute inset-x-12 top-0 h-28 rounded-full bg-surface-glow blur-3xl" />
        <CardHeader className="items-center text-center">
          <Badge variant="accent">
            <Sparkles className="mr-2 h-3.5 w-3.5" />
            Import Workspace
          </Badge>
          <CardTitle className="text-3xl tracking-[0.04em]">
            Drop XML or JSON here
          </CardTitle>
          <CardDescription className="max-w-xl">
            Start with a source file and StructLens will prepare the tree, schema,
            and SQL workspaces for review.
          </CardDescription>
        </CardHeader>

        <CardContent className="space-y-6">
          <div
            className={`group relative flex min-h-[280px] flex-col items-center justify-center rounded-[30px] border border-dashed px-6 text-center transition duration-200 ${
              isDragging
                ? "border-accent bg-accent/10 shadow-[0_0_40px_rgba(111,231,255,0.12)]"
                : "border-white/12 bg-white/[0.03]"
            }`}
          >
            <div className="mb-6 flex h-20 w-20 items-center justify-center rounded-[28px] border border-accent/25 bg-accent/10 text-accent shadow-[0_0_30px_rgba(111,231,255,0.12)] transition duration-200">
              <UploadCloud className="h-9 w-9" />
            </div>
            <h3 className="text-xl font-semibold">Drag a file into the workspace</h3>
            <p className="mt-3 max-w-md text-sm leading-6 text-muted-foreground">
              Drop anywhere in the window, or click Browse to pick a file. Supported
              formats are `.json` and `.xml` only.
            </p>
            <div className="mt-6 flex flex-wrap items-center justify-center gap-3">
              <Badge variant="muted">
                <FileJson2 className="mr-2 h-3.5 w-3.5" />
                JSON
              </Badge>
              <Badge variant="muted">
                <FileCode2 className="mr-2 h-3.5 w-3.5" />
                XML
              </Badge>
            </div>
          </div>

          <div className="flex flex-col gap-4 rounded-[24px] border border-white/8 bg-black/10 p-5 md:flex-row md:items-center md:justify-between">
            <div className="space-y-2">
              <p className="text-sm font-medium text-foreground">Current selection</p>
              {uploadedFile ? (
                <div className="space-y-1">
                  <p className="text-sm text-muted-foreground">
                    Filename: <span className="text-foreground">{uploadedFile.name}</span>
                  </p>
                  {uploadedFile.size > 0 && (
                    <p className="text-sm text-muted-foreground">
                      Size: <span className="text-foreground">{formatFileSize(uploadedFile.size)}</span>
                    </p>
                  )}
                  <p className="text-sm text-muted-foreground">
                    Parse status: <span className="text-foreground">{uploadedFile.message}</span>
                  </p>
                </div>
              ) : (
                <p className="text-sm text-muted-foreground">
                  No file selected yet. Upload a document to begin.
                </p>
              )}
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <Badge variant={statusVariant}>{uploadedFile?.status ?? "idle"}</Badge>
              <Button variant="outline" onClick={handleBrowse}>
                <FolderUp className="h-4 w-4" />
                Browse files
              </Button>
            </div>
          </div>

          {analysisError ? (
            <div className="flex items-start gap-3 rounded-[24px] border border-destructive/30 bg-destructive/8 p-5 text-sm text-destructive">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <p className="font-medium">Analysis error</p>
                <p className="mt-1 text-destructive/90">{analysisError}</p>
              </div>
            </div>
          ) : null}

          {analysisStage !== "idle" && analysisStage !== "ready" && !analysisError ? (
            <div className="rounded-[24px] border border-warning/25 bg-warning/8 p-4 text-sm text-warning">
              {uploadedFile?.message ?? "Preparing analysis..."}
            </div>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
