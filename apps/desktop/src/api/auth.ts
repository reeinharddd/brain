import { invoke } from "@tauri-apps/api/core";

const DEFAULT_DAEMON_URL =
  import.meta.env.VITE_BRAIN_DAEMON_URL || "http://localhost:9090";
const AUTH_STORAGE_FALLBACK_KEY = "brain.desktop.auth.session.fallback";

export const DAEMON_URL = DEFAULT_DAEMON_URL;

export type DesktopAuthUser = {
  id: string;
  email: string;
  name: string;
  role: string;
  capabilities: string[];
  sections: string[];
};

export type DesktopAuthStatus = {
  required: boolean;
  mode: string;
  authenticated: boolean;
  activeSessions: number;
  message: string;
  user: DesktopAuthUser | null;
  sessionExpiresAt: string | null;
  sessionLastUsed: string | null;
  allowedSections: string[];
  capabilities: string[];
};

type StoredAuthSession = {
  token: string;
  refreshToken: string;
  expiresAt: string;
  refreshExpiresAt: string;
  user: DesktopAuthUser;
  mode: string;
  required: boolean;
  allowedSections: string[];
  capabilities: string[];
};

type RawAuthSessionResponse = {
  token?: unknown;
  refresh_token?: unknown;
  expires_at?: unknown;
  expiresAt?: unknown;
  refresh_expires_at?: unknown;
  refreshExpiresAt?: unknown;
  mode?: unknown;
  required?: unknown;
  user?: unknown;
  allowed_sections?: unknown;
  allowedSections?: unknown;
  capabilities?: unknown;
};

type RawAuthStatusResponse = {
  required?: unknown;
  mode?: unknown;
  authenticated?: unknown;
  active_sessions?: unknown;
  message?: unknown;
  user?: unknown;
  session?: unknown;
  allowed_sections?: unknown;
  capabilities?: unknown;
};

export async function readStoredAuthSession(): Promise<StoredAuthSession | null> {
  const raw = await readStoredAuthSessionRaw();
  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw) as StoredAuthSession;
    if (!parsed.token || !parsed.user) {
      return null;
    }
    if (parsed.expiresAt) {
      const expiresAt = new Date(parsed.expiresAt);
      if (
        !Number.isNaN(expiresAt.getTime()) &&
        expiresAt.getTime() <= Date.now()
      ) {
        if (
          parsed.refreshToken &&
          !(
            parsed.refreshExpiresAt &&
            new Date(parsed.refreshExpiresAt).getTime() <= Date.now()
          )
        ) {
          const refreshed = await refreshStoredAuthSession(parsed);
          if (refreshed) {
            return refreshed;
          }
        }
        await clearStoredAuthSession();
        return null;
      }
    }
    return parsed;
  } catch {
    return null;
  }
}

export async function loadStoredAuthToken(): Promise<string> {
  return (await readStoredAuthSession())?.token ?? "";
}

export async function saveStoredAuthSession(
  session: StoredAuthSession,
): Promise<void> {
  const raw = JSON.stringify(session);
  try {
    await saveStoredAuthSessionSecure(raw);
    return;
  } catch {
    // fall through to dev/test storage if secure storage is unavailable
  }

  if (shouldUseFallbackStorage()) {
    await saveStoredAuthSessionFallback(session);
    return;
  }

  throw new Error("unable to save auth session securely");
}

export async function clearStoredAuthSession(): Promise<void> {
  try {
    await clearStoredAuthSessionSecure();
  } catch {
    // ignore and attempt fallback deletion
  }

  if (shouldUseFallbackStorage()) {
    await clearStoredAuthSessionFallback();
  }
}

async function readStoredAuthSessionRaw(): Promise<string | null> {
  try {
    const stored = await invoke<string | null>("load_auth_session");
    if (stored && stored.trim().length > 0) {
      return stored;
    }
  } catch {
    // fall through to the local fallback for development or tests
  }

  if (shouldUseFallbackStorage()) {
    if (typeof window === "undefined") {
      return null;
    }
    const fallback = window.localStorage.getItem(AUTH_STORAGE_FALLBACK_KEY);
    if (fallback && fallback.trim().length > 0) {
      return fallback;
    }
  }

  return null;
}

async function saveStoredAuthSessionSecure(raw: string): Promise<void> {
  await invoke("save_auth_session", { session_json: raw });
}

async function clearStoredAuthSessionSecure(): Promise<void> {
  await invoke("clear_auth_session");
}

async function saveStoredAuthSessionFallback(
  session: StoredAuthSession,
): Promise<void> {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.setItem(
    AUTH_STORAGE_FALLBACK_KEY,
    JSON.stringify(session),
  );
}

