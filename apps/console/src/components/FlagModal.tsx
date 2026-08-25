import { useEffect, useState } from "react";
import { X } from "lucide-react";
import type { Flag, FlagInput } from "../lib/types";
import { Button, Field, Input } from "./ui";
import { Toggle } from "./Toggle";

export function FlagModal({
  flag,
  pending,
  error,
  onSubmit,
  onClose,
}: {
  flag: Flag | null; // null = create
  pending: boolean;
  error: string | null;
  onSubmit: (input: FlagInput) => void;
  onClose: () => void;
}) {
  const editing = flag !== null;
  const [key, setKey] = useState(flag?.key ?? "");
  const [name, setName] = useState(flag?.name ?? "");
  const [description, setDescription] = useState(flag?.description ?? "");
  const [enabled, setEnabled] = useState(flag?.enabled ?? false);
  const [rollout, setRollout] = useState(flag?.rollout ?? 100);
  const [tags, setTags] = useState((flag?.tags ?? []).join(", "));

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === "Escape" && onClose();
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onClose]);

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit({
      key: editing ? undefined : key.trim(),
      name: name.trim(),
      description: description.trim(),
      enabled,
      rollout,
      tags: tags.split(",").map((t) => t.trim()).filter(Boolean),
    });
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={editing ? "Edit flag" : "Create flag"}
        className="w-full max-w-md rounded-xl border border-line bg-surface p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="mb-4 flex items-center justify-between">
          <h2 className="font-mono text-sm font-semibold">{editing ? flag.key : "New flag"}</h2>
          <button onClick={onClose} aria-label="Close" className="rounded p-1 text-muted hover:text-ink">
            <X size={16} />
          </button>
        </div>

        <form onSubmit={submit} className="flex flex-col gap-3">
          {!editing && (
            <Field label="Key">
              <Input value={key} onChange={(e) => setKey(e.target.value)} placeholder="new-checkout" autoFocus />
            </Field>
          )}
          <Field label="Name">
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="New checkout flow" />
          </Field>
          <Field label="Description">
            <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="optional" />
          </Field>
          <Field label="Tags">
            <Input value={tags} onChange={(e) => setTags(e.target.value)} placeholder="web, checkout" />
          </Field>

          <div className="flex items-center justify-between rounded-md border border-line px-3 py-2">
            <span className="font-mono text-[11px] uppercase tracking-wide text-muted">Enabled</span>
            <Toggle on={enabled} onChange={setEnabled} label="Enabled" />
          </div>

          <Field label={`Rollout — ${rollout}%`}>
            <input
              type="range"
              min={0}
              max={100}
              value={rollout}
              onChange={(e) => setRollout(Number(e.target.value))}
              className="accent-brand"
            />
          </Field>

          {error && <p className="text-sm text-danger">{error}</p>}

          <div className="mt-1 flex justify-end gap-2">
            <Button type="button" variant="ghost" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {editing ? "Save changes" : "Create flag"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
