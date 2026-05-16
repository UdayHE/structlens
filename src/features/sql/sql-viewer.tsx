import { useEffect, useMemo, useRef, useState } from "react";
import { Check, ClipboardCopy, Download, FileCode2 } from "lucide-react";
import { save } from "@tauri-apps/api/dialog";
import { writeTextFile } from "@tauri-apps/api/fs";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";
import { type ActiveSchemaSelection } from "@/types/app";

type SqlTokenKind =
  | "keyword"
  | "identifier"
  | "string"
  | "number"
  | "punctuation"
  | "comment"
  | "operator"
  | "whitespace";

interface SqlToken {
  kind: SqlTokenKind;
  value: string;
}

const sqlKeywords = new Set([
  "create",
  "table",
  "primary",
  "key",
  "foreign",
  "references",
  "text",
  "bigint",
  "double",
  "precision",
  "boolean",
  "null",
  "constraint",
]);

function tokenizeSqlLine(line: string): SqlToken[] {
  const tokens: SqlToken[] = [];
  let index = 0;

  while (index < line.length) {
    const current = line[index];
    const rest = line.slice(index);

    if (rest.startsWith("--")) {
      tokens.push({ kind: "comment", value: rest });
      break;
    }

    if (/\s/.test(current)) {
      let end = index + 1;
      while (end < line.length && /\s/.test(line[end])) {
        end += 1;
      }
      tokens.push({ kind: "whitespace", value: line.slice(index, end) });
      index = end;
      continue;
    }

    if (current === "'") {
      let end = index + 1;
      while (end < line.length) {
        if (line[end] === "'" && line[end - 1] !== "\\") {
          end += 1;
          break;
        }
        end += 1;
      }
      tokens.push({ kind: "string", value: line.slice(index, end) });
      index = end;
      continue;
    }

    if (/[(),.;]/.test(current)) {
      tokens.push({ kind: "punctuation", value: current });
      index += 1;
      continue;
    }

    if (/[=<>*+-]/.test(current)) {
      tokens.push({ kind: "operator", value: current });
      index += 1;
      continue;
    }

    if (/\d/.test(current)) {
      let end = index + 1;
      while (end < line.length && /[\d.]/.test(line[end])) {
        end += 1;
      }
      tokens.push({ kind: "number", value: line.slice(index, end) });
      index = end;
      continue;
    }

    let end = index + 1;
    while (end < line.length && /[A-Za-z0-9_]/.test(line[end])) {
      end += 1;
    }
    const word = line.slice(index, end);
    tokens.push({
      kind: sqlKeywords.has(word.toLowerCase()) ? "keyword" : "identifier",
      value: word,
    });
    index = end;
  }

  return tokens;
}

function tokenClassName(kind: SqlTokenKind) {
  switch (kind) {
    case "keyword":
      return "text-accent";
    case "string":
      return "text-success";
    case "number":
      return "text-warning";
    case "comment":
      return "text-muted-foreground/70 italic";
    case "operator":
      return "text-[#ffd6a5]";
    case "punctuation":
      return "text-foreground/85";
    case "identifier":
      return "text-foreground";
    default:
      return "";
  }
}

