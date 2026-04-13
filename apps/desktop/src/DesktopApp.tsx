import { useEffect, useMemo, useState, type FormEvent } from "react";
import {
  Badge,
  Button,
  Card,
  CodeBlock,
  CommandBar,
  Input,
  MetricCard,
  Panel,
  ProgressBar,
  SectionHeader,
  StatusDot,
} from "./design-system";
import {
  COMMAND_HINTS,
  DESKTOP_SECTIONS,
  type DesktopSectionId,
  type ThemeMode,
  toggleTheme,
  applyTheme,
  resolveInitialTheme,
} from "./design-system";
import {
  DAEMON_URL,
  clearStoredAuthSession,
  defaultAuthStatus,
  fetchDesktopAuthStatus,
  loginToDaemon,
  loginWithOidc,
  logoutFromDaemon,
  type DesktopAuthStatus,
} from "./api/auth";
import VisualSystemPage from "./pages/VisualSystem";
import LoginPage from "./components/LoginPage";

type RuntimeProvider = {
  name: string;
  status: string;
  latency: string;
  throughput: string;
  tone: "accent" | "success" | "warning" | "danger" | "info" | "muted";
};

type RuntimeState = {
  connected: boolean;
  status: string;
  environment: string;
  cluster: string;
  latency: string;
  memoryUsed: string;
  memoryAllocated: string;
  processes: number;
  syncStatus: string;
  syncRunning: boolean;
  syncLastRun: string;
  syncError: string;
  activeAgents: number;
  activeSessions: number;
  providers: RuntimeProvider[];
  sequence: string[];
};

type AgentState = {
  id: string;
  name: string;
  description: string;
  version: string;
  model: string;
  temperature: number;
  promptFile: string;
  maintained: boolean;
  tags: string[];
};

type MemoryEntry = {
  time: string;
  label: string;
  kind: string;
  summary: string;
  tags: string[];
};

type McpServer = {
  name: string;
  id: string;
  version: string;
  status: string;
  transport: string;
  tools: number;
  clientCount: number;
  error: string;
  tone: "accent" | "success" | "warning" | "danger" | "info" | "muted";
};

type LogEntry = {
  time: string;
  channel: string;
  tone: "accent" | "success" | "warning" | "danger" | "info" | "muted";
  message: string;
};

type EvalMetric = {
  label: string;
  value: string;
  detail: string;
  progress: number;
  tone: "accent" | "success" | "warning" | "danger" | "info" | "muted";
};

type ReferenceSummary = {
  skillsCount: number;
  docsStatus: string;
  workflowsCount: number;
  autoEvolveStatus: string;
  delegationCount: number;
};

type AuthFormState = {
  email: string;
  password: string;
  busy: boolean;
  error: string;
  success: string;
};

const runtimeProvidersFallback: RuntimeProvider[] = [];

const agentFallback: AgentState[] = [];

const memoryFallback: MemoryEntry[] = [];

const mcpFallback: McpServer[] = [];

const evalFallback: EvalMetric[] = [];

const runtimeLogFallback: LogEntry[] = [];

const rulesValidationFallback = {
  version: "",
  active: "",
  history: [] as string[],
  validation: [] as string[],
};

const sectionLookup = Object.fromEntries(
  DESKTOP_SECTIONS.map((section) => [section.id, section]),
);

function toneForStatus(
  status: string,
): "accent" | "success" | "warning" | "danger" | "info" | "muted" {
  const normalized = status.toLowerCase();
  if (
    normalized.includes("active") ||
    normalized.includes("running") ||
    normalized.includes("stable") ||
    normalized.includes("healthy") ||
    normalized.includes("ready")
  ) {
    return "success";
  }
  if (
    normalized.includes("warning") ||
    normalized.includes("idle") ||
    normalized.includes("degraded")
  ) {
    return "warning";
  }
  if (
    normalized.includes("fault") ||
    normalized.includes("error") ||
    normalized.includes("offline") ||
    normalized.includes("unavailable")
  ) {
    return "danger";
  }
  if (
    normalized.includes("sync") ||
    normalized.includes("connected") ||
    normalized.includes("online")
  ) {
    return "accent";
  }
  return "muted";
}

