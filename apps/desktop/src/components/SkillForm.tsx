import { useState } from "react";

export interface CatalogItem {
  id: string;
  name: string;
  kind: "skill" | "context-pack";
  scope?: string;
  description?: string;
  version?: string;
  type?: string;
  file?: string;
  path?: string;
  tags?: string[];
  maintained?: boolean;
  source?: string;
  category?: string;
  syncTo?: string[];
  requires?: string[];
}

interface SkillFormProps {
  skill?: CatalogItem;
  onSubmit: (skill: CatalogItem) => Promise<void>;
  onCancel: () => void;
  isLoading?: boolean;
}

export default function SkillForm({
  skill,
  onSubmit,
  onCancel,
  isLoading = false,
}: SkillFormProps) {
  const isEdit = !!skill;

  const [formData, setFormData] = useState<CatalogItem>({
    id: skill?.id || "",
    name: skill?.name || "",
    kind: skill?.kind || "skill",
    scope: skill?.scope || "global",
    description: skill?.description || "",
    version: skill?.version || "1.0.0",
    type: skill?.type || "internal",
    file: skill?.file || "",
    path: skill?.path || "",
    tags: skill?.tags || [],
    maintained: skill?.maintained ?? true,
    category: skill?.category || "",
    syncTo: skill?.syncTo || ["cli"],
  });

  const [tagInput, setTagInput] = useState("");
  const [syncInput, setSyncInput] = useState("");
  const [error, setError] = useState("");

  const handleInputChange = (
    e: React.ChangeEvent<
      HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    >,
  ) => {
    const { name, value, type } = e.currentTarget as HTMLInputElement;

    if (type === "checkbox") {
      setFormData((prev) => ({
        ...prev,
        [name]: (e.currentTarget as HTMLInputElement).checked,
      }));
    } else {
      setFormData((prev) => ({
        ...prev,
        [name]: value,
      }));
    }
  };

  const addTag = () => {
    if (tagInput.trim()) {
      setFormData((prev) => ({
        ...prev,
        tags: [...(prev.tags || []), tagInput.trim()],
      }));
      setTagInput("");
    }
  };

  const removeTag = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      tags: prev.tags?.filter((_, i) => i !== index) || [],
    }));
  };

  const addSyncTarget = () => {
    if (syncInput.trim() && !formData.syncTo?.includes(syncInput.trim())) {
      setFormData((prev) => ({
        ...prev,
        syncTo: [...(prev.syncTo || []), syncInput.trim()],
      }));
      setSyncInput("");
    }
  };

  const removeSyncTarget = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      syncTo: prev.syncTo?.filter((_, i) => i !== index) || [],
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!formData.id.trim()) {
      setError("ID is required");
      return;
    }

    if (!formData.name.trim()) {
      setError("Name is required");
      return;
    }

    try {
      await onSubmit(formData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save skill");
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "16px",
        maxWidth: "600px",
      }}
    >
      {error && (
        <div
          style={{
            background: "#fee2e2",
            color: "#991b1b",
            padding: "12px",
            borderRadius: "6px",
            borderLeft: "4px solid #dc2626",
          }}
        >
          {error}
        </div>
      )}

      <div
        style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}
      >
        {/* ID Field - disabled on edit */}
        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            ID <span style={{ color: "#ef4444" }}>*</span>
          </label>
          <input
            type='text'
            name='id'
            value={formData.id}
            onChange={handleInputChange}
            disabled={isEdit}
            placeholder='skill-identifier'
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              fontFamily: "monospace",
              opacity: isEdit ? 0.6 : 1,
              cursor: isEdit ? "not-allowed" : "text",
              boxSizing: "border-box",
            }}
          />
          <small style={{ color: "#6b7280" }}>
            Cannot change after creation
          </small>
        </div>

        {/* Name Field */}
        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            Name <span style={{ color: "#ef4444" }}>*</span>
          </label>
          <input
            type='text'
            name='name'
            value={formData.name}
            onChange={handleInputChange}
            placeholder='Human-readable name'
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          />
        </div>
      </div>

      {/* Kind and Scope */}
      <div
        style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}
      >
        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            Kind
          </label>
          <select
            name='kind'
            value={formData.kind}
            onChange={handleInputChange}
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          >
            <option value='skill'>Skill</option>
            <option value='context-pack'>Context Pack</option>
          </select>
        </div>

        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            Scope
          </label>
          <select
            name='scope'
            value={formData.scope}
            onChange={handleInputChange}
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          >
            <option value='global'>Global</option>
            <option value='local'>Local</option>
            <option value='project'>Project</option>
          </select>
        </div>
      </div>

      {/* Description */}
      <div>
        <label
          style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
        >
          Description
        </label>
        <textarea
          name='description'
          value={formData.description}
          onChange={handleInputChange}
          placeholder='What does this skill do?'
          rows={3}
          style={{
            width: "100%",
            padding: "8px 12px",
            border: "1px solid #d1d5db",
            borderRadius: "6px",
            fontFamily: "inherit",
            boxSizing: "border-box",
          }}
        />
      </div>

      {/* Version and Type */}
      <div
        style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}
      >
        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            Version
          </label>
          <input
            type='text'
            name='version'
            value={formData.version}
            onChange={handleInputChange}
            placeholder='1.0.0'
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          />
        </div>

        <div>
          <label
            style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
          >
            Type
          </label>
          <select
            name='type'
            value={formData.type}
            onChange={handleInputChange}
            style={{
              width: "100%",
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          >
            <option value='internal'>Internal</option>
            <option value='external'>External</option>
          </select>
        </div>
      </div>

      {/* File Path */}
      <div>
        <label
          style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
        >
          File Path
        </label>
        <input
          type='text'
          name='file'
          value={formData.file}
          onChange={handleInputChange}
          placeholder='skills/skill-name/SKILL.md'
          style={{
            width: "100%",
            padding: "8px 12px",
            border: "1px solid #d1d5db",
            borderRadius: "6px",
            fontFamily: "monospace",
            boxSizing: "border-box",
          }}
        />
      </div>

      {/* Tags */}
      <div>
        <label
          style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
        >
          Tags
        </label>
        <div style={{ display: "flex", gap: "8px", marginBottom: "8px" }}>
          <input
            type='text'
            value={tagInput}
            onChange={(e) => setTagInput(e.target.value)}
            onKeyPress={(e) =>
              e.key === "Enter" && (e.preventDefault(), addTag())
            }
            placeholder='Add a tag...'
            style={{
              flex: 1,
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          />
          <button
            type='button'
            onClick={addTag}
            style={{
              padding: "8px 16px",
              background: "#3b82f6",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: "pointer",
            }}
          >
            Add
          </button>
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
          {formData.tags?.map((tag, i) => (
            <div
              key={i}
              style={{
                background: "#dbeafe",
                color: "#1e40af",
                padding: "4px 8px",
                borderRadius: "4px",
                display: "flex",
                alignItems: "center",
                gap: "6px",
                fontSize: "14px",
              }}
            >
              {tag}
              <button
                type='button'
                onClick={() => removeTag(i)}
                style={{
                  background: "none",
                  border: "none",
                  color: "#1e40af",
                  cursor: "pointer",
                  fontSize: "16px",
                  padding: 0,
                }}
              >
                x
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Sync Targets */}
      <div>
        <label
          style={{ display: "block", marginBottom: "6px", fontWeight: 500 }}
        >
          Sync Targets
        </label>
        <div style={{ display: "flex", gap: "8px", marginBottom: "8px" }}>
          <input
            type='text'
            value={syncInput}
            onChange={(e) => setSyncInput(e.target.value)}
            onKeyPress={(e) =>
              e.key === "Enter" && (e.preventDefault(), addSyncTarget())
            }
            placeholder='cli, vscode, cursor...'
            style={{
              flex: 1,
              padding: "8px 12px",
              border: "1px solid #d1d5db",
              borderRadius: "6px",
              boxSizing: "border-box",
            }}
          />
          <button
            type='button'
            onClick={addSyncTarget}
            style={{
              padding: "8px 16px",
              background: "#10b981",
              color: "white",
              border: "none",
              borderRadius: "6px",
              cursor: "pointer",
            }}
          >
            Add
          </button>
        </div>
        <div style={{ display: "flex", flexWrap: "wrap", gap: "8px" }}>
          {formData.syncTo?.map((target, i) => (
            <div
              key={i}
              style={{
                background: "#d1fae5",
                color: "#065f46",
                padding: "4px 8px",
                borderRadius: "4px",
                display: "flex",
                alignItems: "center",
                gap: "6px",
                fontSize: "14px",
              }}
            >
              {target}
              <button
                type='button'
                onClick={() => removeSyncTarget(i)}
                style={{
                  background: "none",
                  border: "none",
                  color: "#065f46",
                  cursor: "pointer",
                  fontSize: "16px",
                  padding: 0,
                }}
              >
                x
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Maintained Checkbox */}
      <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
        <input
          type='checkbox'
          name='maintained'
          checked={formData.maintained}
          onChange={handleInputChange}
          id='maintained'
          style={{ cursor: "pointer" }}
        />
        <label htmlFor='maintained' style={{ cursor: "pointer" }}>
          Actively Maintained
        </label>
      </div>

      {/* Form Actions */}
      <div
        style={{
          display: "flex",
          gap: "12px",
          justifyContent: "flex-end",
          marginTop: "20px",
          borderTop: "1px solid #e5e7eb",
          paddingTop: "16px",
        }}
      >
        <button
          type='button'
          onClick={onCancel}
          disabled={isLoading}
          style={{
            padding: "8px 24px",
            background: "#e5e7eb",
            color: "#374151",
            border: "none",
            borderRadius: "6px",
            cursor: isLoading ? "not-allowed" : "pointer",
            opacity: isLoading ? 0.6 : 1,
          }}
        >
          Cancel
        </button>
        <button
          type='submit'
          disabled={isLoading}
          style={{
            padding: "8px 24px",
            background: isLoading ? "#d1d5db" : "#3b82f6",
            color: "white",
            border: "none",
            borderRadius: "6px",
            cursor: isLoading ? "not-allowed" : "pointer",
          }}
        >
          {isLoading ?
            "Saving..."
          : isEdit ?
            "Update Skill"
          : "Create Skill"}
        </button>
      </div>
    </form>
  );
}
