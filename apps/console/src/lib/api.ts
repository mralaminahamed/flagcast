import type { Analysis, AuditEntry, Flag, FlagInput } from "./types";

const base = "/api";

export function getApiKey(): string {
  try {
    return localStorage.getItem("flagcast.apiKey") || "";
  } catch {
    return "";
  }
}

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const headers: Record<string, string> = { "Content-Type": "application/json", "X-Actor": "console" };
  const key = getApiKey();
  if (key) headers["X-API-Key"] = key;
  const res = await fetch(base + path, { headers, ...init });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  flags: () => req<{ flags: Flag[] }>("/flags"),
  createFlag: (f: FlagInput) => req<Flag>("/flags", { method: "POST", body: JSON.stringify(f) }),
  updateFlag: (key: string, f: FlagInput) =>
    req<Flag>(`/flags/${encodeURIComponent(key)}`, { method: "PUT", body: JSON.stringify(f) }),
  deleteFlag: (key: string) =>
    req<void>(`/flags/${encodeURIComponent(key)}`, { method: "DELETE" }),
  audit: (limit = 100) => req<{ audit: AuditEntry[] }>(`/audit?limit=${limit}`),
  analyze: (flagKey: string) =>
    req<Analysis>("/analyze", { method: "POST", body: JSON.stringify({ flag_key: flagKey }) }),
};
