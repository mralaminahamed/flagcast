import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./api";
import type { FlagInput } from "./types";

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
    onSuccess: () => invalidate(qc),
  });
}

export function useDeleteFlag() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (key: string) => api.deleteFlag(key),
    onSuccess: () => invalidate(qc),
  });
}
