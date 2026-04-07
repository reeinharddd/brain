import { useEffect, useState } from "react";

interface ValidationStatus {
  synced: boolean;
  orphans: string[];
  missing: string[];
}

export default function SkillsStatus() {
  const [status, setStatus] = useState<ValidationStatus | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchValidationStatus = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const res = await fetch("http://localhost:9090/api/skills/validate");
      if (res.ok) {
        const data = await res.json();
        setStatus(data);
      } else {
        setError("Failed to fetch validation status");
      }
    } catch (e) {
      setError("Failed to connect to daemon");
      console.error("Skills validation error:", e);
    } finally {
      setIsLoading(false);
    }
  };

  // Initial fetch and setup polling
  useEffect(() => {
    fetchValidationStatus();
    // Poll every 30 seconds
    const interval = setInterval(fetchValidationStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  if (isLoading) {
    return (
      <div className='skills-status loading'>
        <p>Loading skills status...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className='skills-status error'>
        <p>❌ {error}</p>
      </div>
    );
  }

  if (!status) {
    return (
      <div className='skills-status unknown'>
        <p>No status available</p>
      </div>
    );
  }

  return (
    <div className='skills-status-container'>
      {status.synced ?
        <div className='skills-status success'>
          <h3>✅ Skills Status: Perfect Sync</h3>
          <p>Registry and filesystem are perfectly synchronized.</p>
        </div>
      : <div className='skills-status error'>
          <h3>❌ Skills Status: Out of Sync</h3>

          {status.orphans && status.orphans.length > 0 && (
            <div className='orphans-section'>
              <h4>🚨 Orphan Skills ({status.orphans.length})</h4>
              <p>
                These directories exist in the filesystem but are not in the
                registry:
              </p>
              <ul>
                {status.orphans.map((orphan) => (
                  <li key={orphan}>
                    <code>{orphan}/</code> (delete or register)
                  </li>
                ))}
              </ul>
            </div>
          )}

          {status.missing && status.missing.length > 0 && (
            <div className='missing-section'>
              <h4>⚠️ Missing Skills ({status.missing.length})</h4>
              <p>
                These skills are registered but their directories don't exist on
                the filesystem:
              </p>
              <ul>
                {status.missing.map((missing) => (
                  <li key={missing}>
                    <code>{missing}/</code> (create or remove from registry)
                  </li>
                ))}
              </ul>
            </div>
          )}

          <div className='actions'>
            <button onClick={fetchValidationStatus}>🔄 Refresh Status</button>
          </div>
        </div>
      }

      <style>{`
        .skills-status-container {
          margin: 20px 0;
        }

        .skills-status {
          padding: 15px;
          border-radius: 8px;
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            sans-serif;
        }

        .skills-status.success {
          background-color: #d4edda;
          border: 1px solid #c3e6cb;
          color: #155724;
        }

        .skills-status.error {
          background-color: #f8d7da;
          border: 1px solid #f5c6cb;
          color: #721c24;
        }

        .skills-status.loading,
        .skills-status.unknown {
          background-color: #d1ecf1;
          border: 1px solid #bee5eb;
          color: #0c5460;
        }

        .skills-status h3 {
          margin-top: 0;
          margin-bottom: 10px;
          font-size: 18px;
        }

        .skills-status p {
          margin: 5px 0;
        }

        .orphans-section,
        .missing-section {
          margin-top: 15px;
          margin-bottom: 15px;
        }

        .orphans-section h4,
        .missing-section h4 {
          margin: 10px 0 5px 0;
          font-size: 14px;
        }

        .orphans-section ul,
        .missing-section ul {
          list-style-position: inside;
          padding-left: 0;
          margin: 5px 0;
        }

        .orphans-section li,
        .missing-section li {
          margin: 5px 0;
          padding-left: 10px;
        }

        .orphans-section code,
        .missing-section code {
          background-color: rgba(0, 0, 0, 0.1);
          padding: 2px 6px;
          border-radius: 3px;
          font-family: "Courier New", monospace;
          font-size: 12px;
        }

        .actions {
          margin-top: 15px;
        }

        .actions button {
          padding: 8px 16px;
          background-color: inherit;
          border: 1px solid currentColor;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
          font-weight: 500;
        }

        .actions button:hover {
          background-color: rgba(0, 0, 0, 0.1);
        }
      `}</style>
    </div>
  );
}
