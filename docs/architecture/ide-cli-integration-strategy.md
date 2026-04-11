---
type: design-doc
id: ide-cli-integration-strategy
title: IDE and CLI Integration Strategy
version: 1.0.0
status: active
date_created: 2026-04-11
language: en
category: architecture
relationships:
  - brain-v2-target-architecture
  - ai-runtime-and-context-optimization
  - artifact-system-contract
---

## Overview

This document defines the canonical integration strategy for all IDEs, CLIs, and editor surfaces that Brain supports.

Brain's daemon-centered architecture means every client surface (CLI, IDE extension, desktop app, TUI) is a **thin projection layer** over the daemon's unified state. No client orchestrates independently. All clients read from and write to the same daemon API, ensuring consistent behavior across surfaces.

## Market Research Summary (April 2026)

Research conducted across 15+ AI coding tools reveals three dominant architectural patterns:

1. **Daemon-Centered** (Claude Code KAIROS, OpenClaw, Brain): Always-on background process with persistent memory, autonomous capabilities, cross-session continuity
2. **Stateless CLI** (Codex CLI, Aider, Gemini CLI, OpenCode): Reactive, session-bound, no persistent state between invocations
3. **IDE-Native** (Cursor, Windsurf, VS Code + Copilot, JetBrains + AI): Integrated into editor lifecycle, workspace-aware, file-watcher-driven

**Brain's Position**: Daemon-centered with multi-surface projection. This is the most ambitious architecture but provides the strongest foundation for persistent context, policy enforcement, and cross-tool standardization.

## Target Integration Matrix

### Tier 1: Primary Support (Must Work Flawlessly)

| Surface | Type | Protocol | Architecture | Integration Mode |
|---------|------|----------|--------------|------------------|
| **Brain CLI** | CLI | HTTP + stdio | Daemon client | Official daemon client |
| **Qwen Code CLI** | CLI | ACP + HTTP | CLI frontend + Core backend | Native plugin via daemon HTTP API |
| **VS Code** | IDE Extension | Extension API + WebSocket | Extension Host process | Native extension (TypeScript) |
| **Brain Desktop** | Desktop App | HTTP + WebSocket | Tauri (Rust + React) | Native daemon integration |

### Tier 2: Secondary Support (Plugin/Bridge)

| Surface | Type | Protocol | Architecture | Integration Mode |
|---------|------|----------|--------------|------------------|
| **Claude Code** | CLI | Skills + MCP + Hooks | React TUI (Ink) + KAIROS daemon | Brain as MCP server + skill provider |
| **OpenAI Codex CLI** | CLI | App Server protocol | Rust core + TUI | Brain as custom provider via App Server |
| **Cursor** | IDE | MCP | AI-first IDE with background agents | Brain MCP server |
| **Windsurf (Cascade)** | IDE | MCP | AI-first IDE | Brain MCP server (same as Cursor) |
| **Continue.dev** | IDE Extension | Custom API + MCP | VS Code/JetBrains extension | Bridge plugin |
| **Cline / Roo Code** | IDE Extension | VS Code Extension API | Plan/Act modes, local-first | Custom tool provider |
| **GitHub Copilot** | IDE Extension | VS Code Chat API | Agent mode + workspace context | Context provider integration |
| **Zed** | IDE | ACP | High-performance editor | ACP bridge |
| **JetBrains (IntelliJ)** | IDE Extension | Plugin API (JVM) | Separate JVM process | Native plugin (Kotlin) |
| **Google Antigravity** | IDE | Gemini agents | Agentic AI-first IDE | MCP server bridge |
| **Gemini CLI** | CLI | Skills + MCP | Agent mode with skills | Brain as MCP server |
| **OpenCode** | CLI | Plugin system | Dual-pipeline (MoE) | Brain plugin |
| **Neovim** | Editor | LSP + Custom RPC | LSP-based plugins | LSP plugin |
| **Aider** | CLI | Git-aware | Terminal-based, git commits | Daemon-backed git workflow |
| **Brain TUI** | TUI | HTTP + stdio | Terminal UI (future) | Future |