function App() {
  const [activeSection, setActiveSection] =
    useState<DesktopSectionId>("runtime");
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [theme, setTheme] = useState<ThemeMode>(() => resolveInitialTheme());
  const [command, setCommand] = useState("/system help");
  const [daemonConnected, setDaemonConnected] = useState(false);
  const [refreshTick, setRefreshTick] = useState(0);
  const [auth, setAuth] = useState<DesktopAuthStatus>(() =>
    defaultAuthStatus(),
  );
  const [authForm, setAuthForm] = useState<AuthFormState>({
    email: "",
    password: "",
    busy: false,
    error: "",
    success: "",
  });
  const [runtime, setRuntime] = useState<RuntimeState>({
    connected: false,
    status: "",
    environment: "",
    cluster: "",
    latency: "",
    memoryUsed: "",
    memoryAllocated: "",
    processes: 0,
    syncStatus: "",
    syncRunning: false,
    syncLastRun: "",
    syncError: "",
    activeAgents: 0,
    activeSessions: 0,
    providers: runtimeProvidersFallback,
    sequence: [],
  });
  const [agents, setAgents] = useState<AgentState[]>(agentFallback);
  const [memoryEntries] = useState<MemoryEntry[]>(memoryFallback);
  const [mcpServers, setMcpServers] = useState<McpServer[]>(mcpFallback);
  const [logs, setLogs] = useState<LogEntry[]>(runtimeLogFallback);
  const [filterLogs, setFilterLogs] = useState("");
  const [referenceSummary, setReferenceSummary] = useState<ReferenceSummary>({
    skillsCount: 0,
    docsStatus: "",
    workflowsCount: 0,
    autoEvolveStatus: "",
    delegationCount: 0,
  });

  useEffect(() => {
    applyTheme(theme);
  }, [theme]);

  useEffect(() => {
    let cancelled = false;

    const loadAuth = async () => {
      const status = await fetchDesktopAuthStatus();
      if (!cancelled) {
        setAuth(status);
      }
      if (!status.authenticated) {
        setAuthForm((current) => ({ ...current, password: "" }));
      }
    };

    void loadAuth();
    const interval = window.setInterval(() => {
      void loadAuth();
    }, 15000);

    return () => {
      cancelled = true;
      window.clearInterval(interval);
    };
  }, []);

  useEffect(() => {
    const checkDaemon = async () => {
      try {
        const response = await fetch(`${DAEMON_URL}/health`);
        setDaemonConnected(response.ok);
      } catch {
        setDaemonConnected(false);
      }
    };

    const loadRuntime = async () => {
      try {
        const [statusResponse, providersResponse] = await Promise.all([
          fetch(`${DAEMON_URL}/api/status`),
          fetch(`${DAEMON_URL}/api/providers/available`),
        ]);

        if (statusResponse.ok) {
          const data = await statusResponse.json();
          setRuntime((current) => ({
            ...current,
            connected: true,
            status: String(data.status || current.status),
            environment: String(data.environment || current.environment),
            processes: Number(data.processes || 0),
            syncStatus: String(data.sync_status || ""),
            syncRunning: Boolean(data.sync_running),
            syncLastRun: String(data.sync_last_run || ""),
            syncError: String(data.sync_error || ""),
          }));
        }

        if (providersResponse.ok) {
          const data = await providersResponse.json();
          const available = Array.isArray(data.available) ? data.available : [];
          setRuntime((current) => ({
            ...current,
            providers: available.map((provider: unknown) => ({
              name: String(provider),
              status: "available",
              latency: "",
              throughput: "",
              tone: "success",
            })),
          }));
        }
      } catch {
        setRuntime((current) => ({ ...current, connected: false }));
      }
    };

    const loadAgents = async () => {
      try {
        const response = await fetch(`${DAEMON_URL}/api/agents`);
        if (!response.ok) {
          setAgents([]);
          setRuntime((current) => ({ ...current, activeAgents: 0 }));
          return;
        }
        const data = await response.json();
        const list =
          Array.isArray(data.agents) ?
            data.agents.map((agent: any) => ({
              id: String(agent.id || agent.name),
              name: String(agent.name || agent.id || "Agent"),
              description: String(agent.description || ""),
              version: String(agent.version || ""),
              model: String(agent.model || ""),
              temperature: Number(agent.temperature || 0),
              promptFile: String(agent.prompt_file || ""),
              maintained: Boolean(agent.maintained),
              tags:
                Array.isArray(agent.tags) ?
                  agent.tags.map((tag: unknown) => String(tag))
                : [],
            }))
          : [];
        setAgents(list);
        setRuntime((current) => ({ ...current, activeAgents: list.length }));
      } catch {
        setAgents([]);
        setRuntime((current) => ({ ...current, activeAgents: 0 }));
      }
    };

    const loadMcp = async () => {
      try {
        const response = await fetch(`${DAEMON_URL}/api/mcp/servers`);
        if (!response.ok) {
          setMcpServers([]);
          return;
        }
        const data = await response.json();
        const list =
          Array.isArray(data.servers) ?
            data.servers.map((server: any) => ({
              name: String(server.name || server.id),
              id: String(server.id || server.name),
              version: String(server.version || ""),
              status: String(server.status || "unknown"),
              transport: String(server.transport || "stdio"),
              tools: Number(server.tool_count || server.tools || 0),
              clientCount: Number(server.client_count || 0),
              error: server.error ? String(server.error) : "",
              tone: toneForStatus(String(server.status || "")),
            }))
          : [];
        setMcpServers(list);
      } catch {
        setMcpServers([]);
      }
    };

    const loadReferenceSummary = async () => {
      try {
        const [
          skillsResponse,
          docsResponse,
          workflowsResponse,
          autoEvolveResponse,
          delegationResponse,
        ] = await Promise.all([
          fetch(`${DAEMON_URL}/api/skills`),
          fetch(`${DAEMON_URL}/api/docs/status`),
          fetch(`${DAEMON_URL}/api/workflows/list`),
          fetch(`${DAEMON_URL}/api/autoevolve/status`),
          fetch(`${DAEMON_URL}/api/delegation/executions`),
        ]);

        const nextSummary: ReferenceSummary = {
          skillsCount: 0,
          docsStatus: "",
          workflowsCount: 0,
          autoEvolveStatus: "",
          delegationCount: 0,
        };

        if (skillsResponse.ok) {
          const data = await skillsResponse.json();
          nextSummary.skillsCount =
            Array.isArray(data.skills) ? data.skills.length : 0;
        }

        if (docsResponse.ok) {
          const data = await docsResponse.json();
          nextSummary.docsStatus = JSON.stringify(
            data.index_status || data,
            null,
            2,
          );
        }

        if (workflowsResponse.ok) {
          const data = await workflowsResponse.json();
          nextSummary.workflowsCount =
            Array.isArray(data.workflows) ? data.workflows.length : 0;
        }

        if (autoEvolveResponse.ok) {
          const data = await autoEvolveResponse.json();
          nextSummary.autoEvolveStatus = JSON.stringify(data, null, 2);
        }

        if (delegationResponse.ok) {
          const data = await delegationResponse.json();
          nextSummary.delegationCount =
            Array.isArray(data.executions) ? data.executions.length : 0;
        }

        setReferenceSummary(nextSummary);
      } catch {
        setReferenceSummary({
          skillsCount: 0,
          docsStatus: "",
          workflowsCount: 0,
          autoEvolveStatus: "",
          delegationCount: 0,
        });
      }
    };

    void checkDaemon();
    void loadRuntime();
    void loadAgents();
    void loadMcp();
    void loadReferenceSummary();

    const interval = window.setInterval(() => {
      void checkDaemon();
      void loadRuntime();
      void loadAgents();
      void loadMcp();
      void loadReferenceSummary();
    }, 15000);

    return () => window.clearInterval(interval);
  }, [refreshTick]);

  useEffect(() => {
    const wsUrl = `${DAEMON_URL.replace(/^http/, "ws")}/ws`;
    let socket: WebSocket | null = null;

    try {
      socket = new WebSocket(wsUrl);
      socket.onmessage = (event) => {
        const stamp = new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        });
        try {
          const parsed = JSON.parse(event.data) as {
            event?: string;
            data?: unknown;
          };
          setLogs((current) => [
            ...current.slice(-96),
            {
              time: stamp,
              channel: String(parsed.event || "SYSTEM").toUpperCase(),
              tone: toneForStatus(String(parsed.event || "")),
              message:
                typeof parsed.data === "string" ?
                  parsed.data
                : JSON.stringify(parsed.data || {}),
            },
          ]);
        } catch {
          setLogs((current) => [
            ...current.slice(-96),
            {
              time: stamp,
              channel: "STREAM",
              tone: "info",
              message: String(event.data),
            },
          ]);
        }
      };
    } catch {
      socket = null;
    }

    return () => {
      socket?.close();
    };
  }, []);

  const activeSectionMeta = sectionLookup[activeSection];
  const allowedSectionSet = useMemo(
    () =>
      new Set(
        auth.allowedSections.length > 0 ?
          auth.allowedSections
        : DESKTOP_SECTIONS.map((section) => section.id),
      ),
    [auth.allowedSections],
  );
  const authBadgeTone =
    auth.authenticated ? "success"
    : auth.required ? "warning"
    : "neutral";
  const authBadgeLabel =
    auth.authenticated ? auth.user?.role || "signed in"
    : auth.required ? "login required"
    : "guest";

  const filteredLogs = useMemo(
    () =>
      logs.filter((entry) => {
        if (!filterLogs.trim()) {
          return true;
        }
        const query = filterLogs.toLowerCase();
        return (
          entry.channel.toLowerCase().includes(query) ||
          entry.message.toLowerCase().includes(query) ||
          entry.time.toLowerCase().includes(query)
        );
      }),
    [filterLogs, logs],
  );

  const handleAuthLogin = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setAuthForm((current) => ({
      ...current,
      busy: true,
      error: "",
      success: "",
    }));
    try {
      const nextAuth = await loginToDaemon(authForm.email, authForm.password);
      setAuth(nextAuth);
      setAuthForm((current) => ({
        ...current,
        busy: false,
        password: "",
        email: nextAuth.user?.email || current.email,
        error: "",
        success: `Signed in as ${nextAuth.user?.name || nextAuth.user?.email || "Brain user"}`,
      }));
      setRefreshTick((value) => value + 1);
    } catch (error) {
      setAuthForm((current) => ({
        ...current,
        busy: false,
        error: error instanceof Error ? error.message : "login failed",
        success: "",
      }));
    }
  };

  const handleAuthOidcLogin = async () => {
    setAuthForm((current) => ({
      ...current,
      busy: true,
      error: "",
      success: "",
    }));
    try {
      const nextAuth = await loginWithOidc(authForm.email);
      setAuth(nextAuth);
      setAuthForm((current) => ({
        ...current,
        busy: false,
        password: "",
        email: nextAuth.user?.email || current.email,
        error: "",
        success: `Signed in via OIDC as ${nextAuth.user?.name || nextAuth.user?.email || "Brain user"}`,
      }));
      setRefreshTick((value) => value + 1);
    } catch (error) {
      setAuthForm((current) => ({
        ...current,
        busy: false,
        error: error instanceof Error ? error.message : "oidc login failed",
        success: "",
      }));
    }
  };

  const handleAuthLogout = async () => {
    setAuthForm((current) => ({ ...current, busy: true, error: "" }));
    try {
      await logoutFromDaemon();
      await clearStoredAuthSession();
      const nextAuth = await fetchDesktopAuthStatus();
      setAuth(nextAuth);
      setRefreshTick((value) => value + 1);
      setAuthForm((current) => ({
        ...current,
        busy: false,
        password: "",
        error: "",
      }));
    } catch (error) {
      setAuthForm((current) => ({
        ...current,
        busy: false,
        error: error instanceof Error ? error.message : "logout failed",
      }));
    }
  };

  const handleLoginSuccess = async () => {
    const status = await fetchDesktopAuthStatus();
    setAuth(status);
    setRefreshTick((value) => value + 1);
  };

  if (!auth.authenticated) {
    return <LoginPage onLoginSuccess={handleLoginSuccess} />;
  }

  return (
    <div
      className='app-shell'
      data-sidebar-open={sidebarOpen ? "true" : "false"}
    >
      <button
        type='button'
        className='app-shell__backdrop'
        aria-label='Close navigation drawer'
        onClick={() => setSidebarOpen(false)}
      />

      <aside className='app-sidebar'>
        <div className='app-sidebar__brand'>
          <p className='app-sidebar__name'>brain</p>
          <div className='app-sidebar__version'>v1.0.4-stable</div>
        </div>

        <nav className='app-sidebar__nav' aria-label='Desktop sections'>
          {DESKTOP_SECTIONS.map((section) => {
            const isActive = activeSection === section.id;
            const isLocked = !allowedSectionSet.has(section.id);
            return (
              <button
                key={section.id}
                type='button'
                className={`app-sidebar__nav-item${isActive ? " is-active" : ""}`}
                onClick={() => {
                  setActiveSection(section.id);
                  setSidebarOpen(false);
                }}
              >
                <StatusDot
                  tone={
                    isActive ? "accent"
                    : isLocked ?
                      "warning"
                    : "muted"
                  }
                />
                <span>
                  {section.label}
                  <span className='app-sidebar__nav-meta'>
                    {section.description}
                    {isLocked ? " · locked until login" : " · available"}
                  </span>
                </span>
                {isLocked && <Badge tone='warning'>locked</Badge>}
              </button>
            );
          })}
        </nav>

        <div className='app-sidebar__footer'>
          <Button
            className='app-sidebar__action'
            variant='primary'
            onClick={() => setActiveSection("runtime")}
          >
            New Session
          </Button>
        </div>
      </aside>

      <div className='app-main'>
        <header className='app-topbar'>
          <div className='app-topbar__left'>
            <Button
              className='app-mobile-toggle'
              variant='ghost'
              onClick={() => setSidebarOpen((value) => !value)}
            >
              Menu
            </Button>
            <div className='app-topbar__title'>
              <h1>Brain Control Plane</h1>
              <div className='app-topbar__subtitle'>
                runtime, memory, orchestration, and policies
              </div>
            </div>
          </div>

          <div className='app-topbar__right'>
            <Badge tone={daemonConnected ? "success" : "danger"}>
              {daemonConnected ? "connected" : "offline"}
            </Badge>
            <Badge tone={theme === "dark" ? "accent" : "neutral"}>
              {theme}
            </Badge>
            <Badge tone={authBadgeTone}>{authBadgeLabel}</Badge>
            <Button variant='secondary' onClick={handleAuthLogout}>
              Logout
            </Button>
            <Button
              variant='secondary'
              onClick={() => setTheme((current) => toggleTheme(current))}
            >
              Toggle theme
            </Button>
          </div>
        </header>

        <main className='app-main__content'>
          <AuthSection
            auth={auth}
            form={authForm}
            onSubmit={handleAuthLogin}
            onOidcLogin={handleAuthOidcLogin}
            onLogout={handleAuthLogout}
            onEmailChange={(nextEmail) =>
              setAuthForm((current) => ({ ...current, email: nextEmail }))
            }
            onPasswordChange={(nextPassword) =>
              setAuthForm((current) => ({ ...current, password: nextPassword }))
            }
          />

          <SectionHeader
            kicker='brain runtime'
            title={activeSectionMeta.label}
            subtitle={activeSectionMeta.description}
            actions={
              <>
                <Badge tone={runtime.connected ? "success" : "warning"}>
                  {runtime.connected ? "daemon sync" : "local fallback"}
                </Badge>
                <Badge tone={toneForStatus(runtime.status)}>
                  {runtime.status}
                </Badge>
                <Badge tone={authBadgeTone}>
                  {auth.authenticated ?
                    auth.user?.role || "signed in"
                  : auth.message || "guest"}
                </Badge>
              </>
            }
          />

          {activeSection === "runtime" && <RuntimeSection runtime={runtime} />}

          {activeSection === "agents" && <AgentsSection agents={agents} />}

          {activeSection === "memory" && (
            <MemorySection entries={memoryEntries} />
          )}

          {activeSection === "rules" && <RulesSection />}

          {activeSection === "mcp-tools" && <McpSection servers={mcpServers} />}

          {activeSection === "logs" && (
            <LogsSection
              logs={filteredLogs}
              filter={filterLogs}
              setFilter={setFilterLogs}
            />
          )}

          {activeSection === "evals" && <EvalsSection metrics={evalFallback} />}

          {activeSection === "samples" && <VisualSystemPage />}

          {activeSection === "reference" && (
            <ReferenceSection
              summary={referenceSummary}
              agentCount={agents.length}
              mcpCount={mcpServers.length}
              providerCount={runtime.providers.length}
            />
          )}

          <CommandBar
            value={command}
            onChange={(nextValue) => setCommand(nextValue)}
            onSubmit={() => {
              // Placeholder command surface for the desktop shell.
              setCommand("");
            }}
            prompt='/command'
            status={daemonConnected ? "sys_ok" : "offline"}
            hint={COMMAND_HINTS[0]}
          />
        </main>
      </div>
    </div>
  );
}

