import { useAudit } from "../lib/hooks";
import { ago } from "../lib/format";
import { Empty, Spinner } from "../components/ui";

const actionColor: Record<string, string> = {
  created: "var(--color-brand)",
  updated: "var(--color-muted)",
  deleted: "var(--color-danger)",
};

export function Audit() {
  const q = useAudit();
  const entries = q.data?.audit ?? [];
  const err = q.error as Error | null;

  return (
    <div className="mx-auto max-w-4xl px-6 py-6">
      <h1 className="mb-1 font-mono text-xl font-semibold tracking-tight">Audit</h1>
      <p className="mb-5 text-sm text-muted">Every flag change, newest first.</p>

      {err && <p className="mb-3 text-sm text-danger">Couldn't load audit: {err.message}</p>}

      {q.isLoading ? (
        <Spinner label="Loading audit…" />
      ) : entries.length === 0 ? (
        <Empty>No changes recorded yet.</Empty>
      ) : (
        <div className="overflow-hidden rounded-xl border border-line bg-surface">
          {entries.map((e, i) => (
            <div
              key={i}
              className="grid grid-cols-[auto_1fr_auto] items-center gap-4 border-b border-line px-4 py-2.5 last:border-b-0"
            >
              <span
                className="font-mono text-[11px] uppercase tracking-wide"
                style={{ color: actionColor[e.action] ?? "var(--color-muted)" }}
              >
                {e.action}
              </span>
              <span className="truncate font-mono text-sm">
                {e.flag_key} <span className="text-muted">· {e.actor}</span>
              </span>
              <span className="font-mono text-[11px] text-muted">{ago(e.timestamp)}</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
