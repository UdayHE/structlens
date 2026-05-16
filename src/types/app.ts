export type SidebarSection =
  | "files"
  | "tree"
  | "schema"
  | "sql"
  | "settings";

export type ParseStatus = "idle" | "ready" | "parsing" | "parsed" | "invalid";
export type AnalysisStage =
  | "idle"
  | "reading"
  | "parsing"
  | "inferring"
  | "building"
  | "ready"
  | "error";

export interface UploadedFile {
  name: string;
  size: number;
  extension: "json" | "xml";
  status: ParseStatus;
  message: string;
}

export type SchemaValueType =
  | "object"
  | "array"
  | "string"
  | "number"
  | "boolean"
  | "null"
  | "mixed";

export interface SchemaTreeNode {
  id: string;
  name: string;
  path: string;
  type: SchemaValueType;
  childCount: number;
  optional?: boolean;
  isArray?: boolean;
  children?: SchemaTreeNode[];
}

export interface SchemaMetadata {
  totalFields: number;
  arrayFields: number;
  optionalFields: number;
  tableCount: number;
}

export interface AnalyzeInputRequest {
  fileName: string;
  filePath: string;
  flattenThreshold?: number;
  arrayItemName?: string;
}

export interface RecordChildGroup {
  name: string;
  count: number;
  keyValues?: string[];
}

export interface RecordInstance {
  attributes: Record<string, string>;
  children?: RecordChildGroup[];
}

export interface RecordGroup {
  typeName: string;
  instances: RecordInstance[];
}

export interface AnalyzeInputResponse {
  schemaTree: SchemaTreeNode[];
  sql: string;
  metadata: SchemaMetadata;
  records: RecordGroup[];
}

export interface ActiveSchemaSelection {
  id: string;
  name: string;
  path: string;
}
