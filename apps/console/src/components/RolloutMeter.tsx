// A segmented level meter for the rollout percentage — reads like a channel
// strip on a mixing board. Dimmed when the flag is off.
const SEGMENTS = 10;

export function RolloutMeter({ pct, active }: { pct: number; active: boolean }) {
  const lit = Math.round((pct / 100) * SEGMENTS);
  return (
    <div className="flex items-center gap-2" title={`${pct}% rollout`}>
      <div className="flex gap-0.5" aria-hidden>
        {Array.from({ length: SEGMENTS }).map((_, i) => (
          <span
            key={i}
            className="h-4 w-1 rounded-sm"
            style={{
              background:
                i < lit && active ? "var(--color-brand)" : "var(--color-line)",
            }}
          />
        ))}
      </div>
      <span className="w-9 text-right font-mono text-xs tabular-nums text-muted">{pct}%</span>
    </div>
  );
}