function AuthSection({
  auth,
  form,
  onSubmit,
  onOidcLogin,
  onLogout,
  onEmailChange,
  onPasswordChange,
}: {
  auth: DesktopAuthStatus;
  form: AuthFormState;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onOidcLogin: () => void | Promise<void>;
  onLogout: () => void;
  onEmailChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
}) {
  return (
    <Panel
      id='auth-panel'
      title={auth.authenticated ? "Signed in" : "Login"}
      description={
        auth.authenticated ?
          "Your session is active. The desktop and CLI can now share the same bearer token."
        : auth.required ?
          "Authenticate to unlock private infrastructure, tools, and workspace operations."
        : "Authentication is optional, but signing in unlocks the private sections and shared session state."

      }
      actions={
        auth.authenticated ?
          <Button variant='secondary' onClick={onLogout}>
            Logout
          </Button>
        : <Badge tone={auth.required ? "warning" : "neutral"}>
            {auth.message || "guest mode"}
          </Badge>
      }
    >
      {auth.authenticated ?
        <div className='grid-layout grid-layout--2'>
          <Card tone='raised' className='stack'>
            <div className='timeline-card__header'>
              <div>
                <h3 className='timeline-card__title'>Current account</h3>
                <p className='timeline-card__meta'>
                  shared with daemon and CLI
                </p>
              </div>
              <Badge tone='success'>{auth.user?.role || "signed in"}</Badge>
            </div>
            <div className='stack stack--dense'>
              <div>
                <strong>{auth.user?.name || "Brain user"}</strong>
              </div>
              <div className='utility-secondary'>{auth.user?.email || ""}</div>
              <div className='utility-secondary'>mode: {auth.mode}</div>
              <div className='utility-secondary'>
                active sessions: {auth.activeSessions}
              </div>
              <div className='utility-secondary'>
                expires: {auth.sessionExpiresAt || "—"}
              </div>
            </div>
          </Card>

          <Card tone='default' className='stack'>
            <div className='timeline-card__header'>
              <div>
                <h3 className='timeline-card__title'>Capabilities</h3>
                <p className='timeline-card__meta'>what this identity can do</p>
              </div>
              <Badge tone='accent'>{auth.capabilities.length} granted</Badge>
            </div>
            <div className='utility-inline utility-wrap'>
              {auth.capabilities.length === 0 ?
                <Badge tone='neutral'>no capabilities reported</Badge>
              : auth.capabilities.map((capability) => (
                  <Badge key={capability} tone='neutral'>
                    {capability}
                  </Badge>
                ))
              }
            </div>
            <div className='stack stack--dense'>
              <div className='utility-secondary'>visible sections</div>
              <div className='utility-inline utility-wrap'>
                {auth.allowedSections.map((section) => (
                  <Badge key={section} tone='accent'>
                    {section}
                  </Badge>
                ))}
              </div>
            </div>
          </Card>
        </div>
      : <div className='stack stack--dense'>
          {auth.mode === "oidc" ?
            <div className='stack stack--dense'>
              <label className='field'>
                <span className='field__label'>Email hint</span>
                <Input
                  type='email'
                  value={form.email}
                  onChange={(event) => onEmailChange(event.target.value)}
                  autoComplete='username'
                  placeholder='you@company.com'
                />
              </label>
              <div className='utility-secondary'>
                Passwords are managed by the identity provider.
              </div>
              {form.success && <Badge tone='success'>{form.success}</Badge>}
              {form.error && <Badge tone='danger'>{form.error}</Badge>}
              <div className='utility-inline utility-spread'>
                <Badge tone={auth.required ? "warning" : "neutral"}>
                  {auth.required ? "login required" : "login available"}
                </Badge>
                <Button
                  variant='primary'
                  type='button'
                  disabled={form.busy}
                  onClick={() => void onOidcLogin()}
                >
                  {form.busy ? "Opening provider..." : "Continue with provider"}
                </Button>
              </div>
            </div>
          : <form className='stack stack--dense' onSubmit={onSubmit}>
              <div className='grid-layout grid-layout--2'>
                <label className='field'>
                  <span className='field__label'>Email</span>
                  <Input
                    type='email'
                    value={form.email}
                    onChange={(event) => onEmailChange(event.target.value)}
                    autoComplete='username'
                    placeholder='owner@brain.local'
                  />
                </label>

                <label className='field'>
                  <span className='field__label'>Password</span>
                  <Input
                    type='password'
                    value={form.password}
                    onChange={(event) => onPasswordChange(event.target.value)}
                    autoComplete='current-password'
                    placeholder='••••••••'
                  />
                </label>
              </div>

              {form.success && <Badge tone='success'>{form.success}</Badge>}
              {form.error && <Badge tone='danger'>{form.error}</Badge>}

              <div className='utility-inline utility-spread'>
                <Badge tone={auth.required ? "warning" : "neutral"}>
                  {auth.required ? "login required" : "guest + login available"}
                </Badge>
                <Button variant='primary' type='submit' disabled={form.busy}>
                  {form.busy ? "Signing in..." : "Sign in"}
                </Button>
              </div>
            </form>
          }
        </div>
      }
    </Panel>
  );
}