## Deep Architecture Analysis by Surface

### 1. Claude Code (+ KAIROS Daemon)

**Architecture** (512K+ lines TypeScript, Ink-based React TUI):

- **Core Structure**: `main.tsx` (785KB entry), `bashSecurity.ts` (23 sequential security gates), agent loop with `Submission`/`Event` async channels
- **Subagent Delegation**: Three execution models — `Fork` (isolated), `Teammate` (shared context), `Worktree` (parallel git worktrees)
- **Skills System**: `.claude/skills/` directory with `SKILL.md` (YAML frontmatter + markdown body). Priority: Enterprise → Personal (`~/.claude/skills/`) → Project (`.claude/skills/`) → Plugin (lowest). Progressive disclosure: frontmatter at startup, full content on-demand
- **Advanced Skills**: `/simplify` (3 parallel review agents), `/batch` (5-30 parallel worktree agents), `/debug` (interactive debugging)
- **Hooks**: 17 programmable lifecycle hooks (`PreToolUse`, `PostToolUse`, etc.) with arbitrary validation logic
- **Context**: 200K token window (Opus 4.6). Layered config: `CLAUDE.md` → skills → hooks → rules. AGENTS.md for cross-tool portability
- **Security**: Application-layer enforcement via hooks. Regex-based allow/deny per tool. Pattern-based validation

**KAIROS Daemon** (unreleased, 150+ references in source):

- **Execution Model**: Always-on background process with `<tick>` prompts. Context-aware, event-driven (not cron-based). 15-second blocking budget per decision cycle
- **Memory (autoDream)**: `services/autoDream/` module. Activates when: ≥24h since last cycle, ≥5 completed sessions, no active consolidation lock. Four phases: `Orient` → `Gather Recent Signal` → `Consolidate` → `Prune & Index`. Caps `MEMORY.md` at ~200 lines / 25KB
- **Triggers**: File changes, terminal output, idle/inactivity (triggers autoDream), GitHub webhooks for repo events
- **Exclusive Daemon Tools**: Push notifications, file delivery, PR subscriptions
- **Auditability**: Append-only daily logging. Daemon cannot erase/modify its own logs
- **Persistence**: Session state survives system sleep, laptop closures, restarts

**Brain Integration Strategy**:
1. Expose Brain as MCP server: Claude Code connects via stdio
2. Brain provides skills in `.claude/skills/` location (standard path all tools support)
3. Brain hooks register into Claude Code's lifecycle hook system
4. KAIROS daemon can query Brain's context bundle for cross-session memory
5. Brain policy enforcement overrides Claude Code's local-only rules

**Key Data Structures**:
```yaml
# Brain skill compatible with Claude Code skill format
apiVersion: brain/v1
kind: skill
id: code-refactoring
name: code-refactoring
version: 1.0.0
scope: workspace
# Claude Code compatible frontmatter
claude_compat:
  name: code-refactoring
  description: "3-cycle code refactoring: plan, execute, verify"
  allowed-tools:
    - Read
    - Write
    - Bash
source:
  type: local
  path: artifacts/skills/code-refactoring
content:
  files:
    - path: SKILL.md
      type: markdown
    - path: scripts/refactor.sh
      type: script
```

### 2. OpenAI Codex CLI

**Architecture** (Rust core, stateless inference):

