import { useEffect, useState, useRef } from "react";
import SkillsList from "./components/SkillsList";
import SkillsStatus from "./components/SkillsStatus";

interface Status {
  status: string;
  time: string;
  processes: number;
}

interface Processes {
  [key: string]: string;
}

interface ManagerStatus {
  docker?: { status?: string; error?: string };
  qdrant?: { endpoint?: string; error?: string };
  ollama?: { endpoint?: string; error?: string };
  mcp?: { status?: string; error?: string };
}

interface Provider {
  name: string;
  model?: string;
  endpoint?: string;
  status?: string;
}

interface SyncStatus {
  status?: string;
  running?: boolean;
  last_run?: string;
  error?: string;
  watcher_active?: boolean;
}

export default function App() {
  const [status, setStatus] = useState<Status | null>(null);
  const [processes, setProcesses] = useState<Processes>({});
  const [logs, setLogs] = useState<string[]>([]);
  const [managers, setManagers] = useState<ManagerStatus>({});
  const [providers, setProviders] = useState<Provider[]>([]);
  const [syncStatus, setSyncStatus] = useState<SyncStatus>({});
  const [daemonRunning, setDaemonRunning] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [isSyncing, setIsSyncing] = useState(false);
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  const fetchData = async () => {
    try {
      const stRes = await fetch("http://localhost:9090/api/status");
      const stData = await stRes.json();
      setStatus(stData);
      setDaemonRunning(stData.status === "Ready");

      const psRes = await fetch("http://localhost:9090/api/processes");
      const psData = await psRes.json();
      setProcesses(psData);
    } catch (e) {
      console.error("Failed to fetch state:", e);
      setDaemonRunning(false);
    }
  };

  const fetchManagerStatuses = async () => {
    try {
      const dockerRes = await fetch(
        "http://localhost:9090/api/docker/status",
      ).catch(() => null);
      const qdrantRes = await fetch(
        "http://localhost:9090/api/qdrant/status",
      ).catch(() => null);
      const ollamaRes = await fetch(
        "http://localhost:9090/api/ollama/status",
      ).catch(() => null);
      const mcpRes = await fetch("http://localhost:9090/api/mcp/status").catch(
        () => null,
      );

      const managerData: ManagerStatus = {};

      if (dockerRes && dockerRes.ok) {
        managerData.docker = await dockerRes.json();
      } else {
        managerData.docker = { error: "Unavailable" };
      }

      if (qdrantRes && qdrantRes.ok) {
        managerData.qdrant = await qdrantRes.json();
      } else {
        managerData.qdrant = { error: "Unavailable" };
      }

      if (ollamaRes && ollamaRes.ok) {
        managerData.ollama = await ollamaRes.json();
      } else {
        managerData.ollama = { error: "Unavailable" };
      }

      if (mcpRes && mcpRes.ok) {
        managerData.mcp = await mcpRes.json();
      } else {
        managerData.mcp = { error: "Unavailable" };
      }

      setManagers(managerData);
    } catch (e) {
      console.error("Failed to fetch manager statuses:", e);
    }
  };

  const fetchProviders = async () => {
    try {
      const res = await fetch("http://localhost:9090/api/providers/available");
      if (res.ok) {
        const data = await res.json();
        setProviders(data.available || []);
      }
    } catch (e) {
      console.error("Failed to fetch providers:", e);
    }
  };

  const fetchSyncStatus = async () => {
    try {
      const res = await fetch("http://localhost:9090/api/sync/status");
      if (res.ok) {
        const data = await res.json();
        setSyncStatus(data);
      }
    } catch (e) {
      console.error("Failed to fetch sync status:", e);
    }
  };

  useEffect(() => {
    fetchData();
    fetchManagerStatuses();
    fetchProviders();
    fetchSyncStatus();
    const interval = setInterval(() => {
      fetchData();
      fetchManagerStatuses();
      fetchSyncStatus();
    }, 5000);
    return () => clearInterval(interval);
  }, []);

  useEffect(() => {
    const ws = new WebSocket("ws://localhost:9090/ws");
    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.event === "log") {
          setLogs((prev) => [...prev.slice(-99), data.data]);
        }
      } catch (e) {
        setLogs((prev) => [...prev.slice(-99), String(event.data)]);
      }
    };
    ws.onerror = () => setDaemonRunning(false);
    return () => ws.close();
  }, []);

  const startDaemon = async () => {
    setIsLoading(true);
    try {
      const res = await fetch("http://localhost:9090/api/daemon/start", {
        method: "POST",
      });
      if (res.ok) {
        setDaemonRunning(true);
        setTimeout(() => {
          fetchData();
          fetchManagerStatuses();
          fetchProviders();
          fetchSyncStatus();
        }, 2000);
      }
    } catch (e) {
      console.error("Failed to start daemon:", e);
    } finally {
      setIsLoading(false);
    }
  };

  const stopDaemon = async () => {
    setIsLoading(true);
    try {
      const res = await fetch("http://localhost:9090/api/daemon/stop", {
        method: "POST",
      });
      if (res.ok) {
        setDaemonRunning(false);
        setManagers({});
        setProviders([]);
        setSyncStatus({});
      }
    } catch (e) {
      console.error("Failed to stop daemon:", e);
    } finally {
      setIsLoading(false);
    }
  };

  const startProc = async (id: string, cmd: string, args: string[]) => {
    try {
      await fetch("http://localhost:9090/api/process/start", {
        method: "POST",
        body: JSON.stringify({ id, command: cmd, args }),
        headers: { "Content-Type": "application/json" },
      });
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  const stopProc = async (id: string) => {
    try {
      await fetch("http://localhost:9090/api/process/stop", {
        method: "POST",
        body: JSON.stringify({ id }),
        headers: { "Content-Type": "application/json" },
      });
      fetchData();
    } catch (e) {
      console.error(e);
    }
  };

  const triggerSync = async () => {
    setIsSyncing(true);
    try {
      const res = await fetch("http://localhost:9090/api/sync", {
        method: "POST",
      });
      if (res.ok) {
        setTimeout(fetchSyncStatus, 1000);
      }
    } catch (e) {
      console.error("Failed to trigger sync:", e);
    } finally {
      setIsSyncing(false);
    }
  };

  return (
    <div
      style={{
        padding: "20px",
        fontFamily: "Segoe UI, sans-serif",
        background: "#f0f2f5",
        minHeight: "100vh",
        display: "flex",
        flexDirection: "column",
        gap: "20px",
      }}
    >
      {/* Header with daemon status */}
      <header
        style={{
          background: "#ffffff",
          padding: "20px",
          borderRadius: "10px",
          boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <div>
          <h1 style={{ margin: 0, fontSize: "24px" }}>
            🧠 Brain Desktop Control Plane
          </h1>
          <small style={{ color: "#6b7280" }}>
            Daemon: {status?.time || "Initializing..."}
          </small>
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            alignItems: "flex-end",
            gap: "10px",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
            <div
              style={{
                width: "12px",
                height: "12px",
                background: daemonRunning ? "#10b981" : "#ef4444",
                borderRadius: "50%",
              }}
            ></div>
            <b>{daemonRunning ? "Running" : "Stopped"}</b>
          </div>
          <div style={{ display: "flex", gap: "8px" }}>
            <button
              onClick={startDaemon}
              disabled={daemonRunning || isLoading}
              style={{
                padding: "8px 16px",
                background: daemonRunning || isLoading ? "#d1d5db" : "#10b981",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: daemonRunning || isLoading ? "not-allowed" : "pointer",
              }}
            >
              {isLoading ? "Starting..." : "Start"}
            </button>
            <button
              onClick={stopDaemon}
              disabled={!daemonRunning || isLoading}
              style={{
                padding: "8px 16px",
                background: !daemonRunning || isLoading ? "#d1d5db" : "#ef4444",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: !daemonRunning || isLoading ? "not-allowed" : "pointer",
              }}
            >
              {isLoading ? "Stopping..." : "Stop"}
            </button>
            <button
              onClick={fetchProviders}
              style={{
                padding: "8px 16px",
                background: "#3b82f6",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
              }}
            >
              Refresh Providers
            </button>
            <button
              onClick={triggerSync}
              disabled={isSyncing}
              style={{
                padding: "8px 16px",
                background: isSyncing ? "#d1d5db" : "#8b5cf6",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isSyncing ? "not-allowed" : "pointer",
              }}
            >
              {isSyncing ? "Syncing..." : "Sync Config"}
            </button>
          </div>
        </div>
      </header>

      {/* Manager Status Dashboard */}
      <section
        style={{
          background: "#ffffff",
          padding: "20px",
          borderRadius: "10px",
          boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
        }}
      >
        <h2 style={{ marginTop: 0 }}>Manager Status</h2>
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fit, minmax(180px, 1fr))",
            gap: "15px",
          }}
        >
          {/* Docker Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: managers.docker?.error ? "#fef2f2" : "#f0fdf4",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: managers.docker?.error ? "#ef4444" : "#10b981",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                Docker
              </h3>
            </div>
            <p style={{ margin: 0, fontSize: "12px", color: "#6b7280" }}>
              {managers.docker?.error ?
                `Error: ${managers.docker.error}`
              : managers.docker?.status || "Ready"}
            </p>
          </div>

          {/* Qdrant Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: managers.qdrant?.error ? "#fef2f2" : "#f0fdf4",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: managers.qdrant?.error ? "#ef4444" : "#10b981",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                Qdrant
              </h3>
            </div>
            <p
              style={{
                margin: 0,
                fontSize: "12px",
                color: "#6b7280",
                wordBreak: "break-all",
              }}
            >
              {managers.qdrant?.error ?
                `Error: ${managers.qdrant.error}`
              : managers.qdrant?.endpoint || "http://localhost:6333"}
            </p>
          </div>

          {/* Ollama Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: managers.ollama?.error ? "#fef2f2" : "#f0fdf4",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: managers.ollama?.error ? "#ef4444" : "#10b981",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                Ollama
              </h3>
            </div>
            <p
              style={{
                margin: 0,
                fontSize: "12px",
                color: "#6b7280",
                wordBreak: "break-all",
              }}
            >
              {managers.ollama?.error ?
                `Error: ${managers.ollama.error}`
              : managers.ollama?.endpoint || "http://localhost:11434"}
            </p>
          </div>

          {/* MCP Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: managers.mcp?.error ? "#fef2f2" : "#f0fdf4",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: managers.mcp?.error ? "#ef4444" : "#10b981",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                MCP Registry
              </h3>
            </div>
            <p style={{ margin: 0, fontSize: "12px", color: "#6b7280" }}>
              {managers.mcp?.error ?
                `Error: ${managers.mcp.error}`
              : managers.mcp?.status || "Synced"}
            </p>
          </div>

          {/* Providers Summary Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: providers.length === 0 ? "#fef3f2" : "#f0f9ff",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: providers.length > 0 ? "#0ea5e9" : "#ef4444",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                LLM Providers
              </h3>
            </div>
            <p style={{ margin: 0, fontSize: "12px", color: "#6b7280" }}>
              {providers.length > 0 ?
                `${providers.length} Available`
              : "None detected"}
            </p>
          </div>

          {/* Sync Status Card */}
          <div
            style={{
              border: "1px solid #e5e7eb",
              borderRadius: "8px",
              padding: "15px",
              background: syncStatus.error ? "#fef2f2" : "#faf5ff",
            }}
          >
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: "8px",
                marginBottom: "8px",
              }}
            >
              <div
                style={{
                  width: "12px",
                  height: "12px",
                  background: syncStatus.error ? "#ef4444" : "#8b5cf6",
                  borderRadius: "50%",
                }}
              ></div>
              <h3 style={{ margin: 0, fontSize: "14px", fontWeight: 600 }}>
                Sync Engine
              </h3>
            </div>
            <p style={{ margin: 0, fontSize: "12px", color: "#6b7280" }}>
              {syncStatus.running ? "Running" : syncStatus.status || "Idle"}
            </p>
            <p
              style={{
                margin: "4px 0 0 0",
                fontSize: "12px",
                color: "#6b7280",
              }}
            >
              Watcher: {syncStatus.watcher_active ? "Active" : "Inactive"}
            </p>
            {syncStatus.last_run && (
              <p
                style={{
                  margin: "4px 0 0 0",
                  fontSize: "12px",
                  color: "#6b7280",
                }}
              >
                Last run: {syncStatus.last_run}
              </p>
            )}
            {syncStatus.error && (
              <p
                style={{
                  margin: "4px 0 0 0",
                  fontSize: "12px",
                  color: "#ef4444",
                }}
              >
                Error: {syncStatus.error}
              </p>
            )}
          </div>
        </div>
      </section>

      {/* Providers List */}
      {providers.length > 0 && (
        <section
          style={{
            background: "#ffffff",
            padding: "20px",
            borderRadius: "10px",
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          }}
        >
          <h2 style={{ marginTop: 0 }}>Available LLM Providers</h2>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "repeat(auto-fill, minmax(250px, 1fr))",
              gap: "15px",
            }}
          >
            {providers.map((p, idx) => (
              <div
                key={idx}
                style={{
                  border: "1px solid #e5e7eb",
                  borderRadius: "8px",
                  padding: "15px",
                  background: "#fafafa",
                }}
              >
                <h4 style={{ margin: "0 0 8px 0", fontSize: "14px" }}>
                  {p.name}
                </h4>
                {p.model && (
                  <p
                    style={{
                      margin: "4px 0",
                      fontSize: "12px",
                      color: "#6b7280",
                    }}
                  >
                    Model: {p.model}
                  </p>
                )}
                {p.endpoint && (
                  <p
                    style={{
                      margin: "4px 0",
                      fontSize: "12px",
                      color: "#6b7280",
                      wordBreak: "break-all",
                    }}
                  >
                    Endpoint: {p.endpoint}
                  </p>
                )}
                {p.status && (
                  <p
                    style={{
                      margin: "4px 0",
                      fontSize: "12px",
                      color: p.status === "online" ? "#10b981" : "#ef4444",
                    }}
                  >
                    Status: {p.status}
                  </p>
                )}
              </div>
            ))}
          </div>
        </section>
      )}

      {/* Main content area: Processes and Logs */}
      <div style={{ display: "flex", gap: "20px", flex: 1 }}>
        {/* Managed Processes Section */}
        <section
          style={{
            flex: 1,
            background: "#ffffff",
            padding: "20px",
            borderRadius: "10px",
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
          }}
        >
          <h2 style={{ marginTop: 0 }}>Managed Processes</h2>
          <div style={{ marginBottom: "20px", display: "flex", gap: "10px" }}>
            <button
              onClick={() =>
                startProc("web_ping", "ping", ["-c", "5", "google.com"])
              }
              style={{
                padding: "8px 16px",
                background: "#3b82f6",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
              }}
            >
              Start Diagnostics (Ping)
            </button>
          </div>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              textAlign: "left",
            }}
          >
            <thead>
              <tr style={{ borderBottom: "2px solid #e5e7eb" }}>
                <th style={{ padding: "10px" }}>ID</th>
                <th style={{ padding: "10px" }}>State</th>
                <th style={{ padding: "10px" }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {Object.entries(processes).map(([id, state]) => (
                <tr key={id} style={{ borderBottom: "1px solid #e5e7eb" }}>
                  <td style={{ padding: "10px", fontWeight: "500" }}>{id}</td>
                  <td
                    style={{
                      padding: "10px",
                      color:
                        state === "Running" ? "#10b981"
                        : state === "Failed" ? "#ef4444"
                        : "#6b7280",
                    }}
                  >
                    {state}
                  </td>
                  <td style={{ padding: "10px" }}>
                    {state === "Running" ?
                      <button
                        onClick={() => stopProc(id)}
                        style={{
                          padding: "6px 12px",
                          background: "#ef4444",
                          color: "white",
                          border: "none",
                          borderRadius: "4px",
                          cursor: "pointer",
                        }}
                      >
                        Stop
                      </button>
                    : <button
                        disabled
                        style={{
                          padding: "6px 12px",
                          background: "#e5e7eb",
                          color: "#9ca3af",
                          border: "none",
                          borderRadius: "4px",
                        }}
                      >
                        Stopped
                      </button>
                    }
                  </td>
                </tr>
              ))}
              {Object.keys(processes).length === 0 && (
                <tr>
                  <td
                    colSpan={3}
                    style={{
                      padding: "20px",
                      textAlign: "center",
                      color: "#6b7280",
                    }}
                  >
                    No processes running.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </section>

        {/* Skills Status Section */}
        <SkillsStatus />

        {/* Skills Management Section */}
        <SkillsList />

        {/* Live Logs Section */}
        <section
          style={{
            flex: 1,
            background: "#1f2937",
            color: "#f3f4f6",
            padding: "20px",
            borderRadius: "10px",
            boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
            display: "flex",
            flexDirection: "column",
          }}
        >
          <h2 style={{ color: "#e5e7eb", margin: "0 0 10px 0" }}>
            ⚡ Daemon Logs (Live)
          </h2>
          <div
            style={{
              flex: 1,
              background: "#111827",
              padding: "15px",
              borderRadius: "6px",
              fontFamily: "monospace",
              fontSize: "12px",
              overflowY: "auto",
              maxHeight: "500px",
              wordBreak: "break-all",
            }}
          >
            {logs.length === 0 ?
              <span style={{ color: "#4b5563" }}>Waiting for logs...</span>
            : logs.map((l, i) => (
                <div
                  key={i}
                  style={{
                    marginBottom: "4px",
                    paddingBottom: "4px",
                    borderBottom: "1px solid #1f2937",
                  }}
                >
                  {l}
                </div>
              ))
            }
            <div ref={logsEndRef} />
          </div>
        </section>
      </div>
    </div>
  );
}
