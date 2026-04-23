import { beforeEach, describe, expect, it } from "vitest";
import { useNotificationStore } from "../src/store/notificationStore";

describe("notificationStore", () => {
  beforeEach(() => {
    useNotificationStore.setState({ items: [] });
  });

  it("deduplicates identical notifications within the dedupe window", () => {
    const { push } = useNotificationStore.getState();

    push({
      level: "warning",
      message: "profile refresh failed",
      source: "backend",
      timestamp: "2026-04-14T00:00:00.000Z",
    });
    push({
      level: "warning",
      message: "profile refresh failed",
      source: "backend",
      timestamp: "2026-04-14T00:00:01.500Z",
    });

    const [item] = useNotificationStore.getState().items;
    expect(useNotificationStore.getState().items).toHaveLength(1);
    expect(item?.count).toBe(2);
  });

  it("keeps only the newest notifications within the cap", () => {
    const { push } = useNotificationStore.getState();

    for (let index = 0; index < 6; index += 1) {
      push({
        level: "error",
        message: `error-${index}`,
        source: "backend",
        timestamp: `2026-04-14T00:00:0${index}.000Z`,
      });
    }

    const items = useNotificationStore.getState().items;
    expect(items).toHaveLength(4);
    expect(items.map((item) => item.message)).toEqual(["error-5", "error-4", "error-3", "error-2"]);
  });
});
