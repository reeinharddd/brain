import { DESKTOP_THEME_STORAGE_KEY } from "./tokens";
import type { ThemeMode } from "./tokens";

export function resolveInitialTheme(): ThemeMode {
  if (typeof window === "undefined") {
    return "dark";
  }

  const storedTheme = window.localStorage.getItem(DESKTOP_THEME_STORAGE_KEY);
  if (storedTheme === "dark" || storedTheme === "light") {
    return storedTheme;
  }

  const prefersLight = window.matchMedia?.("(prefers-color-scheme: light)").matches;
  return prefersLight ? "light" : "dark";
}

export function applyTheme(theme: ThemeMode) {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.dataset.theme = theme;

  if (typeof window !== "undefined") {
    window.localStorage.setItem(DESKTOP_THEME_STORAGE_KEY, theme);
  }
}

export function toggleTheme(theme: ThemeMode): ThemeMode {
  return theme === "dark" ? "light" : "dark";
}
