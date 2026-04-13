import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, within } from "@testing-library/react";
import App from "./DesktopApp";

type FetchMode = "bootstrap" | "oidc";

function mockDaemonFetch(mode: FetchMode, authenticated = false) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: RequestInfo | URL) => {
      const url = String(input);
      if (url.includes("/api/auth/status")) {
        return jsonResponse({
          required: true,
          mode,
          authenticated,
          active_sessions: authenticated ? 1 : 0,
          message: authenticated ? "logged in" : "authentication required",
          user: authenticated
            ? {
                id: "owner-brain-local",
                email: "owner@brain.local",
                name: "Brain Owner",
                role: "owner",
                capabilities: ["auth:manage", "infra:read"],
                sections: ["runtime", "agents"],
              }
            : null,
          session: authenticated
            ? {
                expires_at: "2026-04-13T00:00:00Z",
                last_used: "2026-04-12T12:00:00Z",
              }
            : null,
          allowed_sections: authenticated ? ["runtime", "agents"] : ["runtime", "samples", "reference"],
          capabilities: authenticated ? ["auth:manage", "infra:read"] : [],
        });
      }

      if (url.includes("/api/auth/login")) {
        return jsonResponse({
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
        });
      }

      if (url.includes("/api/auth/oidc/start")) {
        return jsonResponse({
          success: true,
          provider: "logto",
          state: "oidc-state",
          authorization_url: "https://idp.example/login",
          expires_at: "2026-04-13T00:00:00Z",
        });
      }

      if (url.includes("/api/auth/oidc/poll")) {
        return jsonResponse({
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
              provider: "logto",
              subject: "subject-123",
            },
            message: "login complete",
          },
        });
      }

      if (url.includes("/api/auth/refresh")) {
        return jsonResponse({
          success: true,
          token: "refreshed-token",
          refresh_token: "refreshed-refresh",
          expires_at: "2026-04-14T00:00:00Z",
          refresh_expires_at: "2026-05-14T00:00:00Z",
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
        });
      }

      if (url.includes("/api/status")) {
        return jsonResponse({ status: "Running", environment: "development", processes: 1, sync_status: "idle", sync_running: false, sync_last_run: "", sync_error: "" });
      }

      if (url.includes("/api/providers/available")) {
        return jsonResponse({ available: [] });
      }

      if (url.includes("/api/agents")) {
        return jsonResponse({ agents: [] });
      }

      if (url.includes("/api/mcp/servers")) {
        return jsonResponse({ servers: [] });
      }

      if (url.includes("/api/skills")) {
        return jsonResponse({ skills: [] });
      }

      if (url.includes("/api/docs/status")) {
        return jsonResponse({ status: "ready" });
      }

      if (url.includes("/api/workflows/list")) {
        return jsonResponse({ workflows: [] });
      }

      if (url.includes("/api/autoevolve/status")) {
        return jsonResponse({ status: "idle" });
      }

      if (url.includes("/api/delegation/executions")) {
        return jsonResponse({ executions: [] });
      }

      return jsonResponse({ ok: true });
    }),
  );
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), { status, headers: { "Content-Type": "application/json" } });
}

beforeEach(() => {
  vi.unstubAllGlobals();
  Reflect.deleteProperty(globalThis as Record<string, unknown>, "__brainDesktopAuthFetchInstalled");
});

afterEach(() => {
  vi.unstubAllGlobals();
  Reflect.deleteProperty(globalThis as Record<string, unknown>, "__brainDesktopAuthFetchInstalled");
});

describe("App", () => {
  it("shows the login gate before a session exists", async () => {
    mockDaemonFetch("bootstrap", false);
    render(<App />);

    expect(await screen.findByRole("heading", { name: "Login", level: 2 })).toBeDefined();
    expect(screen.queryByRole("heading", { name: "Brain Control Plane", level: 1 })).toBeNull();
  });

  it("renders the OIDC gate when the daemon is in oidc mode", async () => {
    mockDaemonFetch("oidc", false);
    render(<App />);

    expect(await screen.findByRole("button", { name: "Continue with provider" })).toBeDefined();
  });

  it("renders the desktop shell once authenticated", async () => {
    mockDaemonFetch("bootstrap", true);
    render(<App />);

    expect(
      await screen.findByRole("heading", {
        name: "Brain Control Plane",
        level: 1,
      }),
    ).toBeDefined();

    const nav = screen.getAllByLabelText("Desktop sections")[0];
    const tabNames = within(nav).getAllByRole("button").map((button) => button.textContent);
    [
      "Runtime",
      "Agents",
      "Memory",
      "Rules",
      "MCP Tools",
      "Logs",
      "Evals",
      "Samples",
      "Reference",
    ].forEach((tab) => {
      expect(tabNames.some((name) => name?.includes(tab))).toBe(true);
    });
  });

  it("shows the runtime screen by default and can switch to the samples gallery", async () => {
    mockDaemonFetch("bootstrap", true);
    render(<App />);

    expect(
      await screen.findByRole("heading", {
        name: "Runtime",
        level: 2,
      }),
    ).toBeDefined();

    fireEvent.click(screen.getAllByRole("button", { name: /Samples/i })[0]);

    expect(
      await screen.findByRole("heading", {
        name: "Design samples",
        level: 2,
      }),
    ).toBeDefined();
  });
});