function toSnakeCase(value: string) {
  return value
    .replace(/\[\]/g, "")
    .replace(/([a-z0-9])([A-Z])/g, "$1_$2")
    .replace(/[^a-zA-Z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .toLowerCase();
}

function selectionTokens(selection: ActiveSchemaSelection | null) {
  if (!selection) {
    return [];
  }

  const rawSegments = selection.path
    .replace(/\[\]/g, "")
    .split(".")
    .filter((segment) => segment && segment !== "root");
  const normalizedSegments = rawSegments
    .map((segment) => toSnakeCase(segment))
    .filter(Boolean);

  const tokens = new Set<string>();
  const nameToken = toSnakeCase(selection.name);
  if (nameToken) {
    tokens.add(nameToken);
  }

  for (let index = 0; index < normalizedSegments.length; index += 1) {
    const segment = normalizedSegments[index];
    tokens.add(segment);

    const next = normalizedSegments[index + 1];
    if (next) {
      tokens.add(`${segment}_${next}`);
    }
  }

  const lastTwo = normalizedSegments.slice(-2);
  if (lastTwo.length === 2) {
    tokens.add(lastTwo.join("_"));
  }

  return [...tokens].filter((token) => token.length > 1);
}

function LoadingSqlSkeleton() {
  return (
    <div className="space-y-2 p-4">
      {Array.from({ length: 12 }, (_, index) => (
        <div
          key={index}
          className="flex items-center gap-4 rounded-2xl border border-white/5 bg-white/[0.03] px-4 py-3"
        >
          <div className="h-4 w-8 animate-pulse rounded-full bg-white/10" />
          <div className="h-3.5 flex-1 animate-pulse rounded-full bg-white/10" />
        </div>
      ))}
    </div>
  );
}

export function SqlViewer({ isPrimary }: { isPrimary: boolean }) {
  const { activeSchemaSelection, analysisError, analysisStage, generatedSQL, uploadedFile } =
    useAppStore();
  const [copyState, setCopyState] = useState<"idle" | "copied">("idle");
  const [downloadState, setDownloadState] = useState<"idle" | "saved">("idle");
  const containerRef = useRef<HTMLDivElement | null>(null);
  const rowRefs = useRef<Map<number, HTMLDivElement>>(new Map());

  const sqlLines = useMemo(
    () => generatedSQL.split("\n").map((line) => tokenizeSqlLine(line)),
    [generatedSQL],
  );
  const relatedTokens = useMemo(
    () => selectionTokens(activeSchemaSelection),
    [activeSchemaSelection],
  );
  const highlightedLineIndices = useMemo(
    () =>
      generatedSQL
        .split("\n")
        .map((line, index) =>
          relatedTokens.some((token) => line.toLowerCase().includes(token)) ? index : -1,
        )
        .filter((index) => index >= 0),
    [generatedSQL, relatedTokens],
  );

  useEffect(() => {
    const firstIndex = highlightedLineIndices[0];
    if (firstIndex === undefined) {
      return;
    }
    const container = containerRef.current;
    const row = rowRefs.current.get(firstIndex);
    if (!container || !row) {
      return;
    }
    row.scrollIntoView({ block: "nearest", behavior: "smooth" });
  }, [highlightedLineIndices]);

  const handleCopy = async () => {
    if (!generatedSQL) {
      return;
    }
    await navigator.clipboard.writeText(generatedSQL);
    setCopyState("copied");
    window.setTimeout(() => setCopyState("idle"), 1400);
  };

  const handleDownload = async () => {
    if (!generatedSQL) {
      return;
    }

    const defaultName = `${uploadedFile?.name.replace(/\.[^.]+$/, "") || "structlens"}.sql`;
    const targetPath = await save({
      defaultPath: defaultName,
      filters: [{ name: "SQL", extensions: ["sql"] }],
    });

    if (!targetPath) {
      return;
    }

    await writeTextFile(targetPath, generatedSQL);
    setDownloadState("saved");
    window.setTimeout(() => setDownloadState("idle"), 1400);
  };

  const isLoading = analysisStage !== "idle" && analysisStage !== "ready" && !analysisError;

  return (
    <Card
      className={cn(
        "flex h-full min-h-0 flex-col overflow-hidden border-white/8",
        isPrimary && "shadow-[0_0_0_1px_rgba(111,231,255,0.16)]",
      )}
    >
      <CardHeader className="border-b border-white/6">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="space-y-2">
            <Badge variant="accent" className="w-fit">
              <FileCode2 className="mr-2 h-3.5 w-3.5" />
              SQL Output
            </Badge>
            <CardTitle className="text-xl">Relational DDL Viewer</CardTitle>
            <CardDescription>
              Editor-style SQL preview with line numbers, syntax emphasis, and export controls.
            </CardDescription>
          </div>

          <div className="flex flex-wrap items-center gap-3">
            <Button variant="outline" onClick={() => void handleCopy()} disabled={!generatedSQL}>
              {copyState === "copied" ? <Check className="h-4 w-4" /> : <ClipboardCopy className="h-4 w-4" />}
              {copyState === "copied" ? "Copied" : "Copy SQL"}
            </Button>
            <Button variant="outline" onClick={() => void handleDownload()} disabled={!generatedSQL}>
              {downloadState === "saved" ? <Check className="h-4 w-4" /> : <Download className="h-4 w-4" />}
              {downloadState === "saved" ? "Saved" : "Download"}
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex min-h-0 flex-1 flex-col p-0">
        {isLoading ? (
          <LoadingSqlSkeleton />
        ) : analysisError ? (
          <div className="flex min-h-[420px] items-center justify-center p-6 text-center text-sm text-destructive">
            {analysisError}
          </div>
        ) : !generatedSQL ? (
          <div className="flex min-h-[420px] items-center justify-center p-6 text-center text-sm text-muted-foreground">
            Upload a file to generate SQL output.
          </div>
        ) : (
          <div
            ref={containerRef}
            className="min-h-0 flex-1 overflow-auto rounded-b-[28px] bg-[#04101a]"
          >
            <div className="min-w-max font-mono text-sm">
              {sqlLines.map((line, index) => (
                <div
                  key={index}
                  ref={(node) => {
                    if (!node) {
                      rowRefs.current.delete(index);
                      return;
                    }
                    rowRefs.current.set(index, node);
                  }}
                  className={cn(
                    "grid grid-cols-[64px_minmax(0,1fr)] border-b border-white/[0.04] hover:bg-white/[0.03]",
                    highlightedLineIndices.includes(index) &&
                      "bg-accent/10 shadow-[inset_0_0_0_1px_rgba(111,231,255,0.12)]",
                  )}
                >
                  <div className="select-none border-r border-white/[0.05] px-4 py-2 text-right text-xs text-muted-foreground">
                    {index + 1}
                  </div>
                  <pre className="overflow-x-auto px-4 py-2 whitespace-pre-wrap break-words text-sm leading-6">
                    {line.length === 0 ? (
                      <span className="text-muted-foreground/40"> </span>
                    ) : (
                      line.map((token, tokenIndex) => (
                        <span
                          key={`${index}-${tokenIndex}`}
                          className={tokenClassName(token.kind)}
                        >
                          {token.value}
                        </span>
                      ))
                    )}
                  </pre>
                </div>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
