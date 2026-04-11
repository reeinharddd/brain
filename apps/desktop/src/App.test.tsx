import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  it("renders header with Brain Daemon title", () => {
    render(<App />);
    expect(
      screen.getAllByRole("heading", { name: "Brain Daemon", level: 1 }).length,
    ).toBeGreaterThan(0);
  });

  it("renders all 5 tabs", () => {
    render(<App />);
    const allButtons = screen.getAllByRole("button");
    const tabNames = allButtons.map((b) => b.textContent);
    ["Status", "Artifacts", "Context", "Policy", "Events"].forEach((tab) => {
      expect(tabNames).toContain(tab);
    });
  });

  it("shows status view by default with 15 subsystems", () => {
    render(<App />);
    const tables = screen.getAllByRole("table");
    expect(tables.length).toBeGreaterThan(0);
    const subsystemTexts = screen.getAllByText("15 subsystems operational");
    expect(subsystemTexts.length).toBeGreaterThan(0);
  });
});
