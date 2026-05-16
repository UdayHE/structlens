import { ArrowRight, Sparkles } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useAppStore } from "@/store/app-store";
import { type SidebarSection } from "@/types/app";

const panelCopy: Record<
  Exclude<SidebarSection, "files">,
  { title: string; description: string; cta: string }
> = {
  tree: {
    title: "Tree view will appear here",
    description:
      "Inspect nested nodes, repeated collections, and object boundaries once a file is imported.",
    cta: "Upload a file to unlock the structural tree.",
  },
  schema: {
    title: "Schema inference preview",
    description:
      "Optional fields, type conflicts, and flattening decisions will be surfaced in this panel.",
    cta: "Bring in XML or JSON to start inference.",
  },
  sql: {
    title: "Relational SQL output",
    description:
      "Generated tables, keys, and child relationships will render in a focused SQL workspace.",
    cta: "Import data to prepare a relational mapping.",
  },
  settings: {
    title: "Workspace settings",
    description:
      "Flattening thresholds, naming rules, and future export controls can live here without crowding the main flow.",
    cta: "This area is ready for configuration controls.",
  },
};

export function EmptyPanel({ section }: { section: Exclude<SidebarSection, "files"> }) {
  const { setActiveSection } = useAppStore();
  const copy = panelCopy[section];

  return (
    <div className="flex h-full items-center justify-center p-6">
      <Card className="max-w-2xl border-white/8">
        <CardHeader>
          <Badge variant="accent" className="w-fit">
            <Sparkles className="mr-2 h-3.5 w-3.5" />
            Workspace Placeholder
          </Badge>
          <CardTitle className="text-2xl">{copy.title}</CardTitle>
          <CardDescription>{copy.description}</CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-5">
          <p className="text-sm text-muted-foreground">{copy.cta}</p>
          <div>
            <Button variant="outline" onClick={() => setActiveSection("files")}>
              Open Files
              <ArrowRight className="h-4 w-4" />
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
