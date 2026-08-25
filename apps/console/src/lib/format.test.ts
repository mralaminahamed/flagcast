import { describe, expect, it } from "vitest";
import { ago } from "./format";

describe("ago", () => {
  it("returns — for empty/invalid input", () => {
    expect(ago("")).toBe("—");
    expect(ago("not-a-date")).toBe("—");
  });

  it("formats recent times", () => {
    const now = Date.now();
    expect(ago(new Date(now - 5_000).toISOString())).toMatch(/s ago$/);
    expect(ago(new Date(now - 5 * 60_000).toISOString())).toBe("5m ago");
    expect(ago(new Date(now - 3 * 3600_000).toISOString())).toBe("3h ago");
    expect(ago(new Date(now - 2 * 86400_000).toISOString())).toBe("2d ago");
  });
});
