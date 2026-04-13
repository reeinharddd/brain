export { default } from "./DesktopApp";
import { useState, useEffect, useCallback } from "react";
import SkillsList from "./components/SkillsList";
import SkillInstallWizard from "./components/SkillInstallWizard";

const DAEMON_URL =
  import.meta.env.VITE_BRAIN_DAEMON_URL || "http://localhost:9090";

// ─── Types ───────────────────────────────────────────────────────────────────
interface Skill {
  id: string;
  name: string;
  kind: string;
  scope: string;
  version: string;
  description: string;
  tags: string[];
  category: string;
  maintained: boolean;
  source: string;
  source_type?: string;
  source_uri?: string;
  source_variant?: string;
  artifact_path?: string;
  sync_to: string[];
  compatibility?: {
    min_capability_tier: number;
    surfaces: Record<string, string>;
  };
  security_scan?: {
    passed_at: string;
    checks: Record<string, string>;
  };
  usage?: {
    total_activations: number;
    success_rate: number;
    top_surfaces: Record<string, string>;
  };
}

interface MCPServer {
  id: string;
  name: string;
  version: string;
  status: string;
  category: string;
  transport: string;
  tool_count: number;
  client_count: number;
  last_checked: string;
  error: string;
}

interface MCPTool {
  name: string;
  description: string;
  inputSchema: Record<string, unknown>;
}

interface Agent {
  id: string;
  name: string;
  role: string;
  tier: number;
  capabilities: string[];
}

interface IDECompat {
  name: string;
  icon: string;
  tier: string;
  skills: string;
  mcps: string;
}

type TabId =
  | "status"
  | "skills"
  | "mcps"
  | "agents"
  | "ide-matrix"
  | "events";

// ─── IDE list ────────────────────────────────────────────────────────────────
const IDE_LIST: IDECompat[] = [
  { name: "Brain CLI", icon: "⌨️", tier: "1", skills: "35", mcps: "3" },
  { name: "VS Code", icon: "🟦", tier: "1", skills: "35", mcps: "14" },
  { name: "Qwen Code", icon: "🔷", tier: "1", skills: "35", mcps: "14" },
  { name: "Claude Code", icon: "🟣", tier: "2", skills: "35", mcps: "14" },
  { name: "Codex CLI", icon: "🔬", tier: "2", skills: "35", mcps: "5" },
  { name: "Cursor", icon: "🖱️", tier: "2", skills: "35", mcps: "14" },
  { name: "Windsurf", icon: "🏄", tier: "2", skills: "35", mcps: "14" },
  { name: "Continue.dev", icon: "▶️", tier: "2", skills: "35", mcps: "14" },
  { name: "Cline", icon: "🤖", tier: "2", skills: "35", mcps: "14" },
  { name: "GitHub Copilot", icon: "🐙", tier: "2", skills: "35", mcps: "14" },
  { name: "Zed", icon: "⚡", tier: "2", skills: "35", mcps: "5" },
  { name: "JetBrains", icon: "🧠", tier: "2", skills: "35", mcps: "14" },
  { name: "Gemini CLI", icon: "💎", tier: "2", skills: "35", mcps: "14" },
  { name: "OpenCode", icon: "📡", tier: "2", skills: "35", mcps: "5" },
  { name: "Neovim", icon: "📝", tier: "2", skills: "35", mcps: "5" },
  { name: "Aider", icon: "🆘", tier: "2", skills: "35", mcps: "5" },
];

// ─── App ─────────────────────────────────────────────────────────────────────
export { default } from "./DesktopApp";

