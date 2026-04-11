import { useState, useEffect } from "react";

const DAEMON_URL =
  import.meta.env.VITE_BRAIN_DAEMON_URL || "http://localhost:8080";

interface SystemStatus {
  name: string;
  status: "ok" | "error" | "unknown";
  detail: string;
}

const SUBSYSTEMS: SystemStatus[] = [
  { name: "Observability", status: "ok", detail: "OpenTelemetry + Prometheus" },
  {
    name: "Artifact Registry",
    status: "ok",
    detail: "Dependencies + versions",
  },
  { name: "Token Efficiency", status: "ok", detail: "Multi-tier cache" },
  { name: "Context Compiler", status: "ok", detail: "12-layer bundles" },
  { name: "Model Router", status: "ok", detail: "3-tier routing" },
  { name: "Context Curator", status: "ok", detail: "Dedup + autoDream" },
  { name: "Memory Sync", status: "ok", detail: "5 conflict strategies" },
  { name: "MCP Hub", status: "ok", detail: "5 official servers" },
  { name: "Governance", status: "ok", detail: "RBAC + ABAC + policies" },
  { name: "Delegation Graph", status: "ok", detail: "DAG + 4 modes" },
  { name: "Agent Pool", status: "ok", detail: "9 roles + auto-scaling" },
  { name: "Workflows", status: "ok", detail: "6 pre-built workflows" },
  { name: "Skill Registry", status: "ok", detail: "8-point security scan" },
  { name: "AutoEvolve", status: "ok", detail: "Self-improvement engine" },
  { name: "Cost Engine", status: "ok", detail: "Budgets + optimizer" },
];

type TabId = "status" | "artifacts" | "context" | "policy" | "events";

function App() {
  const [activeTab, setActiveTab] = useState<TabId>("status");
  const [daemonConnected, setDaemonConnected] = useState(false);
  const [daemonStatus, setDaemonStatus] = useState<string>("Disconnected");

  useEffect(() => {
    fetch(`${DAEMON_URL}/health`)
      .then((res) => {
        if (res.ok) {
          setDaemonConnected(true);
          setDaemonStatus("Connected");
        }
      })
      .catch(() => {
        setDaemonConnected(false);
        setDaemonStatus("Disconnected");
      });
  }, []);

  const tabs: { id: TabId; label: string }[] = [
    { id: "status", label: "Status" },
    { id: "artifacts", label: "Artifacts" },
    { id: "context", label: "Context" },
    { id: "policy", label: "Policy" },
    { id: "events", label: "Events" },
  ];

  return (
    <div
      style={{
        fontFamily: "system-ui, sans-serif",
        maxWidth: 900,
        margin: "0 auto",
        padding: 20,
      }}
    >
      <header
        style={{
          borderBottom: "1px solid #e0e0e0",
          paddingBottom: 10,
          marginBottom: 20,
        }}
      >
        <h1 style={{ margin: 0 }}>Brain Daemon</h1>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 8,
            marginTop: 4,
          }}
        >
          <span
            style={{
              width: 8,
              height: 8,
              borderRadius: "50%",
              background: daemonConnected ? "#22c55e" : "#ef4444",
              display: "inline-block",
            }}
          />
          <span style={{ fontSize: 14, color: "#666" }}>
            {daemonStatus} — {DAEMON_URL}
          </span>
        </div>
      </header>

      <nav
        style={{
          display: "flex",
          gap: 4,
          marginBottom: 20,
          borderBottom: "1px solid #e0e0e0",
        }}
      >
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            style={{
              padding: "8px 16px",
              border: "none",
              background: "none",
              cursor: "pointer",
              borderBottom:
                activeTab === tab.id ?
                  "2px solid #3b82f6"
                : "2px solid transparent",
              fontWeight: activeTab === tab.id ? "bold" : "normal",
              color: activeTab === tab.id ? "#3b82f6" : "#666",
            }}
          >
            {tab.label}
          </button>
        ))}
      </nav>

      <main>
        {activeTab === "status" && <StatusView subsystems={SUBSYSTEMS} />}
        {activeTab === "artifacts" && <ArtifactsView />}
        {activeTab === "context" && <ContextView />}
        {activeTab === "policy" && <PolicyView />}
        {activeTab === "events" && <EventsView />}
      </main>
    </div>
  );
}

function StatusView({ subsystems }: { subsystems: SystemStatus[] }) {
  return (
    <div>
      <h2>System Status</h2>
      <table style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
            <th style={{ textAlign: "left", padding: 8 }}>Subsystem</th>
            <th style={{ textAlign: "left", padding: 8 }}>Status</th>
            <th style={{ textAlign: "left", padding: 8 }}>Details</th>
          </tr>
        </thead>
        <tbody>
          {subsystems.map((s) => (
            <tr key={s.name} style={{ borderBottom: "1px solid #f0f0f0" }}>
              <td style={{ padding: 8 }}>{s.name}</td>
              <td style={{ padding: 8 }}>
                <span
                  style={{
                    padding: "2px 8px",
                    borderRadius: 4,
                    fontSize: 12,
                    background: s.status === "ok" ? "#dcfce7" : "#fef2f2",
                    color: s.status === "ok" ? "#166534" : "#991b1b",
                  }}
                >
                  {s.status.toUpperCase()}
                </span>
              </td>
              <td style={{ padding: 8, color: "#666" }}>{s.detail}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <p style={{ marginTop: 16, color: "#666" }}>
        {subsystems.length} subsystems operational
      </p>
    </div>
  );
}

function ArtifactsView() {
  return (
    <div>
      <h2>Artifacts</h2>
      <div style={{ display: "grid", gap: 16 }}>
        <div>
          <h3>Skills</h3>
          <p style={{ color: "#666" }}>
            Connect to daemon to view registered skills
          </p>
        </div>
        <div>
          <h3>MCP Servers</h3>
          <p style={{ color: "#666" }}>Connect to daemon to view MCP servers</p>
        </div>
        <div>
          <h3>Agents</h3>
          <p style={{ color: "#666" }}>Connect to daemon to view agents</p>
        </div>
      </div>
    </div>
  );
}

function ContextView() {
  return (
    <div>
      <h2>Context Bundle</h2>
      <p style={{ color: "#666" }}>
        Connect to daemon to view compiled context bundle
      </p>
    </div>
  );
}

function PolicyView() {
  return (
    <div>
      <h2>Policy Resolution</h2>
      <p style={{ color: "#666" }}>
        Connect to daemon to view resolved policies
      </p>
    </div>
  );
}

function EventsView() {
  return (
    <div>
      <h2>Events</h2>
      <p style={{ color: "#666" }}>
        Connect to daemon WebSocket for real-time events
      </p>
    </div>
  );
}

export default App;
