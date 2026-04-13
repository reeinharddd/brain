import { readFileSync } from "node:fs";
import { join } from "node:path";
import { describe, expect, it } from "vitest";

const desktopRoot = process.cwd();

const guardedFiles = [
  "src/DesktopApp.tsx",
  "src/components/DocsSearch.tsx",
  "src/pages/VisualSystem.tsx",
];

function read(relativePath: string) {
  return readFileSync(join(desktopRoot, relativePath), "utf8");
}

describe("desktop visual system guard", () => {
  it("keeps the new shell free from inline styles", () => {
    for (const file of guardedFiles) {
      const content = read(file);
      expect(content, `${file} should avoid inline styles`).not.toMatch(/style=\{/);
    }
  });

  it("keeps hardcoded palette values out of the guarded desktop files", () => {
    for (const file of guardedFiles) {
      const content = read(file);
      expect(content, `${file} should avoid raw hex colors`).not.toMatch(/#[0-9a-fA-F]{3,6}\b/);
    }
  });
});