async function clearStoredAuthSessionFallback(): Promise<void> {
  if (typeof window === "undefined") {
    return;
  }
  window.localStorage.removeItem(AUTH_STORAGE_FALLBACK_KEY);
}

function shouldUseFallbackStorage(): boolean {
  const mode = import.meta.env.VITE_BRAIN_AUTH_STORAGE_MODE;
  if (mode) {
    return (
      mode.toLowerCase() === "fallback" || mode.toLowerCase() === "localstorage"
    );
  }
  return import.meta.env.DEV;
}

async function refreshStoredAuthSession(
  session: StoredAuthSession,
): Promise<StoredAuthSession | null> {
  if (!session.refreshToken) {
    return null;
  }

  const response = await fetch(`${DAEMON_URL}/api/auth/refresh`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ refresh_token: session.refreshToken }),
  });

  const payload = await parseJson(response);
  const normalized = normalizeLoginResponse(payload as RawAuthSessionResponse);
  if (!response.ok) {
    return null;
  }

  if (!normalized.user) {
    return null;
  }

  const refreshed: StoredAuthSession = {
    token: String(
      payload && typeof payload === "object" && "token" in payload ?
        ((payload as Record<string, unknown>).token ?? "")
      : "",
    ),
    refreshToken: String(
      payload && typeof payload === "object" && "refresh_token" in payload ?
        ((payload as Record<string, unknown>).refresh_token ??
          session.refreshToken)
      : session.refreshToken,
    ),
    expiresAt: normalized.sessionExpiresAt ?? session.expiresAt,
    refreshExpiresAt: String(
      (
        payload &&
          typeof payload === "object" &&
          "refresh_expires_at" in payload
      ) ?
        ((payload as Record<string, unknown>).refresh_expires_at ??
          session.refreshExpiresAt)
      : session.refreshExpiresAt,
    ),
    user: normalized.user,
    mode: normalized.mode,
    required: normalized.required,
    capabilities: normalized.capabilities,
    allowedSections: normalized.allowedSections,
  };

  await saveStoredAuthSession(refreshed);
  return refreshed;
}

export function installDesktopAuthFetchInterceptor(): void {
  if (
    typeof globalThis === "undefined" ||
    typeof globalThis.fetch !== "function"
  ) {
    return;
  }

  const globalScope = globalThis as typeof globalThis & {
    __brainDesktopAuthFetchInstalled?: boolean;
  };
  if (globalScope.__brainDesktopAuthFetchInstalled) {
    return;
  }

  const originalFetch = globalThis.fetch.bind(globalThis);
  globalThis.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
    const request =
      input instanceof Request ? input.clone() : new Request(input, init);

    try {
      if (shouldAttachAuth(request.url)) {
        const token = await loadStoredAuthToken();
        if (token && !request.headers.has("Authorization")) {
          request.headers.set("Authorization", `Bearer ${token}`);
        }
      }
    } catch {
      // Leave the request unchanged if URL parsing fails.
    }

    return originalFetch(request);
  }) as typeof globalThis.fetch;

  globalScope.__brainDesktopAuthFetchInstalled = true;
}

export async function fetchDesktopAuthStatus(): Promise<DesktopAuthStatus> {
  const token = await loadStoredAuthToken();
  const headers = new Headers();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  try {
    const response = await fetch(`${DAEMON_URL}/api/auth/status`, { headers });
    const payload = await parseJson(response);
    const status = normalizeAuthStatus(payload);

    if (!status.authenticated && token) {
      await clearStoredAuthSession();
    }

    return status;
  } catch {
    return defaultAuthStatus("authentication status unavailable");
  }
}

