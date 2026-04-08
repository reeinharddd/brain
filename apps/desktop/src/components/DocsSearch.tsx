/**
 * DocsSearch Component
 * Main search interface for Brain documentation
 */

import React, { useState } from "react";
import { useDocsSearch } from "../hooks/useDocsSearch";
import type { SearchResult } from "../types/docs";

interface DocsSearchProps {
  onResultsChange?: (results: SearchResult[]) => void;
  defaultDomain?: string;
}

export const DocsSearch: React.FC<DocsSearchProps> = ({
  onResultsChange,
  defaultDomain,
}) => {
  const [selectedDomain, setSelectedDomain] = useState<string>(
    defaultDomain || "",
  );
  const { state, handleInputChange, clear } = useDocsSearch({
    debounceMs: 300,
  });

  const domains = [
    "architecture",
    "skills",
    "testing",
    "standards",
    "templates",
  ];

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const query = e.target.value;
    handleInputChange(query);
  };

  const handleDomainChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const domain = e.target.value;
    setSelectedDomain(domain);
    if (state.query) {
      // Re-search with new domain filter
      handleInputChange(state.query);
    }
  };

  const handleClear = () => {
    clear();
  };

  React.useEffect(() => {
    if (onResultsChange) {
      onResultsChange(state.results);
    }
  }, [state.results, onResultsChange]);

  return (
    <div className='w-full'>
      {/* Header */}
      <div className='mb-4'>
        <h1 className='text-2xl font-bold text-gray-900 dark:text-white mb-2'>
          Brain Documentation Search
        </h1>
        <p className='text-gray-600 dark:text-gray-400'>
          Search across {state.totalResults > 0 ? "thousands of" : "78+"}{" "}
          document chunks using semantic search
        </p>
      </div>

      {/* Search Input */}
      <div className='flex flex-col gap-3 mb-6'>
        <div className='flex gap-3'>
          <div className='flex-1'>
            <input
              type='text'
              value={state.query}
              onChange={handleSearch}
              placeholder="Search documentation (e.g., 'daemon architecture', 'error handling')"
              className='w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-white'
              autoFocus
            />
          </div>
          {state.query && (
            <button
              onClick={handleClear}
              className='px-4 py-2 bg-gray-200 hover:bg-gray-300 dark:bg-gray-700 dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg transition'
            >
              Clear
            </button>
          )}
        </div>

        {/* Domain Filter */}
        <div className='flex gap-3 items-center'>
          <label
            htmlFor='domain'
            className='text-gray-700 dark:text-gray-300 font-medium'
          >
            Domain:
          </label>
          <select
            id='domain'
            value={selectedDomain}
            onChange={handleDomainChange}
            className='px-3 py-1 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 dark:bg-gray-800 dark:border-gray-600 dark:text-white'
          >
            <option value=''>All Domains</option>
            {domains.map((domain) => (
              <option key={domain} value={domain}>
                {domain.charAt(0).toUpperCase() + domain.slice(1)}
              </option>
            ))}
          </select>
        </div>
      </div>

      {/* Loading State */}
      {state.isLoading && (
        <div className='flex items-center gap-2 text-blue-600 dark:text-blue-400 mb-4'>
          <div className='animate-spin'>⚙️</div>
          <span>Searching...</span>
        </div>
      )}

      {/* Error State */}
      {state.error && (
        <div className='p-3 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-200 rounded-lg mb-4'>
          Error: {state.error}
        </div>
      )}

      {/* Results Summary */}
      {state.results.length > 0 && !state.isLoading && (
        <div className='mb-4 text-sm text-gray-600 dark:text-gray-400'>
          Found {state.totalResults} result{state.totalResults !== 1 ? "s" : ""}{" "}
          in {state.queryTime}ms
        </div>
      )}

      {/* No Results */}
      {state.query &&
        state.results.length === 0 &&
        !state.isLoading &&
        !state.error && (
          <div className='p-4 text-center text-gray-500 dark:text-gray-400'>
            No documents found matching "{state.query}"
          </div>
        )}

      {/* Empty State */}
      {!state.query && state.results.length === 0 && (
        <div className='p-4 text-center text-gray-400 dark:text-gray-600'>
          Start typing to search documentation...
        </div>
      )}
    </div>
  );
};

export default DocsSearch;
