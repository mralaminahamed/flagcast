// The switchboard toggle. When on, the knob broadcasts a pulse ring — the
// signature that flagcast is propagating the flag to evaluators.
export function Toggle({
  on,
  onChange,
  disabled,
  label,
}: {
  on: boolean;
  onChange: (next: boolean) => void;
  disabled?: boolean;
  label: string;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={label}
      disabled={disabled}
      onClick={() => onChange(!on)}
      className={`relative inline-flex h-6 w-11 shrink-0 items-center rounded-full border transition disabled:opacity-50 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand ${
        on ? "border-brand bg-brand" : "border-line bg-surface2"
      }`}
    >
      <span
        className={`inline-block h-4 w-4 rounded-full bg-white shadow transition-transform ${
          on ? "translate-x-6 broadcast" : "translate-x-1"
        }`}
        style={on ? undefined : { background: "var(--color-off)" }}
      />
    </button>
  );
}
