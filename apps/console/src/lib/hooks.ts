import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";
import type { Flag, FlagInput } from "./types";

export function useFlags() {
  return useQuery({ queryKey: ["flags"], queryFn: api.flags, refetchInterval: 15000 });
}

export function useAudit() {
  return useQuery({ queryKey: ["audit"], queryFn: () => api.audit(100), refetchInterval: 20000 });
}

function invalidate(qc: ReturnType<typeof useQueryClient>) {
  qc.invalidateQueries({ queryKey: ["flags"] });
  qc.invalidateQueries({ queryKey: ["audit"] });
}

export function useCreateFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (f: FlagInput) => api.createFlag(f),
    onSuccess: () => invalidate(qc),
  });
}

export function useUpdateFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ key, input }: { key: string; input: FlagInput }) => api.updateFlag(key, input),
    // Optimistic: patch the flags cache immediately so the toggle feels live,
    // rolling back if the request fails.
    onMutate: async ({ key, input }) => {
      await qc.cancelQueries({ queryKey: ["flags"] });
      const prev = qc.getQueryData<{ flags: Flag[] }>(["flags"]);
      qc.setQueryData<{ flags: Flag[] }>(["flags"], (old) =>
        old ? { flags: old.flags.map((f) => (f.key === key ? { ...f, ...input, key } : f)) } : old,
      );
      return { prev };
    },
    onError: (_e, _v, ctx) => {
      if (ctx?.prev) qc.setQueryData(["flags"], ctx.prev);
    },
    onSettled: () => invalidate(qc),
  });
}

export function useDeleteFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key: string) => api.deleteFlag(key),
    onSuccess: () => invalidate(qc),
  });
}
