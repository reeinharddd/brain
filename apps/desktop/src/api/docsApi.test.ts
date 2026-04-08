/**
 * Tests for Docs API client
 */

import { describe, it, expect, beforeEach, vi } from "vitest";
import { docsApi } from "../api/docsApi";

// Mock fetch
global.fetch = vi.fn();

describe("docsApi", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("search", () => {
    it("should return empty results for empty query", async () => {
      const result = await docsApi.search("");
      expect(result.error).toBeDefined();
      expect(result.results).toEqual([]);
    });

    it("should return empty results for whitespace-only query", async () => {
      const result = await docsApi.search("   ");
      expect(result.error).toBeDefined();
      expect(result.results).toEqual([]);
    });

    it("should call fetch with correct parameters", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          results: [],
          metadata: {
            total_indexed: 10,
            query_time_ms: 100,
            index_status: "ready",
            results_count: 0,
          },
        }),
      });

      await docsApi.search("test query", 20, "architecture");

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/docs/search"),
        expect.objectContaining({
          method: "GET",
          headers: { "Content-Type": "application/json" },
        }),
      );
    });

    it("should parse successful search response", async () => {
      const mockResults = [
        {
          title: "Test Doc",
          path: "docs/test.md",
          category: "architecture",
          rag_priority: "high" as const,
          score: 0.95,
          snippet: "Test snippet",
        },
      ];

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          results: mockResults,
          metadata: {
            total_indexed: 10,
            query_time_ms: 45,
            index_status: "ready",
            results_count: 1,
          },
        }),
      });

      const result = await docsApi.search("test");

      expect(result.results).toEqual(mockResults);
      expect(result.metadata.results_count).toBe(1);
      expect(result.metadata.query_time_ms).toBe(45);
    });

    it("should handle HTTP errors", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ error: "Server error" }),
      });

      const result = await docsApi.search("test");

      expect(result.error).toBe("Server error");
      expect(result.results).toEqual([]);
    });

    it("should handle network errors", async () => {
      (global.fetch as any).mockRejectedValueOnce(new Error("Network failed"));

      const result = await docsApi.search("test");

      expect(result.error).toBe("Network failed");
      expect(result.results).toEqual([]);
    });
  });

  describe("status", () => {
    it("should fetch and return status", async () => {
      const mockStatus = {
        state: "ready" as const,
        document_count: 78,
        chunk_count: 450,
        last_rebuild_time: "2026-04-03T10:00:00Z",
        qdrant_health: "healthy" as const,
        errors: [],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({ index_status: mockStatus }),
      });

      const result = await docsApi.status();

      expect(result.index_status).toEqual(mockStatus);
    });

    it("should handle status endpoint errors", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 503,
      });

      const result = await docsApi.status();

      expect(result.error).toBe("HTTP 503");
      expect(result.index_status.qdrant_health).toBe("unavailable");
    });

    it("should handle network errors gracefully", async () => {
      (global.fetch as any).mockRejectedValueOnce(
        new Error("Connection refused"),
      );

      const result = await docsApi.status();

      expect(result.error).toBe("Connection refused");
      expect(result.index_status.state).toBe("not_built");
    });
  });

  describe("rebuild", () => {
    it("should send rebuild request", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          document_count: 78,
          duration: "2.5s",
        }),
      });

      const result = await docsApi.rebuild();

      expect(result.success).toBe(true);
      expect(result.document_count).toBe(78);
    });

    it("should send domains filter if provided", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          success: true,
          document_count: 10,
          duration: "0.5s",
        }),
      });

      await docsApi.rebuild(["architecture", "skills"]);

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining("/api/docs/rebuild"),
        expect.objectContaining({
          method: "POST",
          body: JSON.stringify({ domains: ["architecture", "skills"] }),
        }),
      );
    });

    it("should handle rebuild errors", async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({ error: "Rebuild blocked in production" }),
      });

      const result = await docsApi.rebuild();

      expect(result.success).toBe(false);
      expect(result.error).toBe("Rebuild blocked in production");
    });
  });
});
