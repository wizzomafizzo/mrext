import { describe, expect, it } from "vitest";

import { encodeLaunchToken } from "./launchToken";

describe("encodeLaunchToken", () => {
  it("uses unpadded URL-safe base64", () => {
    expect(encodeLaunchToken("ÿÿÿ")).not.toMatch(/[+/=]/);
  });

  it("encodes Unicode paths as UTF-8", () => {
    expect(encodeLaunchToken("/media/fat/games/日本語.chd")).toBe(
      "L21lZGlhL2ZhdC9nYW1lcy_ml6XmnKzoqp4uY2hk",
    );
  });
});
