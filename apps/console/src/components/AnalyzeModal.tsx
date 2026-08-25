import { useEffect } from "react";
import { useQuery } from "@tanstack/react-query";
import { X } from "lucide-react";
import { api } from "../lib/api";
import { Empty, Spinner } from "./ui";

const verdictColor: Record<string, string> = {
  ship: "var(--color-brand)",
  hold: "var(--color-danger)",
  iterate: "var(--color-off)",
};

export function AnalyzeModal({ flagKey, onClose }: { flagKey: string; onClose: () => void }) {
  const q = useQuery({ queryKey: ["analyze", flagKey], queryFn: () => api.analyze(flagKey), retry: false });

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const a = q.data;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={onClose}>
      <div
        role="dialog"
        aria-modal="true"
        aria-label="Rollout analysis"
        className="w-full max-w-md rounded-xl border border-line bg-surface p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-mono text-sm font-semibold">Analysis · {flagKey}</h2>
          <button onClick={onClose} aria-label="Close" className="rounded p-1 text-muted hover:text-ink">
            <X size={16} />
          </button>
        </div>

        {q.isLoading ? (
          <Spinner label="Analyzing…" />
        ) : q.error ? (
          <Empty>{(q.error as Error).message}</Empty>
        ) : a ? (
          <div className="flex flex-col gap-4">
            <div>
              <div className="font-mono text-[11px] uppercase tracking-wide text-muted">Verdict</div>
              <div className="mt-1 font-mono text-2xl font-semibold capitalize" style={{ color: verdictColor[a.verdict] }}>
                {a.verdict}
              </div>
            </div>
            <p className="text-sm leading-relaxed">{a.summary}</p>

            {a.stats && (
              <dl className="grid grid-cols-2 gap-2 rounded-md border border-line p-3 font-mono text-xs">
                <div><dt className="text-muted">control</dt><dd>{(a.stats.control_rate * 100).toFixed(1)}%</dd></div>
                <div><dt className="text-muted">treatment</dt><dd>{(a.stats.treatment_rate * 100).toFixed(1)}%</dd></div>
                <div><dt className="text-muted">rel. lift</dt><dd>{(a.stats.relative_lift * 100).toFixed(0)}%</dd></div>
                <div><dt className="text-muted">p-value</dt><dd>{a.stats.p_value.toExponential(2)}</dd></div>
              </dl>
            )}

            {a.risks && a.risks.length > 0 && (
              <div>
                <div className="mb-1 font-mono text-[11px] uppercase tracking-wide text-muted">Risks</div>
                <ul className="flex flex-col gap-1">
                  {a.risks.map((r) => (
                    <li key={r} className="flex items-start gap-2 text-sm">
                      <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full" style={{ background: "var(--color-danger)" }} />
                      {r}
                    </li>
                  ))}
                </ul>
              </div>
            )}

            <p className="text-[11px] text-muted">{a.model ? `Reasoned by ${a.model}.` : "Rule-based (no model configured)."}</p>
          </div>
        ) : null}
      </div>
    </div>
  );
}
