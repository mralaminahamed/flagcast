import { create } from "zustand";

type Theme = "light" | "dark";

function initialTheme(): Theme {
  try {
    const t = localStorage.getItem("flagcast.theme");
    if (t === "light" || t === "dark") return t;
  } catch {
    /* ignore */
  }
  return "light";
}

export function applyTheme(t: Theme) {
  document.documentElement.classList.toggle("dark", t === "dark");
}

interface UIState {
  theme: Theme;
  toggleTheme: () => void;
}

export const useUI = create<UIState>((set, get) => ({
  theme: initialTheme(),
  toggleTheme: () => {
    const next: Theme = get().theme === "light" ? "dark" : "light";
    applyTheme(next);
    try {
      localStorage.setItem("flagcast.theme", next);
    } catch {
      /* ignore */
    }
    set({ theme: next });
  },
}));
