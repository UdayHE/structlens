import {
  Fragment,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ChangeEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
} from "react";
import {
  Binary,
  Braces,
  ChevronRight,
  ClipboardList,
  Hash,
  Layers3,
  Search,
  Type,
} from "lucide-react";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { useAppStore } from "@/store/app-store";
import { type SchemaTreeNode, type SchemaValueType } from "@/types/app";
import { RecordsView } from "@/features/records/records-view";

const ROW_HEIGHT = 46;
const OVERSCAN = 8;
const TREE_GRID_COLUMNS =
  "minmax(260px,1.8fr)_minmax(132px,0.9fr)_minmax(72px,0.5fr)_minmax(104px,0.7fr)_minmax(72px,0.45fr)";
const TREE_GRID_TEMPLATE = TREE_GRID_COLUMNS.replaceAll("_", " ");

interface FlatTreeNode {
  node: SchemaTreeNode;
  depth: number;
  hasChildren: boolean;
  parentId: string | null;
}

function compareNodes(left: SchemaTreeNode, right: SchemaTreeNode) {
  return left.name.localeCompare(right.name, undefined, {
    numeric: true,
    sensitivity: "base",
  });
}

function normalizeTree(nodes: SchemaTreeNode[]): SchemaTreeNode[] {
  return [...nodes]
    .sort(compareNodes)
    .map((node) => ({
      ...node,
      children: node.children ? normalizeTree(node.children) : undefined,
    }));
}

function flattenVisibleNodes(
  roots: SchemaTreeNode[],
  expandedIds: Set<string>,
  searchTerm: string,
) {
  const items: FlatTreeNode[] = [];
  const matches = new Set<string>();
  const loweredSearchTerm = searchTerm.trim().toLowerCase();

  function markMatches(node: SchemaTreeNode): boolean {
    const selfMatches =
      loweredSearchTerm.length === 0 ||
      node.name.toLowerCase().includes(loweredSearchTerm) ||
      node.path.toLowerCase().includes(loweredSearchTerm);
    let childMatches = false;
    for (const child of node.children ?? []) {
      if (markMatches(child)) childMatches = true;
    }
    const visible = selfMatches || childMatches;

    if (visible) {
      matches.add(node.id);
    }

    return visible;
  }

  roots.forEach(markMatches);

  const stack = [...roots]
    .reverse()
    .map((node) => ({ depth: 0, node, parentId: null as string | null }));

  while (stack.length > 0) {
    const current = stack.pop();
    if (!current) {
      continue;
    }

    const { depth, node, parentId } = current;
    if (!matches.has(node.id)) {
      continue;
    }

    const hasChildren = (node.children?.length ?? 0) > 0;
    items.push({
      node,
      depth,
      hasChildren,
      parentId,
    });

    const shouldExpand =
      loweredSearchTerm.length > 0 || (hasChildren && expandedIds.has(node.id));
    if (!shouldExpand || !node.children) {
      continue;
    }

    for (let index = node.children.length - 1; index >= 0; index -= 1) {
      stack.push({
        depth: depth + 1,
        node: node.children[index],
        parentId: node.id,
      });
    }
  }

  return items;
}

function buildParentLookup(roots: SchemaTreeNode[]) {
  const lookup = new Map<string, string | null>();
  const stack = roots.map((node) => ({ node, parentId: null as string | null }));

  while (stack.length > 0) {
    const current = stack.pop();
    if (!current) {
      continue;
    }

    lookup.set(current.node.id, current.parentId);
    const children = current.node.children ?? [];
    for (let index = children.length - 1; index >= 0; index -= 1) {
      stack.push({ node: children[index], parentId: current.node.id });
    }
  }

  return lookup;
}

function getTypeBadgeVariant(type: SchemaValueType) {
  if (type === "object" || type === "array") {
    return "accent";
  }
  return "muted";
}

function getTypeIcon(type: SchemaValueType) {
  if (type === "object" || type === "array") {
    return Braces;
  }
  if (type === "number") {
    return Hash;
  }
  if (type === "boolean") {
    return Binary;
  }
  return Type;
}

