import { useState, useEffect } from "react";
import SkillForm, { CatalogItem } from "./SkillForm";

interface SkillsListProps {
  onSkillsChange?: () => void;
}

export default function SkillsList({ onSkillsChange }: SkillsListProps) {
  const [skills, setSkills] = useState<CatalogItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedSkill, setSelectedSkill] = useState<CatalogItem | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [filterKind, setFilterKind] = useState<
    "all" | "skill" | "context-pack"
  >("all");
  const [searchQuery, setSearchQuery] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch skills list
  const fetchSkills = async () => {
    setIsLoading(true);
    setError("");
    console.log("[SkillsList] Fetching skills from localhost:9090...");

    try {
      const res = await fetch("http://localhost:9090/api/skills");
      console.log("[SkillsList] Response status:", res.status);

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }

      const data = await res.json();
      console.log("[SkillsList] Response data:", data);
      console.log(
        "[SkillsList] data.skills type:",
        typeof data.skills,
        "isArray:",
        Array.isArray(data.skills),
      );

      // Handle different response formats
      if (Array.isArray(data)) {
        // If data is directly an array, use it
        console.log(
          "[SkillsList] Data is array, loading",
          data.length,
          "items",
        );
        setSkills(data);
      } else if (typeof data === "object" && data.skills) {
        // If response has .skills property
        if (Array.isArray(data.skills)) {
          // If .skills is an array, use it directly
          console.log(
            "[SkillsList] data.skills is array, loading",
            data.skills.length,
            "items",
          );
          setSkills(data.skills);
        } else {
          // If .skills is an object (legacy format), convert entries
          console.log(
            "[SkillsList] data.skills is object, converting to array",
          );
          const skillsArray = Object.entries(data.skills).map(
            ([id, skill]: any) => ({
              id,
              ...(typeof skill === "object" ? skill : { name: String(skill) }),
            }),
          );
          setSkills(skillsArray);
        }
      } else {
        console.log("[SkillsList] Unexpected response format");
        setSkills([]);
      }
    } catch (err) {
      const errMsg = err instanceof Error ? err.message : String(err);
      console.error("[SkillsList] Error:", errMsg);
      setError(errMsg);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchSkills();
  }, []);

  // Create or update skill
  const handleSaveSkill = async (skill: CatalogItem) => {
    setIsSubmitting(true);
    try {
      const method = selectedSkill ? "PUT" : "POST";
      const url =
        selectedSkill ?
          `http://localhost:9090/api/skills/${skill.id}`
        : "http://localhost:9090/api/skills";

      const res = await fetch(url, {
        method,
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(skill),
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(
          errorData.error ||
            `Failed to ${selectedSkill ? "update" : "create"} skill`,
        );
      }

      // Refresh list
      await fetchSkills();
      setShowForm(false);
      setSelectedSkill(null);
      onSkillsChange?.();
    } catch (err) {
      throw err;
    } finally {
      setIsSubmitting(false);
    }
  };

  // Delete skill
  const handleDeleteSkill = async (id: string) => {
    if (!confirm(`Delete "${id}"? This action cannot be undone.`)) return;

    try {
      const res = await fetch(`http://localhost:9090/api/skills/${id}`, {
        method: "DELETE",
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.error || "Failed to delete skill");
      }

      // Refresh list
      await fetchSkills();
      onSkillsChange?.();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete skill");
    }
  };

  // Filter and search
  const filteredSkills = skills.filter((skill) => {
    const matchesKind = filterKind === "all" || skill.kind === filterKind;
    const matchesSearch =
      !searchQuery ||
      skill.id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      skill.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      skill.description?.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesKind && matchesSearch;
  });

  return (
    <div
      style={{
        background: "#ffffff",
        padding: "20px",
        borderRadius: "10px",
        boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
      }}
    >
      <h2 style={{ marginTop: 0 }}>Skills & Context Packs Management</h2>

      {error && !showForm && (
        <div
          style={{
            background: "#fee2e2",
            color: "#7f1d1d",
            padding: "16px",
            marginBottom: "16px",
            borderRadius: "8px",
            borderLeft: "4px solid #dc2626",
            fontWeight: 500,
          }}
        >
          <strong>Error loading skills:</strong>
          <div
            style={{
              marginTop: "8px",
              fontSize: "14px",
              whiteSpace: "pre-wrap",
            }}
          >
            {error}
          </div>
          <div style={{ marginTop: "8px", fontSize: "12px", color: "#6b7280" }}>
            Make sure the Brain daemon is running on port 9090:
            <br />
            <code
              style={{
                background: "#f3f4f6",
                padding: "4px 8px",
                borderRadius: "3px",
              }}
            >
              BRAIN_ROOT=$HOME/.brain ./apps/daemon/braind
            </code>
          </div>
        </div>
      )}

      {/* Show Form or List */}
      {showForm ?
        <div>
          <h3 style={{ marginTop: 0, marginBottom: "20px" }}>
            {selectedSkill ? `Edit: ${selectedSkill.name}` : "Create New Skill"}
          </h3>
          <SkillForm
            skill={selectedSkill || undefined}
            onSubmit={handleSaveSkill}
            onCancel={() => {
              setShowForm(false);
              setSelectedSkill(null);
            }}
            isLoading={isSubmitting}
          />
        </div>
      : <>
          {/* Controls */}
          <div
            style={{
              display: "flex",
              gap: "12px",
              alignItems: "center",
              marginBottom: "16px",
              flexWrap: "wrap",
            }}
          >
            <button
              onClick={() => {
                setSelectedSkill(null);
                setShowForm(true);
              }}
              style={{
                padding: "8px 16px",
                background: "#10b981",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: "pointer",
                fontWeight: 500,
              }}
            >
              + New Skill
            </button>

            <input
              type='text'
              placeholder='Search by ID, name, or description...'
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              style={{
                flex: 1,
                minWidth: "200px",
                padding: "8px 12px",
                border: "1px solid #d1d5db",
                borderRadius: "6px",
                boxSizing: "border-box",
              }}
            />

            <select
              value={filterKind}
              onChange={(e) => setFilterKind(e.target.value as any)}
              style={{
                padding: "8px 12px",
                border: "1px solid #d1d5db",
                borderRadius: "6px",
                boxSizing: "border-box",
              }}
            >
              <option value='all'>All Types</option>
              <option value='skill'>Skills Only</option>
              <option value='context-pack'>Context Packs Only</option>
            </select>

            <button
              onClick={fetchSkills}
              disabled={isLoading}
              style={{
                padding: "8px 16px",
                background: "#3b82f6",
                color: "white",
                border: "none",
                borderRadius: "6px",
                cursor: isLoading ? "not-allowed" : "pointer",
                opacity: isLoading ? 0.6 : 1,
              }}
            >
              {isLoading ? "Loading..." : "Refresh"}
            </button>
          </div>

          {/* Skills Table/Grid */}
          {isLoading && !skills.length ?
            <div
              style={{
                textAlign: "center",
                color: "#6b7280",
                padding: "40px 20px",
              }}
            >
              Loading skills...
            </div>
          : filteredSkills.length === 0 ?
            <div
              style={{
                textAlign: "center",
                color: "#6b7280",
                padding: "40px 20px",
              }}
            >
              {skills.length === 0 ?
                "No skills found. Create your first one!"
              : "No skills match your search."}
            </div>
          : <div style={{ overflowX: "auto" }}>
              <table
                style={{
                  width: "100%",
                  borderCollapse: "collapse",
                  fontSize: "14px",
                }}
              >
                <thead>
                  <tr
                    style={{
                      borderBottom: "2px solid #e5e7eb",
                      background: "#f9fafb",
                    }}
                  >
                    <th style={tableHeaderStyle}>ID</th>
                    <th style={tableHeaderStyle}>Name</th>
                    <th style={tableHeaderStyle}>Kind</th>
                    <th style={tableHeaderStyle}>Description</th>
                    <th style={tableHeaderStyle}>Version</th>
                    <th style={tableHeaderStyle}>Tags</th>
                    <th style={tableHeaderStyle}>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredSkills.map((skill, idx) => (
                    <tr
                      key={skill.id}
                      style={{
                        borderBottom: "1px solid #e5e7eb",
                        background: idx % 2 === 0 ? "#ffffff" : "#f9fafb",
                        transition: "background 0.2s",
                      }}
                      onMouseEnter={(e) =>
                        (e.currentTarget.style.background =
                          idx % 2 === 0 ? "#f3f4f6" : "#e5e7eb")
                      }
                      onMouseLeave={(e) =>
                        (e.currentTarget.style.background =
                          idx % 2 === 0 ? "#ffffff" : "#f9fafb")
                      }
                    >
                      <td style={tableCellStyle}>
                        <code
                          style={{
                            background: "#f3f4f6",
                            padding: "2px 6px",
                            borderRadius: "3px",
                            fontSize: "12px",
                          }}
                        >
                          {skill.id}
                        </code>
                      </td>
                      <td style={tableCellStyle}>
                        <strong>{skill.name}</strong>
                      </td>
                      <td style={tableCellStyle}>
                        <span
                          style={{
                            background:
                              skill.kind === "skill" ? "#dbeafe" : "#ddd6fe",
                            color:
                              skill.kind === "skill" ? "#1e40af" : "#4f46e5",
                            padding: "4px 8px",
                            borderRadius: "4px",
                            fontSize: "12px",
                            fontWeight: "500",
                          }}
                        >
                          {skill.kind || "unknown"}
                        </span>
                      </td>
                      <td style={tableCellStyle}>
                        {skill.description ?
                          <span title={skill.description}>
                            {skill.description.substring(0, 50)}
                            {skill.description.length > 50 ? "..." : ""}
                          </span>
                        : <span style={{ color: "#9ca3af" }}>-</span>}
                      </td>
                      <td style={tableCellStyle}>{skill.version || "-"}</td>
                      <td style={tableCellStyle}>
                        {skill.tags && skill.tags.length > 0 ?
                          <div
                            style={{
                              display: "flex",
                              gap: "4px",
                              flexWrap: "wrap",
                            }}
                          >
                            {skill.tags.slice(0, 2).map((tag) => (
                              <span
                                key={tag}
                                style={{
                                  background: "#f0fdf4",
                                  color: "#15803d",
                                  padding: "2px 6px",
                                  borderRadius: "3px",
                                  fontSize: "11px",
                                }}
                              >
                                {tag}
                              </span>
                            ))}
                            {skill.tags.length > 2 && (
                              <span
                                style={{ color: "#9ca3af", fontSize: "11px" }}
                              >
                                +{skill.tags.length - 2}
                              </span>
                            )}
                          </div>
                        : <span style={{ color: "#9ca3af" }}>-</span>}
                      </td>
                      <td style={tableCellStyle}>
                        <div style={{ display: "flex", gap: "8px" }}>
                          <button
                            onClick={() => {
                              setSelectedSkill(skill);
                              setShowForm(true);
                            }}
                            style={{
                              padding: "4px 12px",
                              background: "#3b82f6",
                              color: "white",
                              border: "none",
                              borderRadius: "4px",
                              fontSize: "12px",
                              cursor: "pointer",
                            }}
                          >
                            Edit
                          </button>
                          <button
                            onClick={() => handleDeleteSkill(skill.id)}
                            style={{
                              padding: "4px 12px",
                              background: "#ef4444",
                              color: "white",
                              border: "none",
                              borderRadius: "4px",
                              fontSize: "12px",
                              cursor: "pointer",
                            }}
                          >
                            Delete
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          }

          {/* Summary */}
          <div
            style={{ marginTop: "16px", color: "#6b7280", fontSize: "14px" }}
          >
            Showing {filteredSkills.length} of {skills.length} skills
          </div>
        </>
      }
    </div>
  );
}

const tableHeaderStyle: React.CSSProperties = {
  padding: "12px",
  textAlign: "left",
  fontWeight: 600,
  color: "#374151",
};

const tableCellStyle: React.CSSProperties = {
  padding: "12px",
  verticalAlign: "middle",
};