- **Agent Loop**: Concurrent agent threads (up to 6 parallel). `Submission`/`Event` async channels. Three nested loops: submission_loop → handler dispatcher → turn loop (prompt → API → stream). No hardcoded iteration limit
- **Sandbox**: Kernel-level enforcement — Seatbelt (macOS) / Landlock + seccomp (Linux). Three presets: `read-only`, `workspace-write`, `danger-full-access`
- **Tool Calling**: Three categories — built-in (shell, file patching), API-returned, MCP-exposed. All injected into prompt's `tools` field
- **Context**: 1M token window (GPT-5.4 default). Automated compaction via `context_manager`: pairs I/O, truncates payloads, compresses history into encrypted content item
- **Prompt Hierarchy**: system prompt → config → `AGENTS.md` files (32 KiB limit, deeper paths override) → tools → environment → full history
- **App Server Protocol**: Bidirectional, decouples Codex App Server from clients. STDIO or Streaming HTTP transports. Enables unlimited custom clients in any language
- **Security**: Process-safety-first. Zero Data Retention (stateless API calls). Explicit user approval gates

**Brain Integration Strategy**:
1. Register Brain as custom provider in Codex App Server config
2. Brain MCP server connects to Codex's MCP client
3. Brain's context bundle replaces Codex's local context assembly
4. `AGENTS.md` files include Brain policy references

**Key Integration Point**:
```json
// Codex App Server config with Brain provider
{
  "providers": {
    "brain": {
      "type": "mcp",
      "command": "braind",
      "args": ["mcp", "stdio"],
      "context_provider": true,
      "policy_enforcement": true
    }
  },
  "sandbox": "workspace-write",
  "context_compaction": {
    "provider": "brain",
    "endpoint": "http://localhost:8080/api/v1/context/bundle"
  }
}
```

### 3. Qwen Code CLI

**Architecture** (CLI frontend + Core backend separation):

- **CLI Package** (`packages/cli`): User input, slash/@/! commands, conversation history, session resumption, terminal UI, theming, configuration
- **Core Package** (`packages/core`): API communication, prompt construction, tool registration/execution, conversation/session state
- **Tools** (`packages/core/src/tools/`): Modular — filesystem, shell, search, web, MCP. Safety approval workflow: read-only auto-execute, write requires approval
- **Context**: Dynamic prompt assembly — history + context files + tool definitions. Configurable via settings (context file names, directory include/exclude, file filtering)
- **MCP**: Connects to external MCP servers to expand capabilities
- **Configuration**: Layered — CLI args → env vars → project/user/system JSON → defaults

**Brain Integration Strategy**:
1. Register Brain as custom tool provider in Qwen Core
2. Tools: `brain_resolve_artifact`, `brain_get_context`, `brain_check_policy`
3. Context bundle injection into Qwen's prompt construction
4. Model routing: Brain's AI runtime decides model based on capability tier policy

**Implementation**:
```typescript
// packages/core/src/tools/brain.ts
import { Tool } from './tool';
import { BrainClient } from '../clients/brain';

export const brainResolveArtifact: Tool = {
  name: 'brain_resolve_artifact',
  description: 'Resolve artifact via Brain daemon hierarchical policy',
  parameters: {
    kind: { type: 'string', required: true },
    id: { type: 'string', required: true },
    scope: { type: 'string', required: false }
  },
  execute: async (params) => {
    const client = new BrainClient(process.env.BRAIN_DAEMON_URL);
    return client.resolveArtifact(params.kind, params.id, params.scope);
  }
};
```

### 4. Cursor

**Architecture** (AI-first IDE, project-level context, background agents):

- **Context**: Project-wide codebase understanding, not file-by-file. Background agents maintain persistent project model
- **Agent System**: Composable "tasks-to-code" agents. Agent teams for different workflows
- **MCP**: Native MCP client. Connects to external MCP servers for tool extension
- **Skills**: Reads `.claude/skills/` (standard Agent Skills spec). Replaces `.cursorrules` with portable skills

**Brain Integration Strategy**:
1. Brain exposes MCP server that Cursor connects to
2. Cursor agents use Brain MCP tools for artifact resolution, policy checks
3. Brain manages Cursor's context (overrides Cursor's local-only approach)
4. Policy enforcement: Brain ensures Cursor respects hierarchical policies

**MCP Configuration** (in Cursor settings):
```json
{
  "mcpServers": {
    "brain": {
      "command": "braind",
      "args": ["mcp", "stdio"],
      "env": {
        "BRAIN_WORKSPACE": "/path/to/workspace"
      }
    }
  }
}
```