export async function loginToDaemon(
  email: string,
  password: string,
): Promise<DesktopAuthStatus> {
  const response = await fetch(`${DAEMON_URL}/api/auth/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  const payload = await parseJson(response);
  if (!response.ok) {
    const errorPayload = asRecord(payload);
    const message =
      typeof errorPayload?.error === "string" ?
        errorPayload.error
      : `login failed (${response.status})`;
    throw new Error(message);
  }

  const session = normalizeLoginResponse(payload as RawAuthSessionResponse);
  if (!session.user) {
    throw new Error("login response missing user payload");
  }
  const token = toStringValue((payload as RawAuthSessionResponse).token);
  if (!token) {
    throw new Error("login response missing token");
  }
  await saveStoredAuthSession({
    token,
    refreshToken: String(
      (payload as Record<string, unknown>).refresh_token ?? "",
    ),
    expiresAt:
      session.sessionExpiresAt ??
      new Date(Date.now() + 12 * 60 * 60 * 1000).toISOString(),
    refreshExpiresAt: String(
      (payload as Record<string, unknown>).refresh_expires_at ?? "",
    ),
    user: session.user,
    mode: session.mode,
    required: session.required,
    allowedSections: session.allowedSections,
    capabilities: session.capabilities,
  });

  return session;
}

export async function loginWithOidc(
  loginHint = "",
): Promise<DesktopAuthStatus> {
  const startUrl = new URL(`${DAEMON_URL}/api/auth/oidc/start`);
  if (loginHint.trim()) {
    startUrl.searchParams.set("login_hint", loginHint.trim());
  }

  const startResponse = await fetch(startUrl.toString());
  const startPayload = await parseJson(startResponse);
  const start = normalizeOidcStartResponse(startPayload);
  if (
    !startResponse.ok ||
    !start.success ||
    !start.state ||
    !start.authorizationURL
  ) {
    const errorPayload = asRecord(startPayload);
    const message =
      typeof errorPayload?.error === "string" ?
        errorPayload.error
      : "oidc login failed to start";
    throw new Error(message);
  }

  await openExternalUrl(start.authorizationURL);

  const deadline = start.expiresAt ?? Date.now() + 10 * 60 * 1000;
  while (Date.now() < deadline) {
    const pollResponse = await fetch(
      `${DAEMON_URL}/api/auth/oidc/poll?state=${encodeURIComponent(start.state)}`,
    );
    const pollPayload = await parseJson(pollResponse);
    const poll = normalizeOidcPollResponse(pollPayload);

    if (poll.ready && poll.session) {
      const storedSession = normalizeOidcSession(poll.session);
      await saveStoredAuthSession(storedSession);
      return sessionToDesktopAuthStatus(storedSession);
    }

    if (poll.message) {
      // Polling is intentionally quiet; the login screen keeps the user informed.
    }

    await delay(2000);
  }

  throw new Error("oidc login timed out");
}

export type OidcProvider = "google" | "github" | "microsoft";

const OIDC_PROVIDER_HINTS: Record<OidcProvider, string> = {
  google: "google",
  github: "github",
  microsoft: "microsoft",
};

export async function loginWithProvider(
  provider: OidcProvider,
  loginHint = "",
): Promise<DesktopAuthStatus> {
  const hint = loginHint.trim() || OIDC_PROVIDER_HINTS[provider];
  return loginWithOidc(hint);
}

export async function logoutFromDaemon(): Promise<void> {
  const token = await loadStoredAuthToken();
  if (!token) {
    await clearStoredAuthSession();
    return;
  }

  const response = await fetch(`${DAEMON_URL}/api/auth/logout`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    const payload = await parseJson(response);
    const errorPayload = asRecord(payload);
    const message =
      typeof errorPayload?.error === "string" ?
        errorPayload.error
      : `logout failed (${response.status})`;
    throw new Error(message);
  }

  await clearStoredAuthSession();
}

export function normalizeAuthStatus(
  input: RawAuthStatusResponse | unknown,
): DesktopAuthStatus {
  const raw = asRecord(input);
  const session = asRecord(raw?.session);
  const user = toDesktopAuthUser(raw?.user);
  const capabilities = user?.capabilities ?? toStringArray(raw?.capabilities);
  const allowedSections = toStringArray(
    raw?.allowed_sections ?? raw?.allowedSections,
  );

  return {
    required: toBoolean(raw?.required),
    mode: toStringValue(raw?.mode) || "bootstrap",
    authenticated: toBoolean(raw?.authenticated),
    activeSessions: toNumber(raw?.active_sessions),
    message: toStringValue(raw?.message),
    user,
    sessionExpiresAt:
      toStringValue(session?.expires_at ?? session?.expiresAt) || null,
    sessionLastUsed:
      toStringValue(session?.last_used ?? session?.lastUsed) || null,
    allowedSections,
    capabilities,
  };
}

export function normalizeLoginResponse(
  input: RawAuthSessionResponse,
): DesktopAuthStatus {
  const user = toDesktopAuthUser(input.user);
  const capabilities = toStringArray(input.capabilities ?? user?.capabilities);
  const allowedSections = toStringArray(
    input.allowed_sections ?? input.allowedSections ?? user?.sections,
  );
  const expiresAt = toStringValue(input.expires_at ?? input.expiresAt) || null;

  return {
    required: toBoolean(input.required),
    mode: toStringValue(input.mode) || "bootstrap",
    authenticated: true,
    activeSessions: 0,
    message: "logged in",
    user,
    sessionExpiresAt: expiresAt,
    sessionLastUsed: new Date().toISOString(),
    allowedSections,
    capabilities,
  };
}

function normalizeOidcStartResponse(input: unknown): {
  success: boolean;
  state: string;
  authorizationURL: string;
  expiresAt: number;
} {
  const raw = asRecord(input);
  return {
    success: toBoolean(raw?.success),
    state: toStringValue(raw?.state),
    authorizationURL: toStringValue(raw?.authorization_url),
    expiresAt:
      raw?.expires_at ?
        new Date(toStringValue(raw.expires_at)).getTime()
      : Date.now() + 10 * 60 * 1000,
  };
}

function normalizeOidcPollResponse(input: unknown): {
  ready: boolean;
  message: string;
  session: unknown | null;
} {
  const raw = asRecord(input);
  return {
    ready: toBoolean(raw?.ready),
    message: toStringValue(raw?.message),
    session: raw?.session ?? null,
  };
}

function normalizeOidcSession(input: unknown): StoredAuthSession {
  const raw = asRecord(input);
  const user = toDesktopAuthUser(raw?.user) ?? {
    id: "brain-user",
    email: "",
    name: "Brain user",
    role: "viewer",
    capabilities: [],
    sections: [],
  };

  return {
    token: toStringValue(raw?.token),
    refreshToken: toStringValue(raw?.refresh_token),
    expiresAt: toStringValue(raw?.expires_at),
    refreshExpiresAt: toStringValue(raw?.refresh_expires_at),
    user,
    mode: "oidc",
    required: true,
    allowedSections: toStringArray(raw?.user && asRecord(raw?.user)?.sections),
    capabilities: toStringArray(raw?.user && asRecord(raw?.user)?.capabilities),
  };
}

function sessionToDesktopAuthStatus(
  session: StoredAuthSession,
): DesktopAuthStatus {
  return {
    required: session.required,
    mode: session.mode,
    authenticated: true,
    activeSessions: 1,
    message: "logged in",
    user: session.user,
    sessionExpiresAt: session.expiresAt,
    sessionLastUsed: new Date().toISOString(),
    allowedSections: session.allowedSections,
    capabilities: session.capabilities,
  };
}

async function openExternalUrl(url: string): Promise<void> {
  try {
    await invoke("open_external_url", { url });
    return;
  } catch {
    if (typeof window !== "undefined") {
      window.open(url, "_blank", "noopener,noreferrer");
      return;
    }
    throw new Error("unable to open browser");
  }
}

async function delay(ms: number): Promise<void> {
  await new Promise((resolve) => window.setTimeout(resolve, ms));
}

export function defaultAuthStatus(
  message = "authentication unavailable",
): DesktopAuthStatus {
  return {
    required: false,
    mode: "bootstrap",
    authenticated: false,
    activeSessions: 0,
    message,
    user: null,
    sessionExpiresAt: null,
    sessionLastUsed: null,
    allowedSections: ["runtime", "samples", "reference"],
    capabilities: [],
  };
}

async function parseJson(response: Response): Promise<unknown> {
  try {
    return await response.json();
  } catch {
    return {};
  }
}

function shouldAttachAuth(url: string): boolean {
  try {
    const parsed = new URL(
      url,
      typeof window !== "undefined" ? window.location.href : DEFAULT_DAEMON_URL,
    );
    const daemonOrigin = new URL(DEFAULT_DAEMON_URL).origin;
    return (
      parsed.origin === daemonOrigin || parsed.pathname.startsWith("/api/")
    );
  } catch {
    return false;
  }
}

function toDesktopAuthUser(input: unknown): DesktopAuthUser | null {
  const raw = asRecord(input);
  if (!raw) {
    return null;
  }

  return {
    id: toStringValue(raw.id),
    email: toStringValue(raw.email),
    name: toStringValue(raw.name),
    role: toStringValue(raw.role),
    capabilities: toStringArray(raw.capabilities),
    sections: toStringArray(raw.sections),
  };
}

function toStringArray(input: unknown): string[] {
  if (!Array.isArray(input)) {
    return [];
  }

  return input
    .map((value) => toStringValue(value))
    .filter((value) => value.length > 0);
}

function toStringValue(input: unknown): string {
  if (typeof input === "string") {
    return input.trim();
  }
  if (typeof input === "number" || typeof input === "boolean") {
    return String(input);
  }
  return "";
}

function toNumber(input: unknown): number {
  if (typeof input === "number" && Number.isFinite(input)) {
    return input;
  }
  if (typeof input === "string") {
    const parsed = Number(input);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return 0;
}

function toBoolean(input: unknown): boolean {
  if (typeof input === "boolean") {
    return input;
  }
  if (typeof input === "string") {
    return ["true", "1", "yes", "on"].includes(input.toLowerCase().trim());
  }
  return false;
}

function asRecord(input: unknown): Record<string, unknown> | null {
  if (typeof input !== "object" || input === null || Array.isArray(input)) {
    return null;
  }
  return input as Record<string, unknown>;
}