function RuntimeSection({ runtime }: { runtime: RuntimeState }) {
  return (
    <div className='stack stack--loose'>
      <div className='grid-layout grid-layout--3'>
        <MetricCard
          label='daemon status'
          value={runtime.status || "—"}
          detail={runtime.environment || "No environment reported yet"}
          tone={runtime.connected ? "accent" : "default"}
        />
        <MetricCard
          label='processes'
          value={runtime.processes > 0 ? String(runtime.processes) : "—"}
          detail='reported by /api/status'
          tone='default'
        />
        <MetricCard
          label='active agents'
          value={runtime.activeAgents > 0 ? String(runtime.activeAgents) : "—"}
          detail='reported by /api/agents'
          tone='default'
        />
        <MetricCard
          label='sync status'
          value={runtime.syncStatus || "—"}
          detail={runtime.syncRunning ? "sync running" : "sync idle"}
          tone='default'
        />
        <MetricCard
          label='available providers'
          value={
            runtime.providers.length > 0 ?
              String(runtime.providers.length)
            : "—"
          }
          detail='reported by /api/providers/available'
          tone='default'
        />
        <MetricCard
          label='sync last run'
          value={runtime.syncLastRun || "—"}
          detail={runtime.syncError || "No sync error reported"}
          tone='default'
        />
      </div>

      <div className='grid-layout grid-layout--2'>
        <Panel
          title='Available providers'
          description='Loaded directly from /api/providers/available.'
        >
          {runtime.providers.length === 0 ?
            <div className='empty-state'>
              <div className='empty-state__title'>
                No providers reported yet
              </div>
              <p>The daemon has not exposed any available providers yet.</p>
            </div>
          : <div className='stack stack--dense'>
              {runtime.providers.map((provider) => (
                <div
                  key={provider.name}
                  className='utility-inline utility-spread'
                >
                  <div>
                    <strong>{provider.name}</strong>
                    <div className='utility-secondary'>{provider.status}</div>
                  </div>
                  <Badge tone={provider.tone}>{provider.status}</Badge>
                </div>
              ))}
            </div>
          }
        </Panel>

        <Panel
          title='Daemon payload'
          description='Current values returned by /api/status.'
        >
          <CodeBlock subtitle='status payload'>
            {JSON.stringify(
              {
                status: runtime.status || null,
                environment: runtime.environment || null,
                processes: runtime.processes || null,
                sync_status: runtime.syncStatus || null,
                sync_running: runtime.syncRunning,
                sync_last_run: runtime.syncLastRun || null,
                sync_error: runtime.syncError || null,
              },
              null,
              2,
            )}
          </CodeBlock>
        </Panel>
      </div>
    </div>
  );
}

