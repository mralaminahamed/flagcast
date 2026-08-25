import { NavLink } from "react-router-dom";
import { Flag, History, Radio, Settings } from "lucide-react";

const links = [
  { to: "/", label: "Flags", icon: Flag, end: true },
  { to: "/audit", label: "Audit", icon: History, end: false },
  { to: "/settings", label: "Settings", icon: Settings, end: false },
];

export function Sidebar() {
  return (
    <aside className="hidden w-56 shrink-0 flex-col border-r border-line bg-surface px-3 py-4 sm:flex">
      <div className="mb-6 flex items-center gap-2 px-2">
        <Radio size={20} className="text-brand" />
        <span className="font-mono text-base font-semibold tracking-tight">flagcast</span>
      </div>
      <nav className="flex flex-col gap-1">
        {links.map(({ to, label, icon: Icon, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            className={({ isActive }) =>
              `flex items-center gap-2.5 rounded-md px-2.5 py-2 text-sm transition ${
                isActive ? "bg-brand-soft text-brand" : "text-muted hover:bg-surface2 hover:text-ink"
              }`
            }
          >
            <Icon size={16} /> {label}
          </NavLink>
        ))}
      </nav>
      <div className="mt-auto px-2 font-mono text-[10px] uppercase tracking-wide text-muted">
        control plane
      </div>
    </aside>
  );
}
