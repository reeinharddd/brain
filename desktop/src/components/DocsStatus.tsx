/**
 * DocsStatus Component
 * Shows documentation index status and health
 */

import React, { useState, useEffect } from "react";
import { docsApi } from "../api/docsApi";
import type { StatusState } from "../types/docs";

export const DocsStatus: React.FC = () => {
  const [state, setState] = useState<StatusState>({
    status: null,
    isLoading: true,
    error: null,
  });

  // Fetch status on mount and every 30 seconds
  useEffect(() => {
    const fetchStatus = async () => {
      const response = await docsApi.status();
      setState((prev) => ({
        ...prev,
        isLoading: false,
        status: response.index_status,
        error: response.error || null,
      }));
    };

    fetchStatus();
    const interval = setInterval(fetchStatus, 30000); // Poll every 30 seconds

    return () => clearInterval(interval);
  }, []);

  const getStatusColor = (status: string | undefined) => {
    switch (status) {
      case "ready":
        return "text-green-600 dark:text-green-400";
      case "indexing":
        return "text-yellow-600 dark:text-yellow-400";
      case "not_built":
        return "text-red-600 dark:text-red-400";
      default:
        return "text-gray-600 dark:text-gray-400";
    }
  };

  const getHealthColor = (health: string | undefined) => {
    switch (health) {
      case "healthy":
        return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200";
      case "degraded":
        return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200";
      case "unavailable":
        return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
      default:
        return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200";
    }
  };

  if (state.isLoading) {
    return (
      <div className='p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700 animate-pulse'>
        <div className='h-4 bg-gray-300 dark:bg-gray-600 rounded w-1/3 mb-3' />
        <div className='space-y-2'>
          <div className='h-3 bg-gray-200 dark:bg-gray-700 rounded' />
          <div className='h-3 bg-gray-200 dark:bg-gray-700 rounded w-5/6' />
        </div>
      </div>
    );
  }

  if (state.error || !state.status) {
    return (
      <div className='p-4 bg-red-50 dark:bg-red-900 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-200 rounded-lg'>
        Unable to load documentation status: {state.error}
      </div>
    );
  }

  const formatTime = (isoString: string) => {
    try {
      const date = new Date(isoString);
      return date.toLocaleString();
    } catch {
      return "Never";
    }
  };

  return (
    <div className='p-4 bg-gray-50 dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-700'>
      <h2 className='text-lg font-semibold text-gray-900 dark:text-white mb-4'>
        Documentation Index Status
      </h2>

      <div className='grid grid-cols-2 gap-4'>
        {/* Status */}
        <div>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-1'>
            Index Status
          </p>
          <div
            className={`flex items-center gap-2 ${getStatusColor(state.status.state)}`}
          >
            <div
              className={`w-3 h-3 rounded-full ${
                state.status.state === "ready" ? "bg-green-500"
                : state.status.state === "indexing" ?
                  "bg-yellow-500 animate-pulse"
                : "bg-red-500"
              }`}
            />
            <span className='capitalize font-medium'>
              {state.status.state === "not_built" ?
                "Not Built"
              : state.status.state}
            </span>
          </div>
        </div>

        {/* Qdrant Health */}
        <div>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-1'>
            Qdrant Health
          </p>
          <span
            className={`inline-block px-3 py-1 rounded text-sm font-medium ${getHealthColor(state.status.qdrant_health)}`}
          >
            {state.status.qdrant_health?.charAt(0).toUpperCase() +
              state.status.qdrant_health?.slice(1)}
          </span>
        </div>

        {/* Document Count */}
        <div>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-1'>
            Documents
          </p>
          <p className='text-2xl font-bold text-gray-900 dark:text-white'>
            {state.status.document_count}
          </p>
        </div>

        {/* Chunk Count */}
        <div>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-1'>
            Chunks Indexed
          </p>
          <p className='text-2xl font-bold text-gray-900 dark:text-white'>
            {state.status.chunk_count || "—"}
          </p>
        </div>

        {/* Last Rebuild */}
        <div className='col-span-2'>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-1'>
            Last Rebuild
          </p>
          <p className='text-sm text-gray-700 dark:text-gray-300'>
            {formatTime(state.status.last_rebuild_time)}
          </p>
        </div>
      </div>

      {/* Errors */}
      {state.status.errors && state.status.errors.length > 0 && (
        <div className='mt-4 pt-4 border-t border-gray-200 dark:border-gray-700'>
          <p className='text-xs text-gray-600 dark:text-gray-400 uppercase tracking-wide font-semibold mb-2'>
            Errors
          </p>
          <div className='space-y-1'>
            {state.status.errors.map((error, idx) => (
              <p key={idx} className='text-sm text-red-600 dark:text-red-400'>
                • {error}
              </p>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default DocsStatus;