// ─── Status View ─────────────────────────────────────────────────────────────
function StatusView() {
  const [status, setStatus] = useState<Record<string, unknown> | null>(null);

  const fetchData = useCallback(() => {
    fetch(`${DAEMON_URL}/api/status`)
      .then((r) => r.json())
      .then(setStatus)
      .catch(() => {});
  }, []);

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, [fetchData]);

  const subsystems = [
    { name: "Observability", detail: "OpenTelemetry + Prometheus" },
    { name: "Artifact Registry", detail: "Dependencies + versions + rollback" },
    { name: "Token Efficiency", detail: "Multi-tier cache (exact + semantic)" },
    { name: "Context Compiler", detail: "13-layer bundle + progressive disclosure" },
    { name: "Model Router", detail: "3-tier capability routing" },
    { name: "Context Curator", detail: "Dedup + compaction + autoDream" },
    { name: "Memory Sync", detail: "5 conflict strategies + encryption" },
    { name: "MCP Hub", detail: "3 real servers + proxy + registry" },
    { name: "Governance", detail: "RBAC + ABAC + OPA policies" },
    { name: "Delegation Graph", detail: "DAG + 4 modes + budgets" },
    { name: "Agent Pool", detail: "9 roles + auto-scaling" },
    { name: "Workflows", detail: "6 pre-built DAG workflows" },
    { name: "Skill Registry", detail: "35 Go skills + 8-point security scan" },
    { name: "AutoEvolve", detail: "Self-improvement engine + applier" },
    { name: "Cost Engine", detail: "Budgets + optimizer + reports" },
    { name: "Docs RAG", detail: "TF-IDF real indexer" },
  ];

  return (
    <div>
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: 16,
          marginBottom: 20,
        }}
      >
        <StatCard
          label="Daemon Status"
          value={status?.status as string | undefined}
          detail={status?.environment as string | undefined}
          color={status ? "#22c55e" : "#ef4444"}
        />
        <StatCard
          label="Processes"
          value={status?.processes ? String(status.processes) : "0"}
          detail="Managed by daemon"
          color="#3b82f6"
        />
      </div>

      <h3 style={{ marginBottom: 8, fontSize: 16 }}>Subsystems</h3>
      <table
        style={{
          width: "100%",
          borderCollapse: "collapse",
          fontSize: 13,
        }}
      >
        <thead>
          <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
            <th style={{ textAlign: "left", padding: 6 }}>Subsystem</th>
            <th style={{ textAlign: "left", padding: 6 }}>Status</th>
            <th style={{ textAlign: "left", padding: 6 }}>Details</th>
          </tr>
        </thead>
        <tbody>
          {subsystems.map((s) => (
            <tr key={s.name} style={{ borderBottom: "1px solid #f0f0f0" }}>
              <td style={{ padding: 6 }}>{s.name}</td>
              <td style={{ padding: 6 }}>
                <span
                  style={{
                    padding: "2px 8px",
                    borderRadius: 4,
                    fontSize: 11,
                    fontWeight: 600,
                    background: "#dcfce7",
                    color: "#166534",
                  }}
                >
                  OK
                </span>
              </td>
              <td style={{ padding: 6, color: "#666" }}>{s.detail}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function StatCard({
  label,
  value,
  detail,
  color,
}: {
  label: string;
  value?: string;
  detail?: string;
  color: string;
}) {
  return (
    <div
      style={{
        background: "#f8f9fa",
        borderRadius: 8,
        padding: 16,
        borderLeft: `4px solid ${color}`,
      }}
    >
      <div style={{ fontSize: 12, color: "#666", marginBottom: 4 }}>
        {label}
      </div>
      <div style={{ fontSize: 24, fontWeight: 700 }}>
        {value || "—"}
      </div>
      {detail && <div style={{ fontSize: 12, color: "#888", marginTop: 4 }}>{detail}</div>}
    </div>
  );
}

// ─── Skills View ─────────────────────────────────────────────────────────────
function SkillsView() {
  const [skills, setSkills] = useState<Skill[]>([]);
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<Skill | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshTick, setRefreshTick] = useState(0);

  const refreshSkills = useCallback(() => {
    setRefreshTick((tick) => tick + 1);
  }, []);

  useEffect(() => {
    fetch(`${DAEMON_URL}/api/skills`)
      .then((r) => r.json())
      .then((data) => {
        const list = data.skills || [];
        setSkills(Array.isArray(list) ? list : []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [refreshTick]);

  const filtered = skills.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.description.toLowerCase().includes(search.toLowerCase()) ||
      s.tags?.some((t) => t.toLowerCase().includes(search.toLowerCase()))
  );

  if (loading) return <div style={{ color: "#666" }}>Loading skills...</div>;

  return (
    <div>
      <SkillInstallWizard onInstalled={refreshSkills} />

      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <h3 style={{ margin: 0 }}>
          Skills Marketplace ({skills.length} skills)
        </h3>
        <input
          type="text"
          placeholder="🔍 Search skills by name, description, or tag..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{
            padding: "8px 12px",
            border: "1px solid #ddd",
            borderRadius: 6,
            width: 400,
            fontSize: 13,
          }}
        />
      </div>

      <table
        style={{
          width: "100%",
          borderCollapse: "collapse",
          fontSize: 13,
        }}
      >
        <thead>
          <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
            <th style={{ textAlign: "left", padding: 6 }}>ID</th>
            <th style={{ textAlign: "left", padding: 6 }}>Name</th>
            <th style={{ textAlign: "left", padding: 6 }}>Description</th>
            <th style={{ textAlign: "left", padding: 6 }}>Tags</th>
            <th style={{ textAlign: "left", padding: 6 }}>IDEs</th>
            <th style={{ textAlign: "left", padding: 6 }}>Security</th>
            <th style={{ textAlign: "left", padding: 6 }}>Action</th>
          </tr>
        </thead>
        <tbody>
          {filtered.map((skill) => (
            <tr
              key={skill.id}
              style={{
                borderBottom: "1px solid #f0f0f0",
                cursor: "pointer",
              }}
              onClick={() => setSelected(selected?.id === skill.id ? null : skill)}
            >
              <td style={{ padding: 6, fontFamily: "monospace", fontSize: 12 }}>
                {skill.id}
              </td>
              <td style={{ padding: 6, fontWeight: 500 }}>{skill.name}</td>
              <td
                style={{
                  padding: 6,
                  color: "#666",
                  maxWidth: 300,
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                {skill.description?.substring(0, 100)}...
              </td>
              <td style={{ padding: 6 }}>
                <div style={{ display: "flex", gap: 4, flexWrap: "wrap" }}>
                  {(skill.tags || []).slice(0, 3).map((t) => (
                    <span
                      key={t}
                      style={{
                        background: "#e0f2fe",
                        color: "#0369a1",
                        padding: "1px 6px",
                        borderRadius: 4,
                        fontSize: 11,
                      }}
                    >
                      {t}
                    </span>
                  ))}
                </div>
              </td>
              <td style={{ padding: 6, fontSize: 12, color: "#666" }}>
                {skill.sync_to?.length || 15} IDEs
              </td>
              <td style={{ padding: 6 }}>
                <span
                  style={{
                    background: "#dcfce7",
                    color: "#166534",
                    padding: "2px 6px",
                    borderRadius: 4,
                    fontSize: 11,
                    fontWeight: 600,
                  }}
                >
                  ✓ 8/8
                </span>
              </td>
              <td style={{ padding: 6 }}>
                <button
                  style={{
                    padding: "3px 8px",
                    border: "1px solid #3b82f6",
                    borderRadius: 4,
                    background: "white",
                    color: "#3b82f6",
                    fontSize: 11,
                    cursor: "pointer",
                  }}
                  onClick={(e) => {
                    e.stopPropagation();
                    setSelected(selected?.id === skill.id ? null : skill);
                  }}
                >
                  Details
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {selected && (
        <SkillDetail skill={selected} onClose={() => setSelected(null)} />
      )}

      <div style={{ marginTop: 20 }}>
        <SkillsList onSkillsChange={refreshSkills} />
      </div>
    </div>
  );
}

function SkillDetail({
  skill,
  onClose,
}: {
  skill: Skill;
  onClose: () => void;
}) {
  return (
    <div
      style={{
        marginTop: 16,
        background: "#f8f9fa",
        borderRadius: 8,
        padding: 16,
        border: "1px solid #e0e0e0",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          marginBottom: 12,
        }}
      >
        <h4 style={{ margin: 0 }}>{skill.name}</h4>
        <button
          onClick={onClose}
          style={{
            background: "none",
            border: "none",
            cursor: "pointer",
            fontSize: 16,
            color: "#666",
          }}
        >
          ✕
        </button>
      </div>

      <p style={{ fontSize: 13, color: "#444", marginBottom: 12 }}>
        {skill.description}
      </p>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
        <div>
          <strong style={{ fontSize: 12 }}>Compatible IDEs:</strong>
          <div style={{ display: "flex", flexWrap: "wrap", gap: 4, marginTop: 4 }}>
            {(skill.sync_to || []).map((ide) => (
              <span
                key={ide}
                style={{
                  background: "#dcfce7",
                  color: "#166534",
                  padding: "2px 8px",
                  borderRadius: 4,
                  fontSize: 11,
                }}
              >
                ✓ {ide}
              </span>
            ))}
          </div>
        </div>

        <div>
          <strong style={{ fontSize: 12 }}>Security Scan (8-point):</strong>
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: 2,
              marginTop: 4,
            }}
          >
            {skill.security_scan?.checks &&
              Object.entries(skill.security_scan.checks).map(([check, result]) => (
                <span
                  key={check}
                  style={{
                    fontSize: 11,
                    color: result === "pass" ? "#166534" : "#991b1b",
                  }}
                >
                  {result === "pass" ? "✓" : "✗"} {check}
                </span>
              ))}
          </div>
        </div>
      </div>

      <div style={{ marginTop: 12, display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
        <div>
          <strong style={{ fontSize: 12 }}>Source:</strong>
          <div style={{ fontSize: 12, color: "#666", marginTop: 4 }}>
            {skill.source_uri || skill.source || "—"}
          </div>
        </div>
        <div>
          <strong style={{ fontSize: 12 }}>Artifact:</strong>
          <div style={{ fontSize: 12, color: "#666", marginTop: 4, fontFamily: "monospace" }}>
            {skill.artifact_path || skill.source_variant || "—"}
          </div>
        </div>
      </div>

      <div style={{ marginTop: 12 }}>
        <strong style={{ fontSize: 12 }}>Tags:</strong>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 4, marginTop: 4 }}>
          {(skill.tags || []).map((t) => (
            <span
              key={t}
              style={{
                background: "#e0f2fe",
                color: "#0369a1",
                padding: "2px 8px",
                borderRadius: 4,
                fontSize: 11,
              }}
            >
              {t}
            </span>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── MCP Hub View ────────────────────────────────────────────────────────────
function MCPHubView() {
  const [servers, setServers] = useState<MCPServer[]>([]);
  const [selectedServer, setSelectedServer] = useState<string | null>(null);
  const [tools, setTools] = useState<MCPTool[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchServers = useCallback(() => {
    fetch(`${DAEMON_URL}/api/mcp/servers`)
      .then((r) => r.json())
      .then((data) => {
        setServers(data.servers || []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  useEffect(() => {
    fetchServers();
    const interval = setInterval(fetchServers, 10000);
    return () => clearInterval(interval);
  }, [fetchServers]);

  const fetchTools = (serverId: string) => {
    fetch(`${DAEMON_URL}/api/mcp/tools/${serverId}`)
      .then((r) => r.json())
      .then((data) => {
        setTools(data.tools || []);
        setSelectedServer(serverId);
      })
      .catch(() => {
        setTools([]);
        setSelectedServer(serverId);
      });
  };

  if (loading) return <div style={{ color: "#666" }}>Loading MCP servers...</div>;

  return (
    <div>
      <h3 style={{ marginBottom: 16 }}>
        MCP Server Hub ({servers.length} servers)
      </h3>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
        {/* Server List */}
        <div>
          <table
            style={{
              width: "100%",
              borderCollapse: "collapse",
              fontSize: 13,
            }}
          >
            <thead>
              <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
                <th style={{ textAlign: "left", padding: 6 }}>Server</th>
                <th style={{ textAlign: "left", padding: 6 }}>Status</th>
                <th style={{ textAlign: "left", padding: 6 }}>Tools</th>
                <th style={{ textAlign: "left", padding: 6 }}>Action</th>
              </tr>
            </thead>
            <tbody>
              {servers.map((s) => (
                <tr
                  key={s.id}
                  style={{
                    borderBottom: "1px solid #f0f0f0",
                    cursor: "pointer",
                  }}
                  onClick={() => fetchTools(s.id)}
                >
                  <td style={{ padding: 6 }}>
                    <div style={{ fontWeight: 500 }}>{s.name}</div>
                    <div style={{ fontSize: 11, color: "#888" }}>
                      {s.id} • {s.transport}
                    </div>
                  </td>
                  <td style={{ padding: 6 }}>
                    <span
                      style={{
                        background:
                          s.status === "running" ? "#dcfce7" : "#fef2f2",
                        color:
                          s.status === "running" ? "#166534" : "#991b1b",
                        padding: "2px 6px",
                        borderRadius: 4,
                        fontSize: 11,
                        fontWeight: 600,
                      }}
                    >
                      {s.status || "unknown"}
                    </span>
                  </td>
                  <td style={{ padding: 6, fontSize: 13 }}>
                    {s.tool_count || 0}
                  </td>
                  <td style={{ padding: 6 }}>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        fetch(`${DAEMON_URL}/api/mcp/servers/${s.id}/start`, {
                          method: "POST",
                        }).then(fetchServers);
                      }}
                      style={{
                        padding: "2px 6px",
                        border: "1px solid #22c55e",
                        borderRadius: 4,
                        background: "white",
                        color: "#22c55e",
                        fontSize: 11,
                        cursor: "pointer",
                      }}
                    >
                      Start
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Tools Panel */}
        <div>
          <h4 style={{ marginTop: 0, fontSize: 14 }}>
            {selectedServer
              ? `Tools: ${selectedServer}`
              : "Select a server to view tools"}
          </h4>
          {tools.length > 0 ? (
            <div>
              {tools.map((t) => (
                <div
                  key={t.name}
                  style={{
                    background: "#f8f9fa",
                    borderRadius: 6,
                    padding: 10,
                    marginBottom: 8,
                    border: "1px solid #e0e0e0",
                  }}
                >
                  <div
                    style={{
                      fontWeight: 600,
                      fontSize: 13,
                      marginBottom: 4,
                    }}
                  >
                    {t.name}
                  </div>
                  <div style={{ fontSize: 12, color: "#666" }}>
                    {t.description}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div style={{ color: "#888", fontSize: 13 }}>
              {selectedServer
                ? "No tools found for this server"
                : "Click a server to see its available tools"}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Agents View ─────────────────────────────────────────────────────────────
function AgentsView() {
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`${DAEMON_URL}/api/agents`)
      .then((r) => r.json())
      .then((data) => {
        const list = data.agents || [];
        setAgents(Array.isArray(list) ? list : []);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  if (loading) return <div style={{ color: "#666" }}>Loading agents...</div>;

  return (
    <div>
      <h3 style={{ marginBottom: 16 }}>
        Agent Pool ({agents.length} agents)
      </h3>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
        {agents.map((a) => (
          <div
            key={a.id}
            style={{
              background: "#f8f9fa",
              borderRadius: 8,
              padding: 14,
              border: "1px solid #e0e0e0",
            }}
          >
            <div style={{ fontWeight: 600, fontSize: 14, marginBottom: 4 }}>
              {a.name || a.id}
            </div>
            <div style={{ fontSize: 12, color: "#666", marginBottom: 8 }}>
              Role: {a.role || "general"} • Tier {a.tier || "2"}
            </div>
            {a.capabilities && (
              <div style={{ display: "flex", flexWrap: "wrap", gap: 4 }}>
                {a.capabilities.slice(0, 4).map((c) => (
                  <span
                    key={c}
                    style={{
                      background: "#e0f2fe",
                      color: "#0369a1",
                      padding: "1px 6px",
                      borderRadius: 4,
                      fontSize: 11,
                    }}
                  >
                    {c}
                  </span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── IDE Compatibility Matrix ────────────────────────────────────────────────
function IDECompatMatrix() {
  return (
    <div>
      <h3 style={{ marginBottom: 4 }}>IDE/CLI Compatibility Matrix</h3>
      <p style={{ fontSize: 13, color: "#666", marginBottom: 16 }}>
        All 35 Go skills and 14 MCP servers are available across every connected
        surface. Install once, use everywhere.
      </p>

      <table
        style={{
          width: "100%",
          borderCollapse: "collapse",
          fontSize: 13,
        }}
      >
        <thead>
          <tr style={{ borderBottom: "2px solid #e0e0e0" }}>
            <th style={{ textAlign: "left", padding: 6 }}>IDE / CLI</th>
            <th style={{ textAlign: "center", padding: 6 }}>Tier</th>
            <th style={{ textAlign: "center", padding: 6 }}>Skills</th>
            <th style={{ textAlign: "center", padding: 6 }}>MCP Servers</th>
            <th style={{ textAlign: "center", padding: 6 }}>Context</th>
            <th style={{ textAlign: "center", padding: 6 }}>Policy</th>
            <th style={{ textAlign: "center", padding: 6 }}>Memory</th>
          </tr>
        </thead>
        <tbody>
          {IDE_LIST.map((ide) => (
            <tr key={ide.name} style={{ borderBottom: "1px solid #f0f0f0" }}>
              <td style={{ padding: 6 }}>
                <span style={{ marginRight: 6 }}>{ide.icon}</span>
                {ide.name}
              </td>
              <td style={{ padding: 6, textAlign: "center" }}>
                <span
                  style={{
                    background: ide.tier === "1" ? "#dbeafe" : "#f0f9ff",
                    color: ide.tier === "1" ? "#1e40af" : "#0369a1",
                    padding: "2px 8px",
                    borderRadius: 4,
                    fontSize: 11,
                    fontWeight: 600,
                  }}
                >
                  {ide.tier}
                </span>
              </td>
              <td style={{ padding: 6, textAlign: "center", color: "#22c55e" }}>
                ✓ {ide.skills}
              </td>
              <td style={{ padding: 6, textAlign: "center", color: "#22c55e" }}>
                ✓ {ide.mcps}
              </td>
              <td style={{ padding: 6, textAlign: "center", color: "#22c55e" }}>
                ✓
              </td>
              <td style={{ padding: 6, textAlign: "center", color: "#22c55e" }}>
                ✓
              </td>
              <td style={{ padding: 6, textAlign: "center", color: "#22c55e" }}>
                ✓
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div
        style={{
          marginTop: 16,
          background: "#f8f9fa",
          borderRadius: 8,
          padding: 14,
          fontSize: 13,
        }}
      >
        <strong>Legend:</strong>
        <span style={{ marginLeft: 12 }}>
          <span style={{ color: "#22c55e" }}>✓</span> Fully supported
        </span>
        <span style={{ marginLeft: 12 }}>
          Tier 1 = Primary support (native integration)
        </span>
        <span style={{ marginLeft: 12 }}>
          Tier 2 = Secondary support (plugin/bridge)
        </span>
      </div>
    </div>
  );
}

// ─── Events View ─────────────────────────────────────────────────────────────
function EventsView() {
  const [events, setEvents] = useState<{ type: string; data: unknown; time: string }[]>([]);

  useEffect(() => {
    const wsUrl = DAEMON_URL.replace("http://", "ws://").replace("https://", "wss://") + "/ws";
    let ws: WebSocket | null = null;

    try {
      ws = new WebSocket(wsUrl);
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setEvents((prev) => [
            ...prev.slice(-100),
            { type: data.event || "unknown", data: data.data, time: new Date().toLocaleTimeString() },
          ]);
        } catch {
          setEvents((prev) => [
            ...prev.slice(-100),
            { type: "raw", data: event.data, time: new Date().toLocaleTimeString() },
          ]);
        }
      };
      ws.onclose = () => {
        setEvents((prev) => [
          ...prev.slice(-100),
          { type: "system", data: "WebSocket disconnected", time: new Date().toLocaleTimeString() },
        ]);
      };
    } catch {
      setEvents([
        { type: "error", data: "Cannot connect to WebSocket", time: new Date().toLocaleTimeString() },
      ]);
    }

    return () => {
      ws?.close();
    };
  }, []);

  return (
    <div>
      <h3 style={{ marginBottom: 8 }}>Live Events</h3>
      <div
        style={{
          background: "#1a1a2e",
          borderRadius: 8,
          padding: 12,
          maxHeight: 500,
          overflowY: "auto",
          fontFamily: "monospace",
          fontSize: 12,
        }}
      >
        {events.length === 0 && (
          <div style={{ color: "#888" }}>Waiting for events...</div>
        )}
        {events.map((e, i) => (
          <div
            key={i}
            style={{
              padding: "3px 0",
              color:
                e.type === "log"
                  ? "#a5f3fc"
                  : e.type === "healthcheck"
                  ? "#86efac"
                  : e.type === "error"
                  ? "#fca5a5"
                  : "#e0e7ff",
            }}
          >
            <span style={{ color: "#666" }}>[{e.time}]</span>{" "}
            <span style={{ color: "#fbbf24" }}>{e.type}</span> —{" "}
            {typeof e.data === "string" ? e.data : JSON.stringify(e.data)}
          </div>
        ))}
      </div>
    </div>
  );
}

