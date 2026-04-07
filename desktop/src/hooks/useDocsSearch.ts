/**
 * Custom hook for documentation search
 * Manages search query, results, loading state, and error handling
 */

import { useState, useCallback } from "react";
import { docsApi } from "../api/docsApi";
import type { SearchState } from "../types/docs";

interface UseDocsSearchOptions {
  debounceMs?: number;
  initialLimit?: number;
}

export const useDocsSearch = (options: UseDocsSearchOptions = {}) => {
  const { debounceMs = 300, initialLimit = 10 } = options;

  const [state, setState] = useState<SearchState>({
    query: "",
    results: [],
    isLoading: false,
    error: null,
    totalResults: 0,
    queryTime: 0,
  });

  const [debounceTimer, setDebounceTimer] = useState<ReturnType<
    typeof setTimeout
  > | null>(null);

  /**
   * Perform search
   */
  const search = useCallback(
    async (query: string, domain?: string) => {
      // Clear previous timer
      if (debounceTimer) {
        clearTimeout(debounceTimer);
      }

      // Set loading state immediately
      setState((prev) => ({
        ...prev,
        query,
        isLoading: true,
        error: null,
      }));

      // Debounce API call
      const timer = setTimeout(async () => {
        try {
          const response = await docsApi.search(query, initialLimit, domain);

          setState((prev) => ({
            ...prev,
            results: response.results || [],
            totalResults: response.metadata?.results_count || 0,
            queryTime: response.metadata?.query_time_ms || 0,
            isLoading: false,
            error: response.error || null,
          }));
        } catch (error) {
          setState((prev) => ({
            ...prev,
            isLoading: false,
            error: error instanceof Error ? error.message : "Search failed",
          }));
        }
      }, debounceMs);

      setDebounceTimer(timer);
    },
    [debounceMs, initialLimit],
  );

  /**
   * Clear search
   */
  const clear = useCallback(() => {
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    setState({
      query: "",
      results: [],
      isLoading: false,
      error: null,
      totalResults: 0,
      queryTime: 0,
    });
  }, [debounceTimer]);

  /**
   * Handle search input change
   */
  const handleInputChange = useCallback(
    (query: string) => {
      if (query.trim()) {
        search(query);
      } else {
        clear();
      }
    },
    [search, clear],
  );

  return {
    state,
    search,
    handleInputChange,
    clear,
  };
};