### 5. Windsurf (Cascade)

**Architecture**: Similar to Cursor, AI-first IDE with Cascade agent system

**Brain Integration Strategy**: Identical to Cursor (MCP server bridge)

### 6. VS Code (+ GitHub Copilot)

**Architecture** (Extension Host process, isolated from main IDE):

- **Extension Model**: Runs in separate Extension Host process. `package.json` declares activation events, contribution points
- **AI APIs** (2026): `vscode.lm` (Language Model API), Copilot Chat API, custom chat participants (`@brain`)
- **Context Providers**: `@workspace`, `@terminal`, `@files` — context mention system that injects specific context into chat
- **Agent Mode**: Copilot Coding Agent — autonomous, works on entire project, handles multi-step tasks
- **Chat Participants**: Register custom participants for direct AI interaction within VS Code Chat

**Brain Integration Strategy**:
1. Native VS Code extension (TypeScript) connects to daemon via HTTP + WebSocket
2. Custom chat participant: `@brain` for direct daemon interaction
3. Custom views: Brain Artifacts, Brain Context, Brain Policy
4. File watchers sync artifact changes to daemon
5. Uses `vscode.SecretStorage` for daemon auth tokens
6. Context provider: Brain-compiled context bundle injected into Copilot Chat

**Extension Architecture**:
```typescript
// src/extension.ts
import * as vscode from 'vscode';
import { BrainDaemonClient } from './client';
import { ArtifactViewProvider } from './views/artifacts';
import { ContextViewProvider } from './views/context';
import { BrainChatParticipant } from './chat/participant';

export function activate(context: vscode.ExtensionContext) {
  const daemon = new BrainDaemonClient('http://localhost:8080');
  
  // Register chat participant
  const participant = vscode.chat.createChatParticipant(
    'brain.chat',
    new BrainChatParticipant(daemon)
  );
  
  // Register views
  const artifactView = new ArtifactViewProvider(daemon);
  vscode.window.registerWebviewViewProvider('brain.artifacts', artifactView);
  
  // File watchers
  const watcher = vscode.workspace.createFileSystemWatcher('**/*.md');
  watcher.onDidChange(async (uri) => {
    if (isBrainArtifact(uri)) {
      await daemon.syncArtifact(uri.fsPath);
    }
  });
  
  // Commands
  context.subscriptions.push(
    vscode.commands.registerCommand('brain.applySkill', async () => {
      const bundle = await daemon.getContextBundle();
      // Apply skill from bundle
    })
  );
}
```

### 7. Continue.dev

**Architecture** (VS Code/JetBrains extension, custom model provider system):

- **Model Providers**: Configurable via `config.ts`. Supports OpenAI, Anthropic, local (Ollama), custom
- **Context Providers**: Custom context providers inject additional context into prompts
- **MCP**: Native MCP client for tool extension
- **Skills**: Supports Agent Skills spec (`.claude/skills/`)

**Brain Integration Strategy**:
1. Brain provides Continue "Custom Provider" plugin
2. Plugin translates Continue API calls to daemon HTTP API
3. Brain acts as MCP server for Continue's MCP client
4. Context management: Brain compiles bundle, Continue consumes

**Config**:
```typescript
// continue.config.ts
import { BrainProvider, BrainMCP } from '@brain/continue-plugin';

export default {
  models: [
    new BrainProvider({ daemonUrl: 'http://localhost:8080' })
  ],
  contextProviders: [
    new BrainContextProvider({ daemonUrl: 'http://localhost:8080' })
  ],
  experimental: {
    modelContextProtocolServers: [
      {
        transport: {
          type: 'stdio',
          command: 'braind',
          args: ['mcp', 'stdio']
        }
      }
    ]
  }
};
```

### 8. Cline / Roo Code

**Architecture** (VS Code extension, Plan/Act modes, local-first, zero-trust):

