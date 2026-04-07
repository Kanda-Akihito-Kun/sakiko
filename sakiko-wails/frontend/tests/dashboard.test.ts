import { describe, expect, it } from "vitest";
import { formatDateTime, getFilteredNodes } from "../src/utils/dashboard";

describe("dashboard utilities", () => {
  it("formats RFC3339 timestamps into local readable text", () => {
    const formatted = formatDateTime("2026-04-07T10:01:02Z");
    expect(formatted).toContain("2026");
    expect(formatted).toContain("04");
  });

  it("filters nodes by case-insensitive substring", () => {
    const result = getFilteredNodes(
      {
        id: "p1",
        name: "demo",
        source: "",
        nodes: [
          { name: "HK-01", payload: "" },
          { name: "US-02", payload: "" },
        ],
        updatedAt: "",
      },
      "hk",
    );

    expect(result).toHaveLength(1);
    expect(result[0]?.name).toBe("HK-01");
  });
});