function AgentsSection({ agents }: { agents: AgentState[] }) {
  return (
    <div className='stack stack--loose'>
      {agents.length === 0 ?
        <Panel
          title='Agents'
          description='No agents are exposed yet by the daemon or canonical artifact tree.'
        >
          <div className='empty-state'>
            <div className='empty-state__title'>No agent data yet</div>
            <p>
              The agents API returned an empty list, so this section stays empty
              instead of fabricating cards.
            </p>
          </div>
        </Panel>
      : <div className='grid-layout grid-layout--2'>
          {agents.map((agent) => (
            <Card key={agent.id} tone='default'>
              <div className='timeline-card__header'>
                <div>
                  <h3 className='timeline-card__title'>
                    {agent.name || agent.id}
                  </h3>
                  <p className='timeline-card__meta'>{agent.id}</p>
                </div>
                <Badge tone={agent.maintained ? "success" : "warning"}>
                  {agent.maintained ? "maintained" : "not maintained"}
                </Badge>
              </div>
              {agent.description && (
                <p className='reference-block__description'>
                  {agent.description}
                </p>
              )}
              <div className='stack stack--dense stack--offset-sm'>
                <div className='utility-secondary'>
                  version: {agent.version || "—"}
                </div>
                <div className='utility-secondary'>
                  model: {agent.model || "—"}
                </div>
                <div className='utility-secondary'>
                  temperature: {agent.temperature.toFixed(2)}
                </div>
                <div className='utility-secondary'>
                  prompt file: {agent.promptFile || "—"}
                </div>
                <div className='utility-inline'>
                  {agent.tags.length === 0 ?
                    <Badge tone='neutral'>no tags</Badge>
                  : agent.tags.map((tag) => (
                      <Badge key={tag} tone='neutral'>
                        {tag}
                      </Badge>
                    ))
                  }
                </div>
              </div>
            </Card>
          ))}
        </div>
      }
    </div>
  );
}