- **Modes**: Plan (information gathering) vs Act (code execution). Explicit mode separation
- **Context**: Local files only, developer-visible before transmission. No external indexing/sync
- **Tool Usage**: Strict human-in-the-loop. Cannot commit/deploy/execute without explicit approval
- **Security**: Zero Trust, fully auditable. Configurable model provider (local Ollama, private cloud, commercial)

**Brain Integration Strategy**:
1. Brain registers as custom tool provider
2. Tools delegate to daemon for artifact resolution, policy checks
3. Context: Brain provides pre-compiled bundle instead of raw file reads
4. Model routing: Brain's AI runtime decides model based on capability tier

### 9. GitHub Copilot (Agent Mode)

**Architecture** (VS Code extension, agent mode, workspace context):

- **Agent Mode**: Autonomous coding agent that analyzes entire project, handles multi-step tasks
- **Context Mentions**: `@workspace` (entire project), `@terminal` (terminal output), `@files` (specific files)
- **Copilot Coding Agent**: GA since September 2025. Works independently on project-level tasks
- **Workspace Context**: Pulls entire project structure into chat context

**Brain Integration Strategy**:
1. Brain provides context provider that replaces `@workspace` with daemon-compiled bundle
2. Brain policy enforcement ensures Copilot respects org/workspace policies
3. Brain skills integrate as Copilot custom instructions

### 10. Google Antigravity

**Architecture** (Agentic AI-first IDE, Gemini 3 powered, launched November 2025):

- **Agent-First Design**: AI as autonomous agent, not just completion assistant
- **Models**: Gemini 3 Pro / Flash / Deep. Multi-model routing
- **Agent Management**: Multiple agents for different workflows (build, refactor, debug)
- **Project Understanding**: Deep codebase comprehension, not surface-level

**Brain Integration Strategy**:
1. Brain exposes MCP server that Antigravity connects to
2. Brain skills register as Antigravity agent capabilities
3. Policy enforcement: Brain ensures Antigravity respects hierarchical policies

### 11. Gemini CLI

