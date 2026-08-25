import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { FlagRow } from "./FlagRow";
import type { Flag } from "../lib/types";

const flag: Flag = {
  key: "new-checkout",
  name: "New checkout",
  enabled: true,
  rollout: 40,
  tags: ["web"],
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
};

function renderRow(over: Partial<Parameters<typeof FlagRow>[0]> = {}) {
  const props = {
    flag,
    onToggle: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onAnalyze: vi.fn(),
    ...over,
  };
  render(
    <MemoryRouter>
      <FlagRow {...props} />
    </MemoryRouter>,
  );
  return props;
}

describe("FlagRow", () => {
  it("renders the flag key, name, rollout, and an on toggle", () => {
    renderRow();
    expect(screen.getByText("new-checkout")).toBeInTheDocument();
    expect(screen.getByText("New checkout")).toBeInTheDocument();
    expect(screen.getByText("40%")).toBeInTheDocument();
    expect(screen.getByRole("switch", { name: /toggle new-checkout/i })).toBeChecked();
  });

  it("fires the action callbacks", () => {
    const p = renderRow();
    fireEvent.click(screen.getByRole("switch", { name: /toggle/i }));
    fireEvent.click(screen.getByRole("button", { name: /analyze/i }));
    fireEvent.click(screen.getByRole("button", { name: /edit/i }));
    fireEvent.click(screen.getByRole("button", { name: /delete/i }));
    expect(p.onToggle).toHaveBeenCalledWith(false);
    expect(p.onAnalyze).toHaveBeenCalled();
    expect(p.onEdit).toHaveBeenCalled();
    expect(p.onDelete).toHaveBeenCalled();
  });
});