function MemorySection({ entries }: { entries: MemoryEntry[] }) {
  return (
    <Panel
      title='Memory'
      description='No memory API exists yet, so this view stays empty instead of inventing entries.'
    >
      {entries.length === 0 ?
        <div className='empty-state'>
          <div className='empty-state__title'>No memory data yet</div>
          <p>
            When a real memory endpoint lands, the desktop can render it here
            directly.
          </p>
        </div>
      : <div className='timeline'>
          {entries.map((entry) => (
            <div
              key={`${entry.time}-${entry.label}`}
              className='timeline__item'
            >
              <div className='timeline__time'>
                <div>{entry.time}</div>
                <div className='utility-secondary'>today</div>
              </div>
              <div className='timeline__marker' />
              <Card
                tone={entry.kind === "json" ? "accent" : "default"}
                className='timeline-card'
              >
                <div className='timeline-card__header'>
                  <div>
                    <h3 className='timeline-card__title'>{entry.label}</h3>
                    <p className='timeline-card__meta'>{entry.kind}</p>
                  </div>
                </div>
                <CodeBlock>{entry.summary}</CodeBlock>
                <div className='timeline-card__footer'>
                  {entry.tags.map((tag) => (
                    <Badge
                      key={tag}
                      tone={
                        tag.includes("persistent") || tag.includes("semantic") ?
                          "accent"
                        : "neutral"
                      }
                    >
                      {tag}
                    </Badge>
                  ))}
                </div>
              </Card>
            </div>
          ))}
        </div>
      }
    </Panel>
  );
}

