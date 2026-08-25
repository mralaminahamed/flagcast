import { useState } from "react";
import { Outlet } from "react-router-dom";
import { MobileNav, Sidebar } from "./components/Sidebar";
import { TopBar } from "./components/TopBar";
import { Toaster } from "./components/Toaster";

export function App() {
  const [navOpen, setNavOpen] = useState(false);
  return (
    <div className="flex h-full">
      <Sidebar />
      <MobileNav open={navOpen} onClose={() => setNavOpen(false)} />
      <div className="flex min-w-0 flex-1 flex-col">
        <TopBar onMenu={() => setNavOpen(true)} />
        <main className="flex-1 overflow-y-auto">
          <Outlet />
        </main>
      </div>
      <Toaster />
    </div>
  );
}
