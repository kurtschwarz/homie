import { describe, expect, it } from "vitest";
import { formatSetupCode, setupCodeCharacters } from "./setup";

describe("setup code presentation", () => {
  it("splits and displays the code in two groups", () => {
    expect(setupCodeCharacters("K7QM-4XRP")).toEqual(["K", "7", "Q", "M", "4", "X", "R", "P"]);
    expect(formatSetupCode("K7QM4XRP")).toBe("K7QM-4XRP");
  });

  it("does not show more than the eight code characters", () => {
    expect(setupCodeCharacters("K7QM-4XRP-extra")).toHaveLength(8);
  });
});
