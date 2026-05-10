import { describe, expect, it } from "vitest";
import { buildRemoteJoinText, parseRemoteJoinText } from "../src/utils/remoteInvite";

describe("remote invite helpers", () => {
  it("builds and parses a join URI", () => {
    const text = buildRemoteJoinText({
      host: "203.0.113.10",
      port: "10492",
      code: "ABCDEF0123456789",
    });

    expect(text).toBe("sakiko-remote://join?host=203.0.113.10&port=10492&code=ABCDEF0123456789");
    expect(parseRemoteJoinText(text)).toEqual({
      host: "203.0.113.10",
      port: "10492",
      code: "ABCDEF0123456789",
    });
  });

  it("parses plain key-value join text", () => {
    expect(parseRemoteJoinText("host=example.com port=10492 code=SKK-1234")).toEqual({
      host: "example.com",
      port: "10492",
      code: "SKK-1234",
    });
  });

  it("returns null for incomplete text", () => {
    expect(parseRemoteJoinText("host=example.com code=SKK-1234")).toBeNull();
  });
});
