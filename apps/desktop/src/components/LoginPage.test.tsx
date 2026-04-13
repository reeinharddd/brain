import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen } from "@testing-library/react";
import LoginPage from "./LoginPage";

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function jsonResponseError(message: string, status = 401) {
  return new Response(JSON.stringify({ error: message }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function isDisabled(element: HTMLElement): boolean {
  return element.hasAttribute("disabled");
}

function isEnabled(element: HTMLElement): boolean {
  return !element.hasAttribute("disabled");
}

beforeEach(() => {
  vi.unstubAllGlobals();
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("LoginPage", () => {
  it("renders the login page with all required elements", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ ok: true })),
    );

    render(<LoginPage />);

    expect(screen.getByText("brain")).toBeDefined();
    expect(screen.getByLabelText("Email")).toBeDefined();
    expect(screen.getByLabelText("Password")).toBeDefined();
    expect(screen.getByRole("button", { name: "Sign in" })).toBeDefined();
    expect(screen.getByText("or continue with")).toBeDefined();
    expect(
      screen.getByRole("button", { name: "Sign in with Google" }),
    ).toBeDefined();
    expect(
      screen.getByRole("button", { name: "Sign in with GitHub" }),
    ).toBeDefined();
    expect(
      screen.getByRole("button", { name: "Sign in with Microsoft" }),
    ).toBeDefined();
  });

  it("disables the sign in button when form is empty", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ ok: true })),
    );

    render(<LoginPage />);

    const signInButton = screen.getByRole("button", { name: "Sign in" });
    expect(isDisabled(signInButton)).toBe(true);
  });

  it("enables the sign in button when email and password are filled", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ ok: true })),
    );

    render(<LoginPage />);

    const emailInput = screen.getByLabelText("Email");
    const passwordInput = screen.getByLabelText("Password");

    fireEvent.change(emailInput, { target: { value: "test@example.com" } });
    fireEvent.change(passwordInput, { target: { value: "password123" } });

    const signInButton = screen.getByRole("button", { name: "Sign in" });
    expect(isEnabled(signInButton)).toBe(true);
  });

  it("shows an error badge when login fails", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/auth/login")) {
          return jsonResponseError("invalid credentials", 401);
        }
        return jsonResponse({ ok: true });
      }),
    );

    render(<LoginPage />);

    const emailInput = screen.getByLabelText("Email");
    const passwordInput = screen.getByLabelText("Password");

    fireEvent.change(emailInput, { target: { value: "test@example.com" } });
    fireEvent.change(passwordInput, { target: { value: "wrong" } });

    const signInButton = screen.getByRole("button", { name: "Sign in" });
    fireEvent.click(signInButton);

    expect(await screen.findByText("invalid credentials")).toBeDefined();
  });

  it("calls onLoginSuccess when login succeeds", async () => {
    const onSuccess = vi.fn();

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/auth/login")) {
          return jsonResponse({
            token: "test-token",
            refresh_token: "test-refresh",
            expires_at: "2026-04-13T00:00:00Z",
            refresh_expires_at: "2026-05-13T00:00:00Z",
            mode: "bootstrap",
            required: true,
            user: {
              id: "test-user",
              email: "test@example.com",
              name: "Test User",
              role: "owner",
              capabilities: ["infra:read"],
              sections: ["runtime"],
            },
            capabilities: ["infra:read"],
            allowed_sections: ["runtime"],
          });
        }
        return jsonResponse({ ok: true });
      }),
    );

    render(<LoginPage onLoginSuccess={onSuccess} />);

    const emailInput = screen.getByLabelText("Email");
    const passwordInput = screen.getByLabelText("Password");

    fireEvent.change(emailInput, { target: { value: "test@example.com" } });
    fireEvent.change(passwordInput, { target: { value: "password123" } });

    const signInButton = screen.getByRole("button", { name: "Sign in" });
    fireEvent.click(signInButton);

    expect(
      await screen.findByText("Signed in successfully. Redirecting..."),
    ).toBeDefined();
    expect(onSuccess).toHaveBeenCalledTimes(1);
  });

  it("disables provider buttons during busy state", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes("/api/auth/login")) {
          await new Promise((r) => setTimeout(r, 200));
          return jsonResponseError("timeout", 500);
        }
        return jsonResponse({ ok: true });
      }),
    );

    render(<LoginPage />);

    const emailInput = screen.getByLabelText("Email");
    const passwordInput = screen.getByLabelText("Password");

    fireEvent.change(emailInput, { target: { value: "test@example.com" } });
    fireEvent.change(passwordInput, { target: { value: "password123" } });

    const signInButton = screen.getByRole("button", { name: "Sign in" });
    fireEvent.click(signInButton);

    expect(
      isDisabled(screen.getByRole("button", { name: "Sign in with Google" })),
    ).toBe(true);
    expect(
      isDisabled(screen.getByRole("button", { name: "Sign in with GitHub" })),
    ).toBe(true);
    expect(
      isDisabled(
        screen.getByRole("button", { name: "Sign in with Microsoft" }),
      ),
    ).toBe(true);
  });

  it("renders email input with correct autocomplete", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ ok: true })),
    );

    render(<LoginPage />);

    const emailInput = screen.getByLabelText("Email");
    expect(emailInput.getAttribute("autocomplete")).toBe("username");
    expect(emailInput.getAttribute("type")).toBe("email");
  });

  it("renders password input with correct autocomplete", () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse({ ok: true })),
    );

    render(<LoginPage />);

    const passwordInput = screen.getByLabelText("Password");
    expect(passwordInput.getAttribute("autocomplete")).toBe("current-password");
    expect(passwordInput.getAttribute("type")).toBe("password");
  });
});
