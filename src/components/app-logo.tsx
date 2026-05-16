import { ScanSearch } from "lucide-react";

export function AppLogo() {
  return (
    <div className="flex items-center gap-3">
      <div className="flex h-11 w-11 items-center justify-center rounded-2xl border border-accent/30 bg-accent/12 text-accent shadow-[0_0_28px_rgba(111,231,255,0.18)]">
        <ScanSearch className="h-5 w-5" />
      </div>
      <div>
        <p className="text-[11px] uppercase tracking-[0.3em] text-accent/75">
          Desktop Studio
        </p>
        <h1 className="text-lg font-semibold tracking-[0.08em]">StructLens</h1>
      </div>
    </div>
  );
}
