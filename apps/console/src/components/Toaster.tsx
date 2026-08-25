import { useToasts } from "../lib/toast";

export function Toaster() {
  const toasts = useToasts((s) => s.toasts);
  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((t) => (
        <div
          key={t.id}
          className="rounded-md border bg-surface px-3 py-2 text-sm shadow-lg"
          style={{ borderColor: t.tone === "error" ? "var(--color-danger)" : "var(--color-brand)" }}
        >
          {t.title}
        </div>
      ))}
    </div>
  );
}
