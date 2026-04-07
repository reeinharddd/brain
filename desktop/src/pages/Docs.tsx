/**
 * Docs Page
 * Main page integrating search, results, and status components
 */

import React, { useState } from "react";
import DocsSearch from "../components/DocsSearch";
import DocsResults from "../components/DocsResults";
import DocsStatus from "../components/DocsStatus";
import type { SearchResult } from "../types/docs";

export const DocsPage: React.FC = () => {
  const [results, setResults] = useState<SearchResult[]>([]);

  return (
    <div className='min-h-screen bg-white dark:bg-gray-950'>
      <div className='max-w-6xl mx-auto px-4 py-8'>
        {/* Header */}
        <div className='mb-8'>
          <h1 className='text-4xl font-bold text-gray-900 dark:text-white mb-2'>
            Brain Docs
          </h1>
          <p className='text-lg text-gray-600 dark:text-gray-400'>
            Semantic search over all Brain documentation
          </p>
        </div>

        <div className='grid grid-cols-3 gap-8'>
          {/* Left Column: Search and Results */}
          <div className='col-span-2 space-y-6'>
            {/* Search Component */}
            <div className='bg-gray-50 dark:bg-gray-900 p-6 rounded-lg border border-gray-200 dark:border-gray-800'>
              <DocsSearch onResultsChange={setResults} />
            </div>

            {/* Results Component */}
            <div>
              <DocsResults results={results} />
            </div>
          </div>

          {/* Right Column: Status Panel */}
          <div className='col-span-1'>
            <div className='sticky top-8'>
              <DocsStatus />
            </div>
          </div>
        </div>

        {/* Footer Info */}
        <div className='mt-12 pt-8 border-t border-gray-200 dark:border-gray-800'>
          <div className='grid grid-cols-3 gap-6 text-sm text-gray-600 dark:text-gray-400'>
            <div>
              <h3 className='font-semibold text-gray-900 dark:text-white mb-2'>
                How Search Works
              </h3>
              <p>
                Uses semantic embeddings (all-MiniLM-L6-v2) stored in Qdrant to
                understand query meaning, not just keywords.
              </p>
            </div>
            <div>
              <h3 className='font-semibold text-gray-900 dark:text-white mb-2'>
                Supported Domains
              </h3>
              <ul className='space-y-1'>
                <li>• Architecture decisions</li>
                <li>• Skills & capabilities</li>
                <li>• Testing strategies</li>
                <li>• Standards & conventions</li>
                <li>• Templates & examples</li>
              </ul>
            </div>
            <div>
              <h3 className='font-semibold text-gray-900 dark:text-white mb-2'>
                Priority Levels
              </h3>
              <ul className='space-y-1'>
                <li>🔴 Critical (1.5x boost)</li>
                <li>🟠 High (1.2x boost)</li>
                <li>🟡 Medium (baseline)</li>
                <li>🟢 Low (0.8x reduction)</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default DocsPage;
