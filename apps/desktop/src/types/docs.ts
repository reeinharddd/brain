/**
 * Types for Docs-RAG API responses and UI state
 */

export interface SearchResult {
  title: string;
  path: string;
  category: string;
  rag_priority: "critical" | "high" | "medium" | "low";
  score: number;
  snippet: string;
}

export interface SearchMetadata {
  total_indexed: number;
  query_time_ms: number;
  index_status: "ready" | "indexing" | "not_built";
  results_count: number;
}

export interface SearchResponse {
  results: SearchResult[];
  metadata: SearchMetadata;
  error?: string;
}

export interface IndexStatus {
  state: "ready" | "indexing" | "not_built";
  document_count: number;
  chunk_count?: number;
  last_rebuild_time: string;
  qdrant_health: "healthy" | "degraded" | "unavailable";
  errors: string[];
}

export interface StatusResponse {
  index_status: IndexStatus;
  error?: string;
}

export interface RebuildResponse {
  success: boolean;
  document_count: number;
  duration: string;
  error?: string;
}

export interface DocsApi {
  search(
    query: string,
    limit?: number,
    domain?: string,
  ): Promise<SearchResponse>;
  status(): Promise<StatusResponse>;
  rebuild(): Promise<RebuildResponse>;
}

export interface SearchState {
  query: string;
  results: SearchResult[];
  isLoading: boolean;
  error: string | null;
  totalResults: number;
  queryTime: number;
}

export interface StatusState {
  status: IndexStatus | null;
  isLoading: boolean;
  error: string | null;
}
