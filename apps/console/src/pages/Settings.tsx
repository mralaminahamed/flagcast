import { useState } from "react";
import { getApiKey } from "../lib/api";
import { useToasts } from "../lib/toast";
import { Button, Field, Input } from "../components/ui";

export function Settings() {
  const [key, setKey] = useState(getApiKey());
  const push = useToasts((s) => s.push);

  const save = () => {
    try {
      if (key) localStorage.setItem("flagcast.apiKey", key);
      else localStorage.removeItem("flagcast.apiKey");
      push({ title: "API key saved", tone: "ok" });
    } catch {
      push({ title: "Couldn't save — storage blocked", tone: "error" });
    }
  };

  return (
    <div className="mx-auto max-w-xl px-6 py-6">
      <h1 className="mb-1 font-mono text-xl font-semibold tracking-tight">Settings</h1>
      <p className="mb-5 text-sm text-muted">Stored in this browser only.</p>

      <div className="rounded-xl border border-line bg-surface p-5">
        <Field label="Gateway API key">
          <Input
            type="password"
            value={key}
            onChange={(e) => setKey(e.target.value)}
            placeholder="only needed if the gateway requires one"
          />
        </Field>
        <p className="mt-2 text-xs text-muted">
          Sent as <span className="font-mono">X-API-Key</span> when the gateway has{" "}
          <span className="font-mono">GATEWAY_API_KEY</span> set.
        </p>
        <div className="mt-4">
          <Button onClick={save}>Save</Button>
        </div>
      </div>
    </div>
  );
}