function RulesSection() {
  const validation = rulesValidationFallback;

  return (
    <div className='grid-layout grid-layout--2'>
      <Panel
        title='Rules'
        description='The daemon does not expose a rules API yet, so this surface stays empty by design.'
      >
        <div className='empty-state'>
          <div className='empty-state__title'>No rules data yet</div>
          <p>
            When a canonical rules source becomes available, it can replace this
            empty state.
          </p>
        </div>
      </Panel>

      <Panel
        title='Rules metadata'
        description='Placeholder metadata until a real rules source is connected.'
      >
        <div className='stack stack--dense'>
          <MetricCard
            label='version'
            value={validation.version || "—"}
            detail='not connected'
            tone='default'
          />
          <div className='stack stack--dense'>
            <div className='utility-secondary'>history</div>
            <div className='utility-inline'>
              {validation.history.length === 0 ?
                <Badge tone='neutral'>no history</Badge>
              : validation.history.map((item) => (
                  <Badge key={item} tone='neutral'>
                    {item}
                  </Badge>
                ))
              }
            </div>
          </div>
          <div className='stack stack--dense'>
            <div className='utility-secondary'>validation</div>
            <div className='utility-inline'>
              {validation.validation.length === 0 ?
                <Badge tone='neutral'>no checks</Badge>
              : validation.validation.map((item) => (
                  <Badge key={item} tone='success'>
                    {item}
                  </Badge>
                ))
              }
            </div>
          </div>
          <Button variant='secondary' disabled>
            Rollback previous
          </Button>
        </div>
      </Panel>
    </div>
  );
}

function McpSection({ servers }: { servers: McpServer[] }) {
  return (
    <div className='grid-layout grid-layout--2'>
      <Panel
        title='Connected MCP servers'
        description='Tool servers, transport, and connection states.'
      >
        {servers.length === 0 ?
          <div className='empty-state'>
            <div className='empty-state__title'>
              No MCP servers reported yet
            </div>
            <p>The daemon is not exposing any servers right now.</p>
          </div>
        : <table className='data-table'>
            <thead>
              <tr>
                <th>Server</th>
                <th>Status</th>
                <th>Tools</th>
                <th>Transport</th>
              </tr>
            </thead>
            <tbody>
              {servers.map((server) => (
                <tr key={server.id}>
                  <td>
                    <div className='stack stack--dense'>
                      <strong>{server.name}</strong>
                      <span className='data-table__muted'>{server.id}</span>
                    </div>
                  </td>
                  <td>
                    <Badge tone={server.tone}>{server.status}</Badge>
                  </td>
                  <td>{server.tools}</td>
                  <td className='data-table__muted'>{server.transport}</td>
                </tr>
              ))}
            </tbody>
          </table>
        }
      </Panel>

      <Card tone='raised' className='stack'>
        <div className='timeline-card__header'>
          <div>
            <h3 className='timeline-card__title'>Node health</h3>
            <p className='timeline-card__meta'>
              live counts from the MCP registry
            </p>
          </div>
          <Badge tone={servers.length > 0 ? "success" : "warning"}>
            {servers.length > 0 ? `${servers.length} servers` : "no servers"}
          </Badge>
        </div>
        <MetricCard
          label='active tools'
          value={String(servers.reduce((sum, server) => sum + server.tools, 0))}
          detail='sum of reported tools'
          tone='accent'
        />
        <MetricCard
          label='clients'
          value={String(
            servers.reduce((sum, server) => sum + server.clientCount, 0),
          )}
          detail='sum of connected clients'
          tone='default'
        />
        <div className='stack stack--dense'>
          <div className='utility-secondary'>server statuses</div>
          {servers.length === 0 ?
            <Badge tone='neutral'>no servers</Badge>
          : servers.slice(0, 3).map((server) => (
              <div key={server.id} className='utility-inline utility-spread'>
                <span>{server.name}</span>
                <Badge tone={server.tone}>{server.status}</Badge>
              </div>
            ))
          }
        </div>
      </Card>
    </div>
  );
}

