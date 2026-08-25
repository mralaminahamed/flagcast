import { useMemo, useState } from "react";
import { Plus } from "lucide-react";
import { useCreateFlag, useDeleteFlag, useFlags, useUpdateFlag } from "../lib/hooks";
import type { Flag, FlagInput } from "../lib/types";
import { useToasts } from "../lib/toast";
import { Button, Empty, Input, Spinner } from "../components/ui";
import { FlagRow } from "../components/FlagRow";
import { FlagModal } from "../components/FlagModal";

type ModalState = { open: false } | { open: true; flag: Flag | null };

export function Flags() {
  const flags = useFlags();
  const create = useCreateFlag();
  const update = useUpdateFlag();
  const del = useDeleteFlag();
  const push = useToasts((s) => s.push);

  const [filter, setFilter] = useState("");
  const [modal, setModal] = useState<ModalState>({ open: false });

  const rows = useMemo(() => {
    const list = flags.data?.flags ?? [];
    const q = filter.toLowerCase();
    return list.filter((f) => f.key.toLowerCase().includes(q) || f.name.toLowerCase().includes(q));
  }, [flags.data, filter]);

  const toInput = (f: Flag, over: Partial<FlagInput>): FlagInput => ({
    name: f.name,
    description: f.description,
    enabled: f.enabled,
    rollout: f.rollout,
    tags: f.tags,
    ...over,
  });

  const toggle = (f: Flag, next: boolean) =>
    update.mutate(
      { key: f.key, input: toInput(f, { enabled: next }) },
      { onError: (e) => push({ title: (e as Error).message, tone: "error" }) },
    );

  const remove = (f: Flag) => {
    if (!window.confirm(`Delete flag "${f.key}"? SDKs will fall back to their default.`)) return;
    del.mutate(f.key, {
      onSuccess: () => push({ title: `Deleted ${f.key}`, tone: "ok" }),
      onError: (e) => push({ title: (e as Error).message, tone: "error" }),
    });
  };

  const submit = (input: FlagInput) => {
    const editing = modal.open && modal.flag;
    const done = {
      onSuccess: () => {
        setModal({ open: false });
        push({ title: editing ? `Saved ${editing.key}` : `Created ${input.key}`, tone: "ok" });
      },
    };
    if (editing) update.mutate({ key: editing.key, input }, done);
    else create.mutate(input, done);
  };

  const mutation = modal.open && modal.flag ? update : create;
  const loadError = flags.error as Error | null;

  return (
    <div className="mx-auto max-w-4xl px-6 py-6">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="font-mono text-xl font-semibold tracking-tight">Flags</h1>
          <p className="text-sm text-muted">Toggle and roll out features. Changes broadcast to SDKs live.</p>
        </div>
        <div className="flex items-center gap-2">
          <Input placeholder="Filter…" value={filter} onChange={(e) => setFilter(e.target.value)} className="w-40" />
          <Button onClick={() => setModal({ open: true, flag: null })}>
            <Plus size={15} /> New flag
          </Button>
        </div>
      </div>

      {loadError && <p className="mb-3 text-sm text-danger">Couldn't reach the API: {loadError.message}</p>}

      {flags.isLoading ? (
        <Spinner label="Loading flags…" />
      ) : rows.length === 0 ? (
        <Empty>{loadError ? "Couldn't load flags. Retrying…" : "No flags yet. Create one to get started."}</Empty>
      ) : (
        <div className="overflow-hidden rounded-xl border border-line bg-surface">
          {rows.map((f) => (
            <FlagRow
              key={f.key}
              flag={f}
              busy={update.isPending}
              onToggle={(next) => toggle(f, next)}
              onEdit={() => setModal({ open: true, flag: f })}
              onDelete={() => remove(f)}
            />
          ))}
        </div>
      )}

      {modal.open && (
        <FlagModal
          flag={modal.flag}
          pending={mutation.isPending}
          error={mutation.error ? (mutation.error as Error).message : null}
          onSubmit={submit}
          onClose={() => setModal({ open: false })}
        />
      )}
    </div>
  );
}
