/**
 * DocsResults Component
 * Displays search results in a card-based layout
 */

import React from "react";
import type { SearchResult } from "../types/docs";

interface DocsResultsProps {
  results: SearchResult[];
  isLoading?: boolean;
}

const PriorityBadge: React.FC<{ priority: SearchResult["rag_priority"] }> = ({
  priority,
}) => {
  const colors = {
    critical: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    high: "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200",
    medium:
      "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
    low: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
  };

  const labels = {
    critical: "Critical",
    high: "High",
    medium: "Medium",
    low: "Low",
  };

  return (
    <span
      className={`inline-block px-2 py-1 text-xs font-semibold rounded ${colors[priority]}`}
    >
      {labels[priority]}
    </span>
  );
};

const ScoreBar: React.FC<{ score: number }> = ({ score }) => {
  const percentage = Math.min(Math.max(score * 100, 0), 100);
  const color =
    percentage > 70 ? "bg-green-500"
    : percentage > 40 ? "bg-yellow-500"
    : "bg-red-500";

  return (
    <div className='flex items-center gap-2'>
      <div className='w-24 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden'>
        <div
          className={`h-full ${color} transition-all duration-300`}
          style={{ width: `${percentage}%` }}
        />
      </div>
      <span className='text-xs text-gray-600 dark:text-gray-400 font-medium'>
        {(score * 100).toFixed(0)}%
      </span>
    </div>
  );
};

const CategoryBadge: React.FC<{ category: string }> = ({ category }) => {
  const colors: Record<string, string> = {
    architecture:
      "bg-blue-50 text-blue-700 dark:bg-blue-900 dark:text-blue-200",
    skills:
      "bg-purple-50 text-purple-700 dark:bg-purple-900 dark:text-purple-200",
    testing:
      "bg-indigo-50 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-200",
    standards: "bg-teal-50 text-teal-700 dark:bg-teal-900 dark:text-teal-200",
    templates: "bg-pink-50 text-pink-700 dark:bg-pink-900 dark:text-pink-200",
  };

  return (
    <span
      className={`text-xs px-2 py-1 rounded ${colors[category] || "bg-gray-100 text-gray-700"}`}
    >
      {category}
    </span>
  );
};

export const DocsResults: React.FC<DocsResultsProps> = ({
  results,
  isLoading,
}) => {
  if (results.length === 0) {
    return null;
  }

  return (
    <div className='w-full'>
      <div className='space-y-3'>
        {isLoading ?
          // Loading skeleton
          Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className='p-4 border border-gray-200 dark:border-gray-700 rounded-lg animate-pulse'
            >
              <div className='h-4 bg-gray-300 dark:bg-gray-600 rounded w-3/4 mb-3' />
              <div className='h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2 mb-2' />
              <div className='space-y-2'>
                <div className='h-2 bg-gray-200 dark:bg-gray-700 rounded' />
                <div className='h-2 bg-gray-200 dark:bg-gray-700 rounded w-5/6' />
              </div>
            </div>
          ))
        : results.map((result, idx) => (
            <div
              key={`${result.path}-${idx}`}
              className='p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:shadow-md hover:border-blue-300 dark:hover:border-blue-600 transition-all cursor-pointer'
              onClick={() => {
                // Navigate to document
                window.open(`/${result.path.replace(".md", "")}`, "_blank");
              }}
            >
              {/* Header with category and priority */}
              <div className='flex items-start justify-between mb-2 gap-2'>
                <h3 className='text-lg font-semibold text-gray-900 dark:text-white line-clamp-2'>
                  {result.title}
                </h3>
              </div>

              {/* Metadata row */}
              <div className='flex items-center gap-2 mb-2 flex-wrap'>
                <CategoryBadge category={result.category} />
                <PriorityBadge priority={result.rag_priority} />
                <span className='text-xs text-gray-500 dark:text-gray-400'>
                  {result.path}
                </span>
              </div>

              {/* Score bar */}
              <div className='mb-3'>
                <ScoreBar score={result.score} />
              </div>

              {/* Snippet preview */}
              <p className='text-sm text-gray-600 dark:text-gray-400 line-clamp-3'>
                {result.snippet}
              </p>
            </div>
          ))
        }
      </div>
    </div>
  );
};

export default DocsResults;