function LogsSection({
  logs,
  filter,
  setFilter,
}: {
  logs: LogEntry[];
  filter: string;
  setFilter: (value: string) => void;
}) {
  return (
    <div className='stack stack--loose'>
      <Card tone='default' className='stack'>
        <div className='editor-pane__header'>
          <div>
            <div className='utility-secondary'>system stream</div>
            <h3 className='panel__title'>Live activity</h3>
          </div>
          <Button variant='secondary'>Export log</Button>
        </div>
        <Input
          placeholder='Filter logs...'
          value={filter}
          onChange={(event) => setFilter(event.target.value)}
        />
      </Card>

      <Card tone='default' className='code-block'>
        <div className='code-block__header'>
          <p className='code-block__title'>system_activity.log</p>
          <p className='code-block__subtitle'>live stream active</p>
        </div>
        <div className='code-block__body log-stream'>
          {logs.length === 0 ?
            <div className='empty-state'>
              <div className='empty-state__title'>No log events yet</div>
              <p>
                The WebSocket stream will populate this panel when the daemon
                emits events.
              </p>
            </div>
          : logs.map((entry) => (
              <div
                key={`${entry.time}-${entry.channel}-${entry.message}`}
                className='utility-inline log-stream__entry'
              >
                <span className='data-table__muted log-stream__time'>
                  {entry.time}
                </span>
                <Badge tone={entry.tone}>{entry.channel}</Badge>
                <span>{entry.message}</span>
              </div>
            ))
          }
        </div>
      </Card>
    </div>
  );
}

function EvalsSection({ metrics }: { metrics: EvalMetric[] }) {
  return (
    <div className='stack stack--loose'>
      {metrics.length === 0 ?
        <Panel
          title='Evaluations'
          description='No evaluation endpoint is exposed yet.'
        >
          <div className='empty-state'>
            <div className='empty-state__title'>No evaluation data yet</div>
            <p>
              The daemon does not provide evaluation metrics, so this surface
              stays blank.
            </p>
          </div>
        </Panel>
      : <div className='grid-layout grid-layout--4'>
          {metrics.map((metric) => (
            <MetricCard
              key={metric.label}
              label={metric.label}
              value={metric.value}
              detail={metric.detail}
              tone={metric.tone === "accent" ? "accent" : "default"}
            >
              <ProgressBar value={metric.progress} tone={metric.tone} />
            </MetricCard>
          ))}
        </div>
      }
    </div>
  );
}

function ReferenceSection({
  summary,
  agentCount,
  mcpCount,
  providerCount,
}: {
  summary: ReferenceSummary;
  agentCount: number;
  mcpCount: number;
  providerCount: number;
}) {
  return (
    <div className='grid-layout grid-layout--2'>
      <Panel
        title='Implementation reference'
        description='The living support surfaces for desktop work.'
      >
        <div className='stack stack--dense'>
          <div className='grid-layout grid-layout--2'>
            <MetricCard
              label='skills'
              value={String(summary.skillsCount)}
              detail='from /api/skills'
              tone='default'
            />
            <MetricCard
              label='agents'
              value={String(agentCount)}
              detail='from /api/agents'
              tone='default'
            />
            <MetricCard
              label='mcp servers'
              value={String(mcpCount)}
              detail='from /api/mcp/servers'
              tone='default'
            />
            <MetricCard
              label='providers'
              value={String(providerCount)}
              detail='from /api/providers/available'
              tone='default'
            />
            <MetricCard
              label='workflows'
              value={String(summary.workflowsCount)}
              detail='from /api/workflows/list'
              tone='default'
            />
            <MetricCard
              label='delegation executions'
              value={String(summary.delegationCount)}
              detail='from /api/delegation/executions'
              tone='default'
            />
          </div>

          <Card tone='default' className='stack'>
            <div className='timeline-card__header'>
              <div>
                <h3 className='timeline-card__title'>Docs status</h3>
                <p className='timeline-card__meta'>
                  real output from /api/docs/status
                </p>
              </div>
              <Badge tone={summary.docsStatus ? "success" : "warning"}>
                {summary.docsStatus ? "loaded" : "empty"}
              </Badge>
            </div>
            {summary.docsStatus ?
              <CodeBlock>{summary.docsStatus}</CodeBlock>
            : <div className='empty-state'>
                <div className='empty-state__title'>No docs status yet</div>
                <p>The docs index has not reported status data yet.</p>
              </div>
            }
          </Card>
        </div>
      </Panel>

      <Card tone='raised' className='stack'>
        <div className='timeline-card__header'>
          <div>
            <h3 className='timeline-card__title'>AutoEvolve status</h3>
            <p className='timeline-card__meta'>
              real output from /api/autoevolve/status
            </p>
          </div>
          <Badge tone={summary.autoEvolveStatus ? "success" : "warning"}>
            {summary.autoEvolveStatus ? "loaded" : "empty"}
          </Badge>
        </div>
        {summary.autoEvolveStatus ?
          <CodeBlock>{summary.autoEvolveStatus}</CodeBlock>
        : <div className='empty-state'>
            <div className='empty-state__title'>No AutoEvolve status yet</div>
            <p>The daemon has not reported AutoEvolve state yet.</p>
          </div>
        }
      </Card>
    </div>
  );
}

export default App;
