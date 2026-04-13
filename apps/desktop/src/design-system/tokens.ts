export type ThemeMode = "dark" | "light";

export const DESKTOP_THEME_STORAGE_KEY = "brain.desktop.theme";

export const DESKTOP_SECTIONS = [
  {
    id: "runtime",
    label: "Runtime",
    description: "Runtime status and execution surface",
  },
  {
    id: "agents",
    label: "Agents",
    description: "Active agent pool and capabilities",
  },
  {
    id: "memory",
    label: "Memory",
    description: "Persistent memory timeline and recall",
  },
  {
    id: "rules",
    label: "Rules",
    description: "Canonical rules editor and validation",
  },
  {
    id: "mcp-tools",
    label: "MCP Tools",
    description: "Connected tool servers and usage",
  },
  {
    id: "logs",
    label: "Logs",
    description: "Live activity and system events",
  },
  {
    id: "evals",
    label: "Evals",
    description: "Evaluation metrics and quality signals",
  },
  {
    id: "samples",
    label: "Samples",
    description: "Dark, light, and mobile visual references",
  },
  {
    id: "reference",
    label: "Reference",
    description: "Docs search, skills, and support surfaces",
  },
] as const;

export type DesktopSectionId = (typeof DESKTOP_SECTIONS)[number]["id"];

export const DESIGN_TOKENS = {
  colors: {
    dark: {
      bgPrimary: "#0A0A0A",
      bgSecondary: "#111111",
      bgTertiary: "#171717",
      borderSubtle: "#232323",
      borderStrong: "#2E2E2E",
      textPrimary: "#F5F5F3",
      textSecondary: "#A1A1A1",
      textMuted: "#6B6B6B",
      accentPrimary: "#7C3AED",
      accentHover: "#8B5CF6",
      accentActive: "#6D28D9",
      success: "#10B981",
      warning: "#F59E0B",
      error: "#EF4444",
      info: "#60A5FA",
    },
    light: {
      bgPrimary: "#F5F5F3",
      bgSecondary: "#FFFFFF",
      bgTertiary: "#EDEDED",
      borderSubtle: "#E0E0E0",
      borderStrong: "#CFCFCF",
      textPrimary: "#0A0A0A",
      textSecondary: "#525252",
      textMuted: "#8A8A8A",
      accentPrimary: "#6D28D9",
      accentHover: "#7C3AED",
      accentActive: "#5B21B6",
      success: "#059669",
      warning: "#B45309",
      error: "#DC2626",
      info: "#2563EB",
    },
  },
  spacing: [4, 8, 12, 16, 20, 24, 32, 40, 48] as const,
  radius: {
    small: 4,
    default: 6,
    large: 8,
  },
  typography: {
    headingFont: "Inter, Space Grotesk, system-ui, sans-serif",
    monoFont: "IBM Plex Mono, SFMono-Regular, Menlo, Consolas, monospace",
    h1: 28,
    h2: 22,
    h3: 18,
    body: 14,
    small: 12,
    micro: 11,
  },
  layout: {
    sidebarWidth: 240,
    contentMaxWidth: 1280,
    topBarHeight: 64,
  },
  motion: {
    snap: "0ms",
    quick: "50ms",
  },
} as const;

export const SCREEN_GROUPS = [
  {
    id: "dark",
    label: "Dark",
    description: "Primary production presentation",
  },
  {
    id: "light",
    label: "Light",
    description: "Alternative high-contrast presentation",
  },
  {
    id: "mobile",
    label: "Mobile",
    description: "Single-column compact presentation",
  },
] as const;

export type ScreenGroupId = (typeof SCREEN_GROUPS)[number]["id"];

export const COMMAND_HINTS = [
  "/runtime status",
  "/agents list",
  "/memory recall",
  "/rules validate",
  "/mcp tools",
  "/logs tail",
  "/evals report",
] as const;
