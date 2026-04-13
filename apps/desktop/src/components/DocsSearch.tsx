import { useEffect, useState } from "react";
import { Badge, Button, Card, Input, Select, StatusDot } from "../design-system";
import { useDocsSearch } from "../hooks/useDocsSearch";
import type { SearchResult } from "../types/docs";

interface DocsSearchProps {
  onResultsChange?: (results: SearchResult[]) => void;
  defaultDomain?: string;
}

export default function DocsSearch({ onResultsChange, defaultDomain }: DocsSearchProps) {
  const [selectedDomain, setSelectedDomain] = useState(defaultDomain || "");
  const { state, search, clear } = useDocsSearch({ debounceMs: 300 });

  const domains = ["architecture", "skills", "testing", "standards", "templates"];

  useEffect(() => {
    onResultsChange?.(state.results);
  }, [state.results, onResultsChange]);

  return (
    <div className="stack stack--loose">
      <div className="stack stack--dense">
        <Badge tone="accent">docs search</Badge>
        <h2 className="section-header__title">Brain Documentation Search</h2>
        <p className="section-header__subtitle">
          Search across architecture, skills, testing, standards, and templates using the daemon-backed Docs RAG surface.
        </p>
      </div>

      <Card tone="default" className="stack">
        <div className="grid-layout grid-layout--2">
          <Input
            type="text"
            value={state.query}
            onChange={(event) => search(event.target.value, selectedDomain || undefined)}
            placeholder="Search documentation (e.g., daemon architecture)"
            autoFocus
          />
          <Select
            value={selectedDomain}
            onChange={(event) => {
              const domain = event.target.value;
              setSelectedDomain(domain);
              if (state.query) {
                search(state.query, domain || undefined);
              }
            }}
          >
            <option value="">All domains</option>
            {domains.map((domain) => (
              <option key={domain} value={domain}>
                {domain.charAt(0).toUpperCase() + domain.slice(1)}
              </option>
            ))}
          </Select>
        </div>

        <div className="utility-inline">
          <Button variant="secondary" onClick={clear} disabled={!state.query}>
            Clear
          </Button>
          {state.isLoading && <Badge tone="warning">Searching</Badge>}
          {state.error && <Badge tone="danger">{state.error}</Badge>}
        </div>
      </Card>

      <Card tone="raised" className="stack">
        <div className="timeline-card__header">
          <div>
            <h3 className="timeline-card__title">Search status</h3>
            <p className="timeline-card__meta">results and query metadata</p>
          </div>
          <StatusDot tone={state.query ? "accent" : "muted"} />
        </div>

        <div className="grid-layout grid-layout--3">
          <Metric label="results" value={String(state.totalResults)} />
          <Metric label="query time" value={`${state.queryTime}ms`} />
          <Metric label="status" value={state.isLoading ? "loading" : state.error ? "error" : "ready"} />
        </div>

        {state.query && state.results.length === 0 && !state.isLoading && !state.error && (
          <div className="empty-state">
            <div className="empty-state__title">No documents found</div>
            <p>Start a new query or change the domain filter.</p>
          </div>
        )}

        {!state.query && state.results.length === 0 && (
          <div className="empty-state">
            <div className="empty-state__title">Start typing to search documentation</div>
            <p>The panel will surface semantic matches and query timing once the first request runs.</p>
          </div>
        )}
      </Card>
    </div>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <Card tone="default">
      <div className="metric-card__label">{label}</div>
      <div className="metric-card__value">{value}</div>
    </Card>
  );
}