**Architecture** (Google's terminal CLI, agent mode with skills):

- **Agent Skills**: Full Agent Skills spec compliance (`SKILL.md` + `scripts/` + `references/` + `assets/`)
- **MCP**: Native MCP server and client
- **Agent Mode**: Pair programmer with project-level context

**Brain Integration Strategy**:
1. Brain provides skills in `.claude/skills/` (Gemini CLI reads this location)
2. Brain MCP server for Gemini CLI's MCP client
3. Skills registered via Brain daemon, discovered by Gemini CLI at startup

### 12. Zed

**Architecture** (High-performance editor, ACP - Agent Client Protocol):

- **ACP** (launched January 2026 by Zed + JetBrains): Standardized protocol connecting AI agents to editor clients. LSP moment for AI agents
- **ACP Registry**: Curated directory of ACP-compatible agents
- **Agent Server**: External process implementing ACP, communicates via stdio/HTTP

**Brain Integration Strategy**:
1. Brain implements ACP server (stdio transport for local, HTTP for remote)
2. ACP bridge translates ACP calls to daemon HTTP API
3. Context: Brain compiles bundle, provides to Zed via ACP

**ACP Server Configuration**:
```yaml
# .zed/settings.json
{
  "agent_servers": {
    "brain": {
      "name": "Brain Daemon",
      "command": "braind",
      "args": ["acp", "stdio"],
      "capabilities": ["tools", "context", "policy", "memory"]
    }
  }
}
```

### 13. JetBrains (IntelliJ)

**Architecture** (Plugin system, separate JVM process):

- **Plugin API**: JVM-based (Kotlin/Java). Tool windows, file listeners, password storage
- **ACP Registry**: JetBrains co-launched ACP with Zed
- **Agent Mode**: Available in Android Studio (Gemini), IntelliJ (AI Assistant)

**Brain Integration Strategy**:
1. Brain provides IntelliJ plugin (Kotlin)
2. Plugin connects to daemon via HTTP + WebSocket
3. Tool windows for artifacts, context, policy
4. File watchers sync artifact changes

### 14. OpenCode

**Architecture** (Dual-pipeline Mixture of Experts):

- **Manager Layers**: `manager-opencode` (COO, 80% routine, cheap models) + `manager-gemini` (CTO, complex problems)
- **Specialized Agents**: `@architect-core`, `@builder-fast`, `@qa-hawk`, `@biz-strategist`, `@creative-lead`
- **Plugin System**: Declarative `"plugin"` array in `config.json`. Startup-loaded modules
- **Tool Loading**: Heavy tools disabled globally by default. Enabled per-agent to reduce context bloat
- **Config**: `~/.config/opencode/config.json` with `$schema`

**Brain Integration Strategy**:
1. Brain provides OpenCode plugin (registered in config.json plugin array)
2. Plugin adds Brain tools: resolve_artifact, get_context, check_policy
3. Brain context bundle available to all OpenCode agents

**Config**:
```json
{
  "plugins": ["@brain/opencode-plugin@latest"],
  "agents": {
    "brain-aware": {
      "model": "brain-routed",
      "tools": ["brain_resolve", "brain_context"],
      "context_provider": "brain"
    }
  }
}
```

### 15. Aider

**Architecture** (Terminal-based, git-aware AI coding):

- **Git Integration**: Every AI edit is a commit with descriptive message. Full rollback and review capability
- **Model Support**: All major providers, local (Ollama), custom endpoints
- **Context**: Reads repository files, git history for understanding

**Brain Integration Strategy**:
1. Brain provides git hook that queries daemon for context before each aider session
2. Brain skills registered as aider custom instructions
3. Aider commits include Brain artifact metadata

### 16. Neovim

**Architecture** (LSP-based plugin system):

- **LSP**: Language Server Protocol for code intelligence
- **Plugins**: Telescope/fzf for fuzzy finding, LSP for code actions

**Brain Integration Strategy**:
1. Brain provides LSP server (`brain-lsp`)
2. LSP translates to daemon HTTP calls
3. Telescope integration for artifact browsing

**Config**:
```lua
-- init.lua
require('lspconfig').brain.setup({
  cmd = { 'brain-lsp', '--daemon-url', 'http://localhost:8080' },
  on_attach = function(client, bufnr)
    vim.keymap.set('n', '<leader>ba', ':BrainApplySkill<CR>', { buffer = bufnr })
    vim.keymap.set('n', '<leader>bc', ':BrainContext<CR>', { buffer = bufnr })
  end
})
```

## Canonical Data Structures

### Standardized Skill Format (Cross-Tool Compatible)

This format is compatible with all tools that support the Agent Skills spec (Claude Code, Codex, Gemini CLI, Cursor, Continue, OpenCode):

```yaml
---
# Brain artifact envelope (apiVersion: brain/v1)
apiVersion: brain/v1
kind: skill
id: code-refactoring
name: code-refactoring
version: 1.2.0
scope: workspace
owner: reeinharrrd
visibility: internal

# Agent Skills spec compatibility (all tools read this)
description: "3-cycle code refactoring: plan changes, execute safely, verify correctness"
compatibility:
  claude_code: true
  codex_cli: true
  gemini_cli: true
  cursor: true
  continue: true
  opencode: true
  aider: true

# Brain-specific fields
source:
  type: local
  path: artifacts/skills/code-refactoring
  checksum: sha256:abc123...

sync:
  enabled: true
  mode: cloud-synced
  last_synced: "2026-04-11T10:00:00Z"

security:
  trust_level: verified
  permissions:
    - read:workspace
    - execute:local_script
  requires_approval: false

content:
  status: active
  compatibility:
    min_capability_tier: 2
    context_cost_estimate: 400
    tool_dependencies:
      - shell
      - read
      - write
    output_strictness: high
  
  activation: manual
  prerequisites:
    - language:go
    - language:typescript
  
  files:
    - path: SKILL.md
      type: markdown
    - path: scripts/refactor.sh
      type: script
    - path: examples/before.go
      type: example
    - path: tests/skill_test.go
      type: test
---

# SKILL.md body follows (markdown with instruction content)
```

### Context Bundle (Compiled Output, All Tools Consume This)

```json
{
  "bundle_id": "ctx_20260411_abc123",
  "compiled_at": "2026-04-11T10:00:00Z",
  "scope_chain": ["org:myorg", "user:reeinharrrd", "workspace:brain", "project:core"],
  "layers": {
    "hard_policy": {
      "source": "org:myorg",
      "content": "No hardcoded secrets, Go-only orchestration",
      "enforcement": "hard"
    },
    "org_baseline": { "token_count": 200 },
    "user_baseline": { "token_count": 150 },
    "workspace_context": { "token_count": 300 },
    "project_context": { "token_count": 250 },
    "task_local": { "token_count": 100 },
    "active_skills": [
      {
        "id": "code-refactoring",
        "scope": "workspace:brain",
        "instructions": "3-cycle refactoring process...",
        "token_count": 400
      }
    ],
    "memory_semantic": [
      {
        "source": "qdrant",
        "matches": [{ "id": "mem_abc123", "similarity": 0.92 }]
      }
    ]
  },
  "totals": {
    "token_count": 1400,
    "token_limit": 8000,
    "utilization_percent": 17.5,
    "compression_applied": true,
    "original_count": 3200
  }
}
```

### Policy Resolution Result (All Tools Enforce This)

```json
{
  "policy_id": "pol_20260411_xyz789",
  "resolved_at": "2026-04-11T10:00:00Z",
  "scope_chain": ["org:myorg", "user:reeinharrrd", "workspace:brain"],
  "rules": {
    "no-hardcoded-secrets": {
      "value": "enforced",
      "source": "org:myorg",
      "class": "hard",
      "override_allowed": false
    },
    "model-routing": {
      "value": "prefer-local-first",
      "source": "user:reeinharrrd",
      "class": "guarded",
      "override_allowed": true,
      "override_requires_approval": true
    }
  },
  "trust_boundaries": {
    "third_party_artifacts": "deny-by-default",
    "privileged_capabilities": "require-approval",
    "external_mcp_servers": "allowlist-only"
  }
}
```

## Integration Protocol Summary

| Protocol | Used By | Brain Role | Transport |
|----------|---------|------------|-----------|
| **HTTP REST API** | All surfaces | Daemon primary API | HTTP/JSON |
| **WebSocket** | VS Code, Desktop, TUI | Real-time events | WS/JSON |
| **MCP** | Cursor, Windsurf, Continue, Gemini CLI, Cline | Brain as MCP server | stdio / HTTPS |
| **ACP** | Zed, JetBrains, OpenClaw | Brain as ACP server | stdio / HTTP |
| **App Server** | Codex CLI | Brain as custom provider | STDIO / Streaming HTTP |
| **VS Code Extension API** | VS Code, Cline, Roo Code, Continue | Native extension | Node.js IPC |
| **LSP** | Neovim, editors | Brain LSP server | stdio |
| **JetBrains Plugin API** | IntelliJ, Android Studio | Native plugin | JVM |

## Standardization Rules

### 1. Daemon is Single Source of Truth
No client implements independent artifact resolution, context assembly, or policy evaluation. All clients call daemon endpoints.
**Rule**: `client ≠ orchestrator`. Client is projection layer only.

### 2. Context Bundle is Daemon-Compiled
Clients never assemble context independently. Daemon applies hierarchical resolution, compression, deduplication.
**Rule**: Context compilation is daemon responsibility.

### 3. Policy is Daemon-Enforced
Clients read resolved policy. They do not evaluate independently. Violations rejected by daemon.
**Rule**: Policy enforcement is daemon-only.

### 4. Skills Follow Agent Skills Spec
All Brain skills use `.claude/skills/` location (standard path). `SKILL.md` with frontmatter + markdown body. Progressive disclosure.
**Rule**: Brain skills are cross-tool compatible.

### 5. Model Routing are Daemon-Decided
Clients submit job requests with requirements. Daemon's AI runtime selects model based on policy and cost.
**Rule**: Model selection is daemon-decided.

### 6. KAIROS-Inspired Idle Processing
Brain daemon implements autoDream-like consolidation during user inactivity. Memory hygiene with strict caps.
**Rule**: Autonomous context hygiene, transparent and auditable.

### 7. Execution Budgets
All daemon autonomous operations have hard time/resource caps (15s decision budget from KAIROS pattern).
**Rule**: Daemon never monopolizes system resources.

## Implementation Phases

### Phase 1: Core Daemon API (Current)
- [ ] Implement daemon HTTP API (health, artifacts, context, policy, runtime)
- [ ] Implement WebSocket event stream
- [ ] Brain CLI as primary client
- [ ] Artifact CRUD + hierarchical resolution
- [ ] Context bundle compiler (org→user→workspace→project→task)

### Phase 2: MCP Server
- [ ] Implement Brain MCP server (Go, stdio + HTTP transport)
- [ ] Expose tools: `brain_resolve_artifact`, `brain_get_context`, `brain_get_policy`, `brain_list_skills`, `brain_list_agents`, `brain_apply_skill`
- [ ] Expose resources: `brain://artifacts/{kind}/{id}`, `brain://context/bundle`, `brain://policy/resolved`
- [ ] Test with Cursor, Windsurf, Continue, Gemini CLI, Cline

### Phase 3: VS Code Extension
- [ ] Scaffold VS Code extension (TypeScript)
- [ ] Connect to daemon HTTP API + WebSocket
- [ ] Register `@brain` chat participant
- [ ] Custom views: Brain Artifacts, Brain Context, Brain Policy
- [ ] File watchers for artifact sync
- [ ] Commands: `brain.applySkill`, `brain.getContext`, `brain.resolvePolicy`

### Phase 4: ACP Server
- [ ] Implement ACP server (stdio + HTTP)
- [ ] Register in ACP Registry (Zed + JetBrains)
- [ ] Bridge ACP → daemon HTTP API
- [ ] Test with Zed IDE
- [ ] Test with JetBrains plugin

### Phase 5: Qwen Code Integration
- [ ] Register Brain as custom tool provider in Qwen Core
- [ ] Implement Brain tools: `brain_resolve`, `brain_context`, `brain_policy`
- [ ] Context bundle injection into prompt construction
- [ ] Model routing integration

### Phase 6: Codex CLI Integration
- [ ] Register Brain as custom provider in Codex App Server
- [ ] Brain MCP → Codex MCP client
- [ ] Context bundle replaces local assembly
- [ ] `AGENTS.md` with Brain policy references

### Phase 7: Additional IDEs/CLIs
- [ ] Continue.dev bridge plugin
- [ ] OpenCode plugin
- [ ] Aider git hook integration
- [ ] Neovim LSP plugin
- [ ] Brain Desktop (Tauri + React)
- [ ] Brain TUI

## Validation

For each integration, validate:

1. **Artifact Resolution**: Client resolves any artifact via daemon
2. **Context Bundle**: Client receives compiled context bundle
3. **Policy Compliance**: Client respects resolved policy
4. **Skill Loading**: Client discovers and loads Brain skills (progressive disclosure)
5. **Event Streaming**: Client receives real-time daemon events
6. **Auth**: Client authenticates with daemon securely
7. **Offline Mode**: Client handles daemon unavailability gracefully
8. **Cross-Tool Compatibility**: Same skill works in Claude Code, Codex, Cursor, VS Code, etc.
9. **Memory Continuity**: Cross-session memory persists via daemon (KAIROS-inspired autoDream)
10. **Execution Budgets**: Daemon autonomous ops respect time/resource caps
