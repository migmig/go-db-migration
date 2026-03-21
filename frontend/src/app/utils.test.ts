import { describe, it, expect } from "vitest";
import { 
  normalizeTableKey, 
  parseReplayedTables, 
  toBool, 
  toNumber, 
  toObjectGroup, 
  toString, 
  toStringArray 
} from "./utils";

describe("utils", () => {
  describe("normalizeTableKey", () => {
    it("converts to uppercase and trims", () => {
      expect(normalizeTableKey("  users  ")).toBe("USERS");
      expect(normalizeTableKey("Orders")).toBe("ORDERS");
    });
  });

  describe("parseReplayedTables", () => {
    it("extracts tables from valid JSON", () => {
      const json = JSON.stringify({ tables: ["USERS", "ORDERS"] });
      expect(parseReplayedTables(json)).toEqual(["USERS", "ORDERS"]);
    });

    it("returns empty array for invalid JSON or missing tables", () => {
      expect(parseReplayedTables("invalid")).toEqual([]);
      expect(parseReplayedTables("{}")).toEqual([]);
    });
  });

  describe("toBool", () => {
    it("converts various types to boolean", () => {
      expect(toBool(true, false)).toBe(true);
      expect(toBool("true", false)).toBe(true);
      expect(toBool("false", true)).toBe(false);
      // Actual implementation returns fallback for numbers
      expect(toBool(1, false)).toBe(false); 
      expect(toBool(undefined, true)).toBe(true);
      expect(toBool(null, false)).toBe(false);
    });
  });

  describe("toNumber", () => {
    it("converts to number with fallback", () => {
      expect(toNumber("123", 0)).toBe(123);
      expect(toNumber(456, 0)).toBe(456);
      expect(toNumber("abc", 99)).toBe(99);
    });
  });

  describe("toString", () => {
    it("converts to string with fallback", () => {
      expect(toString("hello", "")).toBe("hello");
      // Actual implementation returns fallback for non-strings
      expect(toString(123, "fallback")).toBe("fallback");
      expect(toString(null, "fallback")).toBe("fallback");
    });
  });

  describe("toStringArray", () => {
    it("converts to string array", () => {
      expect(toStringArray(["a", "b"])).toEqual(["a", "b"]);
      // Actual implementation returns empty array if not array
      expect(toStringArray("single")).toEqual([]);
      expect(toStringArray(undefined)).toEqual([]);
    });
  });

  describe("toObjectGroup", () => {
    it("converts to valid ObjectGroup", () => {
      expect(toObjectGroup("tables", "all")).toBe("tables");
      expect(toObjectGroup("sequences", "all")).toBe("sequences");
      expect(toObjectGroup("invalid", "all")).toBe("all");
    });
  });
});
