import { Badge, Button, Card, CodeBlock, SectionHeader } from "../design-system";

type SampleGroupId = "dark" | "light" | "mobile";

type SampleItem = {
  id: string;
  title: string;
  description: string;
  group: SampleGroupId;
  image: string;
  note: string;
  codeHref: string;
};

const sampleImage = (relativePath: string) =>
  new URL(`../assets/brain-os-samples/${relativePath}/screen.png`, import.meta.url).href;

const sampleCode = (relativePath: string) =>
  new URL(`../assets/brain-os-samples/${relativePath}/code.html`, import.meta.url).href;

const sampleGroups: Array<{
  id: SampleGroupId;
  label: string;
  description: string;
}> = [
  {
    id: "dark",
    label: "Dark",
    description: "Primary production presentation",
  },
  {
    id: "light",
    label: "Light",
    description: "High-contrast light presentation",
  },
  {
    id: "mobile",
    label: "Mobile",
    description: "Compact single-column presentation",
  },
];

const samples: SampleItem[] = [
  {
    id: "runtime-dark",
    title: "Runtime",
    description: "Main runtime console and system orchestration.",
    group: "dark",
    image: sampleImage("runtime_dark"),
    note: "ghost 80 / signal 20",
    codeHref: sampleCode("runtime_dark"),
  },
  {
    id: "agents-dark",
    title: "Agents",
    description: "Agent pool, status states, and action-heavy cards.",
    group: "dark",
    image: sampleImage("agents_dark"),
    note: "card density with one primary action",
    codeHref: sampleCode("agents_dark"),
  },
  {
    id: "memory-dark",
    title: "Memory",
    description: "Timeline layout with recall actions and tags.",
    group: "dark",
    image: sampleImage("memory_dark"),
    note: "editorial timeline with high-density records",
    codeHref: sampleCode("memory_dark"),
  },
  {
    id: "rules-dark",
    title: "Rules",
    description: "Canonical file editor layout with validation sidebar.",
    group: "dark",
    image: sampleImage("rules_dark"),
    note: "editor-first, validation second",
    codeHref: sampleCode("rules_dark"),
  },
  {
    id: "mcp-dark",
    title: "MCP Tools",
    description: "Tool registry and connection table.",
    group: "dark",
    image: sampleImage("mcp_tools_dark"),
    note: "structured list of runtime integrations",
    codeHref: sampleCode("mcp_tools_dark"),
  },
  {
    id: "logs-dark",
    title: "Logs",
    description: "Monospace event stream with command line footer.",
    group: "dark",
    image: sampleImage("logs_dark"),
    note: "live feed, no decorative chrome",
    codeHref: sampleCode("logs_dark"),
  },
  {
    id: "evals-dark",
    title: "Evals",
    description: "Metrics dashboard with score cards and outcomes.",
    group: "dark",
    image: sampleImage("evals_dark"),
    note: "evaluation view with one dominant KPI",
    codeHref: sampleCode("evals_dark"),
  },
  {
    id: "runtime-light",
    title: "Runtime",
    description: "Light-mode mirror of the main runtime surface.",
    group: "light",
    image: sampleImage("runtime_light"),
    note: "same structure, inverted palette",
    codeHref: sampleCode("runtime_light"),
  },
  {
    id: "agents-light",
    title: "Agents",
    description: "Light-mode agent cards and launch states.",
    group: "light",
    image: sampleImage("agents_light"),
    note: "production readable without decoration",
    codeHref: sampleCode("agents_light"),
  },
  {
    id: "runtime-mobile",
    title: "Runtime Mobile",
    description: "Compact runtime layout for narrow screens.",
    group: "mobile",
    image: sampleImage("runtime_mobile_dark"),
    note: "single-column and reduced chrome",
    codeHref: sampleCode("runtime_mobile_dark"),
  },
  {
    id: "memory-mobile",
    title: "Memory Mobile",
    description: "Timeline records stacked for mobile scanning.",
    group: "mobile",
    image: sampleImage("memory_mobile_dark"),
    note: "timeline collapses cleanly",
    codeHref: sampleCode("memory_mobile_dark"),
  },
  {
    id: "logs-mobile",
    title: "Logs Mobile",
    description: "Log stream designed for single-hand reading.",
    group: "mobile",
    image: sampleImage("logs_mobile_dark"),
    note: "terminal-like but vertically safe",
    codeHref: sampleCode("logs_mobile_dark"),
  },
  {
    id: "evals-mobile",
    title: "Evals Mobile",
    description: "Compact metrics and active-session summary.",
    group: "mobile",
    image: sampleImage("evals_mobile_dark"),
    note: "small-screen KPI focus",
    codeHref: sampleCode("evals_mobile_dark"),
  },
];

function VisualSystemPage() {
  return (
    <div className="sample-gallery">
      <SectionHeader
        kicker="brain visual system"
        title="Design samples"
        subtitle="These are repo-owned references for the desktop visual language: dark, light, and mobile, grouped so implementation can stay coherent instead of drifting screen by screen."
      />

      <Card tone="raised" className="sample-gallery__intro">
        <div className="sample-gallery__intro-copy">
          <Badge tone="accent">Source bundle imported</Badge>
          <p>
            The bundle from the provided zip lives under `apps/desktop/src/assets/brain-os-samples/`.
            Each reference card links back to the source HTML so the visuals stay inspectable.
          </p>
        </div>
        <CodeBlock title="Design north star" subtitle="brain_kernel/DESIGN.md">
          {`The UI is a technical OS, not a marketing page.
Use a 4pt grid, a single violet accent, monospace for technical data,
and hard borders instead of glow or shadow.`}
        </CodeBlock>
      </Card>

      {sampleGroups.map((group) => {
        const items = samples.filter((sample) => sample.group === group.id);

        return (
          <section key={group.id} className="sample-group">
            <div className="sample-group__heading">
              <div>
                <h2 className="sample-group__title">{group.label}</h2>
                <p className="sample-group__description">{group.description}</p>
              </div>
              <Badge tone={group.id === "light" ? "neutral" : "accent"}>{items.length} samples</Badge>
            </div>

            <div className="sample-grid">
              {items.map((sample) => (
                <Card key={sample.id} tone="default" className="sample-card">
                  <img className="sample-card__image" src={sample.image} alt={`${sample.title} sample`} />
                  <div>
                    <h3 className="sample-card__title">{sample.title}</h3>
                    <p className="sample-card__meta">{sample.description}</p>
                  </div>
                  <div className="sample-card__link-row">
                    <Badge tone="neutral">{sample.note}</Badge>
                    <Button
                      variant="secondary"
                      onClick={() => window.open(sample.codeHref, "_blank", "noopener,noreferrer")}
                    >
                      Open markup
                    </Button>
                  </div>
                </Card>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}

export default VisualSystemPage;