function renderHighlightedText(text: string, searchTerm: string) {
  const query = searchTerm.trim();
  if (!query) {
    return text;
  }

  const loweredText = text.toLowerCase();
  const loweredQuery = query.toLowerCase();
  const parts: ReactNode[] = [];
  let cursor = 0;
  let key = 0;

  while (cursor < text.length) {
    const matchIndex = loweredText.indexOf(loweredQuery, cursor);
    if (matchIndex === -1) {
      parts.push(<Fragment key={key}>{text.slice(cursor)}</Fragment>);
      break;
    }

    if (matchIndex > cursor) {
      parts.push(<Fragment key={key}>{text.slice(cursor, matchIndex)}</Fragment>);
      key += 1;
    }

    const matchEnd = matchIndex + query.length;
    parts.push(
      <mark
        key={key}
        className="rounded bg-accent/18 px-1 py-0.5 text-accent shadow-[0_0_0_1px_rgba(111,231,255,0.16)]"
      >
        {text.slice(matchIndex, matchEnd)}
      </mark>,
    );
    key += 1;
    cursor = matchEnd;
  }

  return parts;
}

function LoadingSkeleton() {
  return (
    <div className="space-y-2 p-4">
      {Array.from({ length: 12 }, (_, index) => (
        <div
          key={index}
          className="grid h-[46px] gap-3 rounded-2xl border border-white/5 bg-white/[0.03] px-4"
          style={{ gridTemplateColumns: TREE_GRID_TEMPLATE }}
        >
          <div className="my-auto h-3.5 w-2/3 animate-pulse rounded-full bg-white/10" />
          <div className="my-auto h-7 w-20 animate-pulse rounded-full bg-white/10" />
          <div className="my-auto h-3.5 w-10 animate-pulse rounded-full bg-white/10" />
          <div className="my-auto h-3.5 w-12 animate-pulse rounded-full bg-white/10" />
          <div className="my-auto h-3.5 w-8 animate-pulse rounded-full bg-white/10" />
        </div>
      ))}
    </div>
  );
}

