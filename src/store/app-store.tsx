import {
  createContext,
  useContext,
  useMemo,
  useState,
  type PropsWithChildren,
} from "react";
import {
  type ActiveSchemaSelection,
  type AnalysisStage,
  type AnalyzeInputResponse,
  type RecordGroup,
  type SchemaMetadata,
  type SchemaTreeNode,
  type SidebarSection,
  type UploadedFile,
} from "@/types/app";

interface AppStoreValue {
  activeSection: SidebarSection;
  isSidebarCollapsed: boolean;
  uploadedFile: UploadedFile | null;
  analysisStage: AnalysisStage;
  analysisError: string | null;
  schemaTree: SchemaTreeNode[] | null;
  generatedSQL: string;
  schemaMetadata: SchemaMetadata | null;
  records: RecordGroup[] | null;
  activeSchemaSelection: ActiveSchemaSelection | null;
  setActiveSection: (section: SidebarSection) => void;
  toggleSidebar: () => void;
  setUploadedFile: (file: UploadedFile | null) => void;
  updateUploadedFileStatus: (
    status: UploadedFile["status"],
    message: string,
  ) => void;
  setAnalysisStage: (stage: AnalysisStage) => void;
  setAnalysisResult: (result: AnalyzeInputResponse) => void;
  setAnalysisError: (message: string) => void;
  setActiveSchemaSelection: (selection: ActiveSchemaSelection | null) => void;
  clearAnalysis: () => void;
}

const AppStoreContext = createContext<AppStoreValue | null>(null);

export function AppStoreProvider({ children }: PropsWithChildren) {
  const [activeSection, setActiveSection] = useState<SidebarSection>("files");
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);
  const [uploadedFile, setUploadedFile] = useState<UploadedFile | null>(null);
  const [analysisStage, setAnalysisStage] = useState<AnalysisStage>("idle");
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const [schemaTree, setSchemaTree] = useState<SchemaTreeNode[] | null>(null);
  const [generatedSQL, setGeneratedSQL] = useState("");
  const [schemaMetadata, setSchemaMetadata] = useState<SchemaMetadata | null>(null);
  const [records, setRecords] = useState<RecordGroup[] | null>(null);
  const [activeSchemaSelection, setActiveSchemaSelection] =
    useState<ActiveSchemaSelection | null>(null);

  const value = useMemo<AppStoreValue>(
    () => ({
      activeSection,
      isSidebarCollapsed,
      uploadedFile,
      analysisStage,
      analysisError,
      schemaTree,
      generatedSQL,
      schemaMetadata,
      records,
      activeSchemaSelection,
      setActiveSection,
      toggleSidebar: () => setIsSidebarCollapsed((current) => !current),
      setUploadedFile,
      updateUploadedFileStatus: (status, message) =>
        setUploadedFile((current) =>
          current
            ? {
                ...current,
                status,
                message,
              }
            : current,
        ),
      setAnalysisStage,
      setAnalysisResult: (result) => {
        setSchemaTree(result.schemaTree);
        setGeneratedSQL(result.sql);
        setSchemaMetadata(result.metadata);
        setRecords(result.records ?? null);
        setActiveSchemaSelection(null);
        setAnalysisError(null);
        setAnalysisStage("ready");
      },
      setAnalysisError: (message) => {
        setSchemaTree(null);
        setGeneratedSQL("");
        setSchemaMetadata(null);
        setRecords(null);
        setActiveSchemaSelection(null);
        setAnalysisError(message);
        setAnalysisStage("error");
      },
      setActiveSchemaSelection,
      clearAnalysis: () => {
        setSchemaTree(null);
        setGeneratedSQL("");
        setSchemaMetadata(null);
        setRecords(null);
        setActiveSchemaSelection(null);
        setAnalysisError(null);
        setAnalysisStage("idle");
      },
    }),
    [
      activeSection,
      activeSchemaSelection,
      analysisError,
      analysisStage,
      generatedSQL,
      isSidebarCollapsed,
      records,
      schemaMetadata,
      schemaTree,
      uploadedFile,
    ],
  );

  return (
    <AppStoreContext.Provider value={value}>{children}</AppStoreContext.Provider>
  );
}

export function useAppStore() {
  const context = useContext(AppStoreContext);

  if (!context) {
    throw new Error("useAppStore must be used within AppStoreProvider");
  }

  return context;
}
