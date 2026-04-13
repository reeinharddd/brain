import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
vi.mock("@tauri-apps/api/core", () => ({
  invoke: vi.fn(async (command: string) => {
    if (command === "open_external_url") {
      return undefined;
    }
    throw new Error("tauri unavailable in tests");
  }),
}));
import {
  clearStoredAuthSession,
  fetchDesktopAuthStatus,
  installDesktopAuthFetchInterceptor,
  loginToDaemon,
  normalizeAuthStatus,
  readStoredAuthSession,
  saveStoredAuthSession,
  loginWithOidc,
} from "./auth";

describe("desktop auth api", () => {
  beforeEach(async () => {
    await clearStoredAuthSession();
    vi.unstubAllGlobals();
    Reflect.deleteProperty(globalThis as Record<string, unknown>, "__brainDesktopAuthFetchInstalled");
  });

  afterEach(async () => {
    await clearStoredAuthSession();
    vi.unstubAllGlobals();
    Reflect.deleteProperty(globalThis as Record<string, unknown>, "__brainDesktopAuthFetchInstalled");
  });

  it("normalizes daemon auth status payloads", () => {
    const status = normalizeAuthStatus({
      required: true,
      mode: "bootstrap",
      authenticated: true,
      active_sessions: 2,
      message: "logged in",
      user: {
        id: "owner-brain-local",
        email: "owner@brain.local",
        name: "Brain Owner",
        role: "owner",
        capabilities: ["auth:manage", "infra:read"],
        sections: ["runtime", "agents"],
      },
      session: {
        expires_at: "2026-04-13T00:00:00Z",
        last_used: "2026-04-12T12:00:00Z",
      },
      allowed_sections: ["runtime", "agents"],
      capabilities: ["auth:manage", "infra:read"],
    });

    expect(status.authenticated).toBe(true);
    expect(status.user?.email).toBe("owner@brain.local");
    expect(status.allowedSections).toEqual(["runtime", "agents"]);
    expect(status.capabilities).toContain("auth:manage");
    expect(status.sessionExpiresAt).toBe("2026-04-13T00:00:00Z");
  });

  it("stores the token after a successful login", async () => {
    const loginResponse = {
      success: true,
      mode: "bootstrap",
      required: true,
      token: "token-123",
      refresh_token: "refresh-123",
      expires_at: "2026-04-13T00:00:00Z",
      refresh_expires_at: "2026-05-13T00:00:00Z",
      user: {
        id: "owner-brain-local",
        email: "owner@brain.local",
        name: "Brain Owner",
        role: "owner",
        capabilities: ["auth:manage", "infra:read"],
        sections: ["runtime", "agents"],
      },
      capabilities: ["auth:manage", "infra:read"],
      allowed_sections: ["runtime", "agents"],
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(JSON.stringify(loginResponse), { status: 200, headers: { "Content-Type": "application/json" } })),
    );

    const status = await loginToDaemon("owner@brain.local", "secret123");

    expect(status.authenticated).toBe(true);
    expect(status.user?.email).toBe("owner@brain.local");
    expect((await readStoredAuthSession())?.token).toBe("token-123");
  });

  it("attaches the stored bearer token to daemon requests", async () => {
    await saveStoredAuthSession({
      token: "token-abc",
      refreshToken: "refresh-abc",
      expiresAt: "2026-04-13T00:00:00Z",
      refreshExpiresAt: "2026-05-13T00:00:00Z",
      user: {
        id: "owner-brain-local",
        email: "owner@brain.local",
        name: "Brain Owner",
        role: "owner",
        capabilities: ["auth:manage"],
        sections: ["runtime"],
      },
      mode: "bootstrap",
      required: true,
      allowedSections: ["runtime"],
      capabilities: ["auth:manage"],
    });

    const fetchMock = vi.fn(async (request: Request) => {
      expect(request.headers.get("Authorization")).toBe("Bearer token-abc");
      return new Response(JSON.stringify({ authenticated: true }), { status: 200, headers: { "Content-Type": "application/json" } });
    });

    vi.stubGlobal("fetch", fetchMock);
    installDesktopAuthFetchInterceptor();

    const response = await fetch(`${"http://localhost:9090"}/api/status`);

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(response.ok).toBe(true);
  });

  it("falls back to an unauthenticated status when the daemon is unreachable", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => {
      throw new Error("network error");
    }));

    const status = await fetchDesktopAuthStatus();

    expect(status.authenticated).toBe(false);
    expect(status.allowedSections).toContain("runtime");
  });

  it("completes the oidc login handoff and stores the session", async () => {
    const fetchMock = vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/api/auth/oidc/start")) {
        return new Response(JSON.stringify({
          success: true,
          state: "oidc-state",
          authorization_url: "https://idp.example/login",
          expires_at: "2026-04-13T00:00:00Z",
        }), { status: 200, headers: { "Content-Type": "application/json" } });
      }

      if (url.includes("/api/auth/oidc/poll")) {
        return new Response(JSON.stringify({
          ready: true,
          state: "oidc-state",
          session: {
            success: true,
            state: "oidc-state",
            token: "oidc-token-123",
            refresh_token: "oidc-refresh-123",
            expires_at: "2026-04-13T00:00:00Z",
            refresh_expires_at: "2026-05-13T00:00:00Z",
            user: {
              id: "owner-brain-local",
              email: "owner@brain.local",
              name: "Brain Owner",
              role: "owner",
              capabilities: ["auth:manage", "infra:read"],
              sections: ["runtime", "agents"],
            },
          },
        }), { status: 200, headers: { "Content-Type": "application/json" } });
      }

      return new Response(JSON.stringify({ ok: true }), { status: 200, headers: { "Content-Type": "application/json" } });
    });

    vi.stubGlobal("fetch", fetchMock);

    const status = await loginWithOidc("owner@brain.local");

    expect(status.authenticated).toBe(true);
    expect(status.user?.email).toBe("owner@brain.local");
    expect((await readStoredAuthSession())?.token).toBe("oidc-token-123");
  });
});
