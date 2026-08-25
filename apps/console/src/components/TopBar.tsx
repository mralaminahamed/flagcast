import { Menu, Moon, Sun } from "lucide-react";
import { useFlags } from "../lib/hooks";
import { useUI } from "../lib/store";

export function TopBar({ onMenu }: { onMenu: () => void }) {
  const { data } = useFlags();
  const { theme, toggleTheme } = useUI();
  const flags = data?.flags ?? [];
  const on = flags.filter((f) => f.enabled).length;

  return (
    <header className="flex items-center justify-between border-b border-line px-6 py-3">
      <div className="flex items-center gap-3">
        <button
          onClick={onMenu}
          aria-label="Open navigation"
          className="rounded-md border border-line bg-surface p-2 text-muted hover:text-ink sm:hidden"
        >
          <Menu size={15} />
        </button>
        <span className="font-mono text-xs text-muted">
          {flags.length === 0 ? "no flags" : `${on} of ${flags.length} live`}
        </span>
      </div>
      <button
        onClick={toggleTheme}
        aria-label="Toggle theme"
        className="rounded-md border border-line bg-surface p-2 text-muted hover:border-brand hover:text-ink"
      >
        {theme === "light" ? <Moon size={15} /> : <Sun size={15} />}
      </button>
    </header>
  );
}
