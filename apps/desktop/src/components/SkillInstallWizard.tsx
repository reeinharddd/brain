import { useMemo, useState } from "react";

const DAEMON_URL =
  import.meta.env.VITE_BRAIN_DAEMON_URL || "http://localhost:9090";

type SourceType = "auto" | "git" | "local" | "npm";
type ScopeType =
  | "global"
  | "user"
  | "workspace"
  | "project"
  | "organization"
  | "team";

interface SkillInstallVariant {
  id: string;
  name: string;
  description?: string;
  path?: string;
  internal?: boolean;
}

interface SkillInstallPreview {
  source: string;
  source_type: string;
  scope: string;
  requires_selection: boolean;
  available: SkillInstallVariant[];
  selected?: string[];
  notes?: string[];
}

interface SkillInstallWizardProps {
  onInstalled?: () => void;
}

export default function SkillInstallWizard({ onInstalled }: SkillInstallWizardProps) {
  const [source, setSource] = useState("");
  const [sourceType, setSourceType] = useState<SourceType>("auto");
  const [scope, setScope] = useState<ScopeType>("global");
  const [preview, setPreview] = useState<SkillInstallPreview | null>(null);
  const [selectedSkills, setSelectedSkills] = useState<string[]>([]);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);
  const [isInstallLoading, setIsInstallLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [status, setStatus] = useState<string | null>(null);

  const availableSkills = preview?.available || [];
  const canInstall = useMemo(() => {
    if (!source.trim()) return false;
    if (!preview) return false;
    if (preview.requires_selection && selectedSkills.length === 0) return false;
    return true;
  }, [source, preview, selectedSkills]);

  const requestPayload = () => ({
    source,
    source_type: sourceType,
    scope,
    skills: selectedSkills,
    install_all: selectedSkills.length === 0 && !!preview && !preview.requires_selection,
    include_internal: true,
    copy: true,
  });

  const loadPreview = async () => {
    if (!source.trim()) {
      setError("Enter a repository URL or local path first.");
      return;
    }

    setIsPreviewLoading(true);
    setError(null);
    setStatus(null);

    try {
      const res = await fetch(`${DAEMON_URL}/api/skills/install/preview`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          source,
          source_type: sourceType,
          scope,
          skills: selectedSkills,
          install_all: false,
          include_internal: true,
          copy: true,
        }),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Failed to preview skill source");
      }

      const typedPreview = data as SkillInstallPreview;
      setPreview(typedPreview);
      setSelectedSkills((current) => {
        if (typedPreview.available.length === 1) {
          return [typedPreview.available[0].id];
        }
        if (current.length > 0) {
          return current.filter((skill) =>
            typedPreview.available.some((available) => available.id === skill || available.name === skill),
          );
        }
        return [];
      });
      setStatus(
        typedPreview.requires_selection
          ? "Select one or more skills before installing."
          : `Found ${typedPreview.available.length} skill(s) ready to install.`,
      );
    } catch (err) {
      setError(err instanceof Error ? err.message : "Preview failed");
      setPreview(null);
      setSelectedSkills([]);
    } finally {
      setIsPreviewLoading(false);
    }
  };

  const install = async () => {
    if (!canInstall) {
      setError("Preview the source and select at least one skill if multiple variants exist.");
      return;
    }

    setIsInstallLoading(true);
    setError(null);
    setStatus(null);

    try {
      const res = await fetch(`${DAEMON_URL}/api/skills/install`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(requestPayload()),
      });

      const data = await res.json();
      if (!res.ok) {
        throw new Error(data.error || "Failed to install skill source");
      }

      const installed = Array.isArray(data.installed) ? data.installed.length : 0;
      const names = Array.isArray(data.installed)
        ? data.installed
            .map((item: any) => item?.name || item?.id)
            .filter(Boolean)
        : [];
      setStatus(`Installed ${installed} skill(s): ${names.join(", ") || "done"}`);
      onInstalled?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Install failed");
    } finally {
      setIsInstallLoading(false);
    }
  };

  const toggleSkill = (skillId: string) => {
    setSelectedSkills((current) =>
      current.includes(skillId)
        ? current.filter((id) => id !== skillId)
        : [...current, skillId],
    );
  };

  return (
    <section
      style={{
        background: "#ffffff",
        border: "1px solid #e5e7eb",
        borderRadius: 12,
        padding: 16,
        marginBottom: 18,
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12, alignItems: "flex-start" }}>
        <div>
          <h3 style={{ margin: 0, marginBottom: 4 }}>Install Skill Source</h3>
          <p style={{ margin: 0, fontSize: 13, color: "#6b7280" }}>
            Preview a repository, package, or local skill tree before Brain materializes it.
          </p>
        </div>
        <button
          type="button"
          onClick={loadPreview}
          disabled={isPreviewLoading}
          style={{
            padding: "8px 14px",
            borderRadius: 8,
            border: "1px solid #3b82f6",
            color: "#3b82f6",
            background: "white",
            cursor: isPreviewLoading ? "not-allowed" : "pointer",
            opacity: isPreviewLoading ? 0.7 : 1,
          }}
        >
          {isPreviewLoading ? "Previewing..." : "Preview"}
        </button>
      </div>

      <div
        style={{
          display: "grid",
          gridTemplateColumns: "2fr 1fr 1fr",
          gap: 12,
          marginTop: 14,
        }}
      >
        <label style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <span style={{ fontSize: 12, fontWeight: 600, color: "#374151" }}>Source</span>
          <input
            type="text"
            value={source}
            onChange={(e) => setSource(e.target.value)}
            placeholder="https://github.com/Leonxlnx/taste-skill"
            style={{
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: 8,
              boxSizing: "border-box",
            }}
          />
        </label>

        <label style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <span style={{ fontSize: 12, fontWeight: 600, color: "#374151" }}>Source Type</span>
          <select
            value={sourceType}
            onChange={(e) => setSourceType(e.target.value as SourceType)}
            style={{
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: 8,
              boxSizing: "border-box",
            }}
          >
            <option value="auto">Auto detect</option>
            <option value="git">Git / Repo URL</option>
            <option value="npm">npx / Package</option>
            <option value="local">Local path</option>
          </select>
        </label>

        <label style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          <span style={{ fontSize: 12, fontWeight: 600, color: "#374151" }}>Scope</span>
          <select
            value={scope}
            onChange={(e) => setScope(e.target.value as ScopeType)}
            style={{
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: 8,
              boxSizing: "border-box",
            }}
          >
            <option value="global">Global</option>
            <option value="user">User</option>
            <option value="workspace">Workspace</option>
            <option value="project">Project</option>
            <option value="organization">Organization</option>
            <option value="team">Team</option>
          </select>
        </label>
      </div>

      {error && (
        <div
          style={{
            marginTop: 14,
            background: "#fef2f2",
            color: "#991b1b",
            border: "1px solid #fecaca",
            borderRadius: 8,
            padding: 12,
            fontSize: 13,
          }}
        >
          {error}
        </div>
      )}

      {status && !error && (
        <div
          style={{
            marginTop: 14,
            background: "#ecfdf5",
            color: "#065f46",
            border: "1px solid #a7f3d0",
            borderRadius: 8,
            padding: 12,
            fontSize: 13,
          }}
        >
          {status}
        </div>
      )}

      {preview && (
        <div style={{ marginTop: 16 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 10 }}>
            <h4 style={{ margin: 0 }}>Available Variants</h4>
            <button
              type="button"
              onClick={() => setSelectedSkills(availableSkills.map((skill) => skill.id))}
              disabled={availableSkills.length === 0}
              style={{
                padding: "6px 10px",
                borderRadius: 8,
                border: "1px solid #d1d5db",
                background: "white",
                cursor: availableSkills.length === 0 ? "not-allowed" : "pointer",
              }}
            >
              Select All
            </button>
          </div>

          <div style={{ display: "grid", gap: 8 }}>
            {availableSkills.map((skill) => {
              const checked = selectedSkills.includes(skill.id) || selectedSkills.includes(skill.name);
              return (
                <label
                  key={skill.id}
                  style={{
                    display: "flex",
                    alignItems: "flex-start",
                    gap: 10,
                    padding: 12,
                    border: "1px solid #e5e7eb",
                    borderRadius: 10,
                    background: checked ? "#eff6ff" : "#fff",
                    cursor: "pointer",
                  }}
                >
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => toggleSkill(skill.id)}
                    style={{ marginTop: 3 }}
                  />
                  <div style={{ flex: 1 }}>
                    <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
                      <strong>{skill.name}</strong>
                      {skill.internal && (
                        <span style={{ fontSize: 11, background: "#fef3c7", color: "#92400e", padding: "2px 6px", borderRadius: 999 }}>
                          internal
                        </span>
                      )}
                    </div>
                    <div style={{ fontSize: 13, color: "#6b7280", marginTop: 4 }}>
                      {skill.description || "No description provided"}
                    </div>
                    {skill.path && (
                      <div style={{ fontSize: 11, color: "#9ca3af", marginTop: 4, fontFamily: "monospace" }}>
                        {skill.path}
                      </div>
                    )}
                  </div>
                </label>
              );
            })}
          </div>
        </div>
      )}

      <div style={{ display: "flex", justifyContent: "flex-end", marginTop: 16 }}>
        <button
          type="button"
          onClick={install}
          disabled={isInstallLoading || !canInstall}
          style={{
            padding: "9px 16px",
            borderRadius: 8,
            border: "none",
            background: isInstallLoading || !canInstall ? "#d1d5db" : "#10b981",
            color: "white",
            cursor: isInstallLoading || !canInstall ? "not-allowed" : "pointer",
            fontWeight: 600,
          }}
        >
          {isInstallLoading ? "Installing..." : "Install"}
        </button>
      </div>
    </section>
  );
}
