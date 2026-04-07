import { describe, expect, it } from "vitest";
import { createImportProfilePayload, createSubmitProfileTaskPayload } from "../src/store/dashboardPayloads";

describe("dashboard payload builders", () => {
  it("creates a plain object payload for ImportProfile", () => {
    const payload = createImportProfilePayload({
      name: "  My Profile  ",
      source: "  https://example.com/sub.yaml  ",
      content: "  proxies: []  ",
    });

    expect(Object.getPrototypeOf(payload)).toBe(Object.prototype);
    expect(payload).toEqual({
      name: "My Profile",
      source: "https://example.com/sub.yaml",
      content: "proxies: []",
    });
  });

  it("creates a plain object payload for SubmitProfileTask", () => {
    const taskConfig = {
      pingAddress: "https://www.gstatic.com/generate_204",
      pingAverageOver: 2,
      taskRetry: 2,
      taskTimeoutMillis: 6000,
      downloadURL: "https://speed.cloudflare.com/__down?bytes=10000000",
      downloadDuration: 10,
      downloadThreading: 1,
    };

    const payload = createSubmitProfileTaskPayload("profile-123", "ping", taskConfig);

    expect(Object.getPrototypeOf(payload)).toBe(Object.prototype);
    expect(Object.getPrototypeOf(payload.config ?? {})).toBe(Object.prototype);
    expect(payload).toEqual({
      profileId: "profile-123",
      preset: "ping",
      config: {
        pingAddress: "https://www.gstatic.com/generate_204",
        pingAverageOver: 2,
        taskRetry: 2,
        taskTimeoutMillis: 6000,
        downloadURL: "https://speed.cloudflare.com/__down?bytes=10000000",
        downloadDuration: 10,
        downloadThreading: 1,
      },
    });
  });
});
