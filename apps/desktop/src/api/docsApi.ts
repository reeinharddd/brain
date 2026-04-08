/**
 * API client for Docs-RAG daemon endpoints
 * Handles /api/docs/search, /api/docs/status, /api/docs/rebuild
 */

import type {
  SearchResponse,
  StatusResponse,
  RebuildResponse,
} from "../types/docs";

const API_BASE = "/api/docs";

export const docsApi = {
  /**
   * Search documentation by query
   */
  async search(
    query: string,
    limit: number = 10,
    domain?: string,
  ): Promise<SearchResponse> {
    if (!query.trim()) {
      return {
        results: [],
        metadata: {
          total_indexed: 0,
          query_time_ms: 0,
          index_status: "not_built",
          results_count: 0,
        },
        error: "Query cannot be empty",
      };
    }

    try {
      const params = new URLSearchParams({
        q: query,
        limit: limit.toString(),
      });

      if (domain) {
        params.append("domain", domain);
      }

      const response = await fetch(`${API_BASE}/search?${params}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        const error = await response
          .json()
          .catch(() => ({ error: "Unknown error" }));
        return {
          results: [],
          metadata: {
            total_indexed: 0,
            query_time_ms: 0,
            index_status: "not_built",
            results_count: 0,
          },
          error: error.error || `HTTP ${response.status}`,
        };
      }

      return await response.json();
    } catch (error) {
      return {
        results: [],
        metadata: {
          total_indexed: 0,
          query_time_ms: 0,
          index_status: "not_built",
          results_count: 0,
        },
        error: error instanceof Error ? error.message : "Network error",
      };
    }
  },

  /**
   * Get documentation index status
   */
  async status(): Promise<StatusResponse> {
    try {
      const response = await fetch(`${API_BASE}/status`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        return {
          index_status: {
            state: "not_built",
            document_count: 0,
            last_rebuild_time: "0001-01-01T00:00:00Z",
            qdrant_health: "unavailable",
            errors: [`HTTP ${response.status}`],
          },
          error: `HTTP ${response.status}`,
        };
      }

      return await response.json();
    } catch (error) {
      return {
        index_status: {
          state: "not_built",
          document_count: 0,
          last_rebuild_time: "0001-01-01T00:00:00Z",
          qdrant_health: "unavailable",
          errors: [error instanceof Error ? error.message : "Network error"],
        },
        error: error instanceof Error ? error.message : "Network error",
      };
    }
  },

  /**
   * Rebuild documentation index (dev-only)
   */
  async rebuild(domains?: string[]): Promise<RebuildResponse> {
    try {
      const response = await fetch(`${API_BASE}/rebuild`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: domains ? JSON.stringify({ domains }) : undefined,
      });

      if (!response.ok) {
        const error = await response
          .json()
          .catch(() => ({ error: "Unknown error" }));
        return {
          success: false,
          document_count: 0,
          duration: "0s",
          error: error.error || `HTTP ${response.status}`,
        };
      }

      return await response.json();
    } catch (error) {
      return {
        success: false,
        document_count: 0,
        duration: "0s",
        error: error instanceof Error ? error.message : "Network error",
      };
    }
  },
};
