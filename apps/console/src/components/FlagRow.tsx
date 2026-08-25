import { Pencil, Trash2 } from "lucide-react";
import type { Flag } from "../lib/types";
import { ago } from "../lib/format";
import { Toggle } from "./Toggle";
import { RolloutMeter } from "./RolloutMeter";
import { Tag } from "./ui";

export function FlagRow({
  flag,
  busy,
  onToggle,
  onEdit,
  onDelete,
}: {
  flag: Flag;
  busy?: boolean;
  onToggle: (next: boolean) => void;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="group grid grid-cols-[auto_1fr_auto] items-center gap-4 border-b border-line px-4 py-3 last:border-b-0 hover:bg-surface2/60 sm:grid-cols-[auto_2fr_1.5fr_auto]">
      {/* status node */}
      <span
        className={`h-2.5 w-2.5 rounded-full ${flag.enabled ? "bg-brand broadcast" : ""}`}
        style={flag.enabled ? undefined : { background: "var(--color-off)" }}
        aria-hidden
      />

      {/* identity */}
      <div className="min-w-0">
        <div className="truncate font-mono text-sm font-semibold text-ink">{flag.key}</div>
        <div className="flex items-center gap-2">
          <span className="truncate text-xs text-muted">{flag.name}</span>
          {flag.tags?.map((t) => <Tag key={t}>{t}</Tag>)}
        </div>
      </div>

      {/* rollout — hidden on the narrowest layout */}
      <div className="hidden sm:block">
        <RolloutMeter pct={flag.rollout} active={flag.enabled} />
      </div>

      {/* controls */}
      <div className="flex items-center gap-3 justify-self-end">
        <span className="hidden font-mono text-[11px] text-muted md:inline">{ago(flag.updated_at)}</span>
        <button
          onClick={onEdit}
          aria-label={`Edit ${flag.key}`}
          className="rounded p-1 text-muted opacity-0 transition hover:text-ink focus-visible:opacity-100 group-hover:opacity-100"
        >
          <Pencil size={15} />
        </button>
        <button
          onClick={onDelete}
          aria-label={`Delete ${flag.key}`}
          className="rounded p-1 text-muted opacity-0 transition hover:text-danger focus-visible:opacity-100 group-hover:opacity-100"
        >
          <Trash2 size={15} />
        </button>
        <Toggle on={flag.enabled} disabled={busy} onChange={onToggle} label={`Toggle ${flag.key}`} />
      </div>
    </div>
  );
}