export function SchemaTreeView({ isPrimary = false }: { isPrimary?: boolean }) {
  const {
    activeSchemaSelection,
    analysisError,
    analysisStage,
    generatedSQL,
    schemaMetadata,
    schemaTree,
    setActiveSchemaSelection,
    uploadedFile,
  } = useAppStore();
  const [activeTab, setActiveTab] = useState<"schema" | "records">("schema");
  const [search, setSearch] = useState("");
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [activeNodeId, setActiveNodeId] = useState("");
  const [scrollTop, setScrollTop] = useState(0);
  const [isTreeFocused, setIsTreeFocused] = useState(false);
  const containerRef = useRef<HTMLDivElement | null>(null);

  const treeData = useMemo(
    () => normalizeTree(schemaTree ?? []),
    [schemaTree],
  );

  useEffect(() => {
    if (treeData.length === 0) {
      setExpandedIds(new Set());
      setActiveNodeId("");
      setActiveSchemaSelection(null);
      return;
    }

    const nextExpandedIds = new Set<string>();
    nextExpandedIds.add(treeData[0].id);
    for (const child of treeData[0].children ?? []) {
      nextExpandedIds.add(child.id);
    }
    setExpandedIds(nextExpandedIds);
    setActiveNodeId(treeData[0].id);
    setActiveSchemaSelection({
      id: treeData[0].id,
      name: treeData[0].name,
      path: treeData[0].path,
    });
  }, [treeData]);

  const parentLookup = useMemo(() => buildParentLookup(treeData), [treeData]);
  const visibleNodes = useMemo(
    () => flattenVisibleNodes(treeData, expandedIds, search),
    [treeData, expandedIds, search],
  );

  useEffect(() => {
    if (visibleNodes.length === 0) {
      return;
    }

    const stillVisible = visibleNodes.some((item) => item.node.id === activeNodeId);
    if (!stillVisible) {
      setActiveNodeId(visibleNodes[0].node.id);
    }
  }, [activeNodeId, visibleNodes]);

  useEffect(() => {
    const selectedNode = visibleNodes.find((item) => item.node.id === activeNodeId)?.node;
    if (!selectedNode) {
      return;
    }
    if (
      activeSchemaSelection?.id === selectedNode.id &&
      activeSchemaSelection.path === selectedNode.path
    ) {
      return;
    }
    setActiveSchemaSelection({
      id: selectedNode.id,
      name: selectedNode.name,
      path: selectedNode.path,
    });
  }, [activeNodeId, activeSchemaSelection, setActiveSchemaSelection, visibleNodes]);

  const totalHeight = visibleNodes.length * ROW_HEIGHT;
  const viewportHeight = containerRef.current?.clientHeight || 520;
  const startIndex = Math.max(0, Math.floor(scrollTop / ROW_HEIGHT) - OVERSCAN);
  const endIndex = Math.min(
    visibleNodes.length,
    Math.ceil((scrollTop + viewportHeight) / ROW_HEIGHT) + OVERSCAN,
  );
  const virtualRows = visibleNodes.slice(startIndex, endIndex);
  const topSpacer = startIndex * ROW_HEIGHT;
  const activeIndex = visibleNodes.findIndex((item) => item.node.id === activeNodeId);
  const activeNode = activeIndex >= 0 ? visibleNodes[activeIndex] : null;

  const toggleNode = (nodeId: string) => {
    setExpandedIds((current) => {
      const next = new Set(current);
      if (next.has(nodeId)) {
        next.delete(nodeId);
      } else {
        next.add(nodeId);
      }
      return next;
    });
  };

  const scrollToIndex = (index: number) => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const itemTop = index * ROW_HEIGHT;
    const itemBottom = itemTop + ROW_HEIGHT;
    const viewTop = container.scrollTop;
    const viewBottom = viewTop + container.clientHeight;

    if (itemTop < viewTop) {
      container.scrollTo({ top: itemTop, behavior: "smooth" });
      return;
    }
    if (itemBottom > viewBottom) {
      container.scrollTo({
        top: itemBottom - container.clientHeight,
        behavior: "smooth",
      });
    }
  };

  const setActiveByIndex = (index: number) => {
    if (index < 0 || index >= visibleNodes.length) {
      return;
    }
    setActiveNodeId(visibleNodes[index].node.id);
    scrollToIndex(index);
  };

  const handleKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
    if (visibleNodes.length === 0) {
      return;
    }

    const currentItem = activeNode ?? visibleNodes[0];
    const currentIndex = activeNode ? activeIndex : 0;

    if (event.key === "ArrowDown") {
      event.preventDefault();
      setActiveByIndex(Math.min(visibleNodes.length - 1, currentIndex + 1));
      return;
    }
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setActiveByIndex(Math.max(0, currentIndex - 1));
      return;
    }
    if (event.key === "Home") {
      event.preventDefault();
      setActiveByIndex(0);
      return;
    }
    if (event.key === "End") {
      event.preventDefault();
      setActiveByIndex(visibleNodes.length - 1);
      return;
    }
    if (event.key === "ArrowRight") {
      event.preventDefault();
      if (currentItem.hasChildren && !expandedIds.has(currentItem.node.id)) {
        toggleNode(currentItem.node.id);
      }
      return;
    }
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      if (currentItem.hasChildren && expandedIds.has(currentItem.node.id)) {
        toggleNode(currentItem.node.id);
        return;
      }
      const parentId = parentLookup.get(currentItem.node.id);
      if (!parentId) {
        return;
      }
      const parentIndex = visibleNodes.findIndex((item) => item.node.id === parentId);
      setActiveByIndex(parentIndex);
    }
  };

  const handleSearchChange = (event: ChangeEvent<HTMLInputElement>) => {
    setSearch(event.target.value);
    setScrollTop(0);
    if (containerRef.current) {
      containerRef.current.scrollTop = 0;
    }
  };

  const isLoading = analysisStage !== "idle" && analysisStage !== "ready" && !analysisError;

  return (
      <Card
        className={cn(
          "flex h-full min-h-[640px] flex-col overflow-hidden border-white/8",
          isPrimary && "shadow-[0_0_0_1px_rgba(111,231,255,0.16)]",
        )}
      >
        <CardContent className="flex min-h-0 flex-1 flex-col gap-5 p-5">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div className="space-y-2">
              <Badge variant="accent" className="w-fit">
                <Layers3 className="mr-2 h-3.5 w-3.5" />
                Interactive Tree
              </Badge>
              <p className="text-sm text-muted-foreground">
                Searchable schema explorer with virtualization and keyboard navigation.
              </p>
            </div>

            <div className="grid gap-3 text-sm text-muted-foreground sm:grid-cols-3">
              <div className="rounded-2xl border border-white/8 bg-white/4 px-4 py-3">
                <p className="font-medium leading-tight text-foreground">
                  {schemaMetadata?.totalFields ?? 0}
                </p>
                <p>Total fields</p>
              </div>
              <div className="rounded-2xl border border-white/8 bg-white/4 px-4 py-3">
                <p className="font-medium leading-tight text-foreground">
                  {schemaMetadata?.tableCount ?? 0}
                </p>
                <p className="leading-tight">Generated tables</p>
              </div>
              <div className="rounded-2xl border border-white/8 bg-white/4 px-4 py-3">
                <p className="truncate font-medium leading-tight text-foreground">
                  {uploadedFile?.name ?? "No source"}
                </p>
                <p className="leading-tight">Current source</p>
              </div>
            </div>
          </div>

          {/* Tab switcher */}
          <div className="flex items-center gap-1 rounded-2xl border border-white/8 bg-white/[0.025] p-1 w-fit">
            <button
              type="button"
              onClick={() => setActiveTab("schema")}
              className={cn(
                "flex items-center gap-1.5 rounded-xl px-4 py-1.5 text-xs font-medium transition-colors duration-150",
                activeTab === "schema"
                  ? "bg-accent/15 text-accent shadow-[inset_0_0_0_1px_rgba(111,231,255,0.2)]"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <Layers3 className="h-3 w-3" />
              Schema
            </button>
            <button
              type="button"
              onClick={() => setActiveTab("records")}
              className={cn(
                "flex items-center gap-1.5 rounded-xl px-4 py-1.5 text-xs font-medium transition-colors duration-150",
                activeTab === "records"
                  ? "bg-accent/15 text-accent shadow-[inset_0_0_0_1px_rgba(111,231,255,0.2)]"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              <ClipboardList className="h-3 w-3" />
              Records
            </button>
          </div>

          {activeTab === "records" ? (
            <div className="min-h-0 flex-1 overflow-hidden rounded-[26px] border border-white/8 bg-black/12 p-3">
              <RecordsView />
            </div>
          ) : (
            <>
          <div className="sticky top-0 z-10 flex flex-col gap-3 bg-[linear-gradient(180deg,rgba(6,19,31,0.96),rgba(6,19,31,0.88))] pb-2 backdrop-blur-xl lg:flex-row lg:items-center lg:justify-between">
            <label className="flex w-full max-w-xl items-center gap-3 rounded-2xl border border-white/8 bg-white/4 px-4 py-3 shadow-[0_12px_24px_rgba(0,0,0,0.14)]">
              <Search className="h-4 w-4 text-muted-foreground" />
              <input
                value={search}
                onChange={handleSearchChange}
                placeholder="Filter by field or path"
                className="w-full bg-transparent text-sm text-foreground outline-none placeholder:text-muted-foreground"
              />
            </label>
            <div className="flex items-center gap-2 text-xs uppercase tracking-[0.16em] text-muted-foreground">
              <span className="rounded-full border border-white/8 px-3 py-1">Sticky search</span>
              <span className="rounded-full border border-white/8 px-3 py-1">Visible rows only</span>
            </div>
          </div>

          <div className="grid min-h-0 flex-1 gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
            <div className="min-h-0 overflow-hidden rounded-[26px] border border-white/8 bg-black/12">
              <div className="flex h-full flex-col overflow-x-auto">
                <div
                  className="grid min-w-[700px] shrink-0 gap-3 border-b border-white/6 px-4 py-3 font-mono text-[11px] uppercase tracking-[0.14em] text-muted-foreground"
                  style={{ gridTemplateColumns: TREE_GRID_TEMPLATE }}
                >
                <span>Field</span>
                <span>Resolved Type</span>
                <span>Array</span>
                <span>Optional</span>
                <span>Children</span>
                </div>

              {isLoading ? (
                <LoadingSkeleton />
              ) : (
                <div
                  ref={containerRef}
                  tabIndex={0}
                  role="tree"
                  aria-label="Schema tree"
                  onKeyDown={handleKeyDown}
                  onFocus={() => setIsTreeFocused(true)}
                  onBlur={() => setIsTreeFocused(false)}
                  onScroll={(event) => setScrollTop(event.currentTarget.scrollTop)}
                  className="min-h-0 flex-1 overflow-y-auto scroll-smooth outline-none"
                >
                  {analysisError ? (
                    <div className="flex min-h-[420px] items-center justify-center p-6 text-center text-sm text-destructive">
                      {analysisError}
                    </div>
                  ) : visibleNodes.length === 0 ? (
                    <div className="flex min-h-[420px] items-center justify-center p-6 text-center text-sm text-muted-foreground">
                      No matching fields found
                    </div>
                  ) : (
                    <div
                      className="min-w-[700px]"
                      style={{ height: totalHeight, position: "relative" }}
                    >
                      <div
                        style={{
                          transform: `translateY(${topSpacer}px)`,
                        }}
                      >
                        {virtualRows.map((item) => {
                          const isExpanded = expandedIds.has(item.node.id);
                          const isActive = item.node.id === activeNodeId;
                          const Icon = getTypeIcon(item.node.type);

                          return (
                            <button
                              key={item.node.id}
                              type="button"
                              role="treeitem"
                              aria-level={item.depth + 1}
                              aria-expanded={item.hasChildren ? isExpanded : undefined}
                              aria-selected={isActive}
                              onClick={() => setActiveNodeId(item.node.id)}
                              onDoubleClick={() => {
                                if (item.hasChildren) {
                                  toggleNode(item.node.id);
                                }
                              }}
                              className={cn(
                                "tree-row-appear grid h-[46px] w-full items-center gap-3 border-b border-white/5 px-4 font-mono text-sm transition-[background-color,box-shadow,border-color] duration-150",
                                isActive
                                  ? "border-white/10 bg-accent/10 shadow-[inset_0_0_0_1px_rgba(111,231,255,0.18)]"
                                  : "hover:bg-white/4",
                                isActive &&
                                  isTreeFocused &&
                                  "shadow-[inset_0_0_0_1px_rgba(111,231,255,0.22),0_0_0_1px_rgba(111,231,255,0.14)]",
                              )}
                              style={{ gridTemplateColumns: TREE_GRID_TEMPLATE }}
                            >
                              <span
                                className="flex min-w-0 items-center gap-2"
                                style={{ paddingLeft: `${item.depth * 18}px` }}
                              >
                                <span
                                  className={cn(
                                    "flex h-6 w-6 shrink-0 items-center justify-center rounded-md border transition-[transform,border-color,color,background-color] duration-150",
                                    item.hasChildren
                                      ? "border-white/10 bg-white/4 text-muted-foreground hover:border-accent/40 hover:text-accent"
                                      : "border-transparent bg-transparent text-transparent",
                                  )}
                                  onClick={(event) => {
                                    event.stopPropagation();
                                    if (item.hasChildren) {
                                      toggleNode(item.node.id);
                                    }
                                  }}
                                >
                                  {item.hasChildren ? (
                                    <ChevronRight
                                      className={cn(
                                        "h-3.5 w-3.5 transition-transform duration-150",
                                        isExpanded && "rotate-90",
                                      )}
                                    />
                                  ) : (
                                    <ChevronRight className="h-3.5 w-3.5" />
                                  )}
                                </span>
                                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg border border-white/8 bg-white/4 text-muted-foreground">
                                  <Icon className="h-3.5 w-3.5" />
                                </span>
                                <span
                                  className="truncate text-left text-foreground"
                                  title={item.node.path}
                                >
                                  {renderHighlightedText(item.node.name, search)}
                                </span>
                                <span className="sr-only">
                                  {renderHighlightedText(item.node.path, search)}
                                </span>
                              </span>

                              <span className="flex justify-start">
                                <Badge
                                  variant={getTypeBadgeVariant(item.node.type)}
                                  className="min-w-[88px] justify-center px-2.5 tracking-[0.12em]"
                                >
                                  {item.node.type}
                                </Badge>
                              </span>
                              <span className="flex justify-start">
                                {item.node.isArray ? (
                                  <Badge
                                    variant="accent"
                                    className="min-w-[56px] justify-center px-2.5 tracking-[0.12em]"
                                  >
                                    []
                                  </Badge>
                                ) : (
                                  <span className="text-muted-foreground">no</span>
                                )}
                              </span>
                              <span className="flex justify-start">
                                {item.node.optional ? (
                                  <Badge
                                    variant="muted"
                                    className="min-w-[74px] justify-center px-2.5 tracking-[0.12em]"
                                  >
                                    optional
                                  </Badge>
                                ) : (
                                  <span className="text-muted-foreground">required</span>
                                )}
                              </span>
                              <span className="text-muted-foreground">
                                {item.node.childCount}
                              </span>
                            </button>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </div>
              )}
              </div>
            </div>

            <aside className="flex min-h-0 flex-col gap-4 rounded-[26px] border border-white/8 bg-black/12 p-4">
              <div>
                <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
                  Active Node
                </p>
                {activeNode ? (
                  <div className="mt-4 space-y-4">
                    <div>
                      <p className="text-lg font-semibold text-foreground">
                        {activeNode.node.name}
                      </p>
                      <p className="mt-1 break-all font-mono text-xs text-muted-foreground">
                        {activeNode.node.path}
                      </p>
                    </div>

                    <div className="flex flex-wrap gap-2">
                      <Badge variant={getTypeBadgeVariant(activeNode.node.type)}>
                        {activeNode.node.type}
                      </Badge>
                      {activeNode.node.isArray ? (
                        <Badge variant="accent">array</Badge>
                      ) : null}
                      {activeNode.node.optional ? (
                        <Badge variant="muted">optional</Badge>
                      ) : (
                        <Badge variant="success">required</Badge>
                      )}
                    </div>

                    <div className="space-y-2 rounded-2xl border border-white/8 bg-white/4 p-4 text-sm text-muted-foreground">
                      <p>
                        Child nodes:{" "}
                        <span className="text-foreground">{activeNode.node.childCount}</span>
                      </p>
                      <p>
                        Depth: <span className="text-foreground">{activeNode.depth}</span>
                      </p>
                      <p>
                        Parent:{" "}
                        <span className="text-foreground">
                          {activeNode.parentId ?? "root"}
                        </span>
                      </p>
                    </div>
                  </div>
                ) : (
                  <p className="mt-4 text-sm text-muted-foreground">
                    Select a field to inspect its schema metadata.
                  </p>
                )}
              </div>

              <div className="rounded-2xl border border-white/8 bg-white/4 p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
                  Analysis
                </p>
                <div className="mt-3 space-y-2 text-sm text-muted-foreground">
                  <p>
                    Arrays:{" "}
                    <span className="text-foreground">
                      {schemaMetadata?.arrayFields ?? 0}
                    </span>
                  </p>
                  <p>
                    Optional fields:{" "}
                    <span className="text-foreground">
                      {schemaMetadata?.optionalFields ?? 0}
                    </span>
                  </p>
                  <p>
                    SQL ready:{" "}
                    <span className="text-foreground">
                      {generatedSQL ? "yes" : "no"}
                    </span>
                  </p>
                </div>
              </div>

              <div className="rounded-2xl border border-white/8 bg-white/4 p-4">
                <p className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted-foreground">
                  Controls
                </p>
                <div className="mt-3 space-y-2 text-sm text-muted-foreground">
                  <p>
                    <span className="text-foreground">Up / Down</span> moves focus
                  </p>
                  <p>
                    <span className="text-foreground">Right</span> expands a group
                  </p>
                  <p>
                    <span className="text-foreground">Left</span> collapses or jumps to parent
                  </p>
                  <p>
                    <span className="text-foreground">Home / End</span> jumps through visible nodes
                  </p>
                </div>
              </div>
            </aside>
          </div>
            </>
          )}
        </CardContent>
      </Card>
  );
}
