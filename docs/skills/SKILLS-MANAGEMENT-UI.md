# Skills Management UI - User Guide

## Overview

The Brain Desktop application now includes a comprehensive **Skills Management** interface where you can:

- **View** all skills and context-packs in a searchable, filterable table
- **Create** new skills or context-packs with a user-friendly form
- **Edit** existing items to update descriptions, versions, and metadata
- **Delete** items (with confirmation to prevent accidents)
- **Search** by ID, name, or description
- **Filter** by type (Skills vs Context-Packs)

## Accessing Skills Management

1. Open **Brain Desktop Control Plane**
2. Scroll down to the **"Skills & Context Packs Management"** section
3. You'll see a table of all loaded skills with action buttons

## Creating a New Skill

### Via UI (Recommended)

1. Click the **"+ New Skill"** button
2. Fill out the form:
   - **ID** (required): Unique identifier, e.g., `my-awesome-skill`
   - **Name** (required): Human-readable name
   - **Kind**: Choose `Skill` or `Context-Pack`
   - **Scope**: Global, Local, or Project
   - **Description**: What does it do?
   - **Version**: Semantic version (e.g., 1.0.0)
   - **Type**: Internal or External
   - **File Path**: Path to the skill definition file
   - **Tags**: Add searchable tags
   - **Sync Targets**: Which tools should have access (CLI, VS Code, Cursor, etc.)
   - **Maintained**: Is it actively maintained?
3. Click **"Create Skill"**
4. The skill is immediately:
   - Written to `registry.yml` (atomically)
   - Synced to all configured targets
   - Available in CLI and other IDEs

### Via API (For Automation)

```bash
curl -X POST http://localhost:9090/api/skills \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-skill",
    "name": "My Awesome Skill",
    "kind": "skill",
    "scope": "global",
    "description": "Does something awesome",
    "version": "1.0.0",
    "type": "internal",
    "tags": ["awesome", "productivity"],
    "syncTo": ["cli", "vscode"],
    "maintained": true
  }'
```

## Editing a Skill

1. Find the skill in the table
2. Click the **"Edit"** button on the right
3. Update any fields (ID cannot be changed)
4. Click **"Update Skill"**
5. Changes are immediately persisted and synced

```bash
# Or via API:
curl -X PUT http://localhost:9090/api/skills/my-skill \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Name", "version": "2.0.0"}'
```

## Deleting a Skill

1. Find the skill in the table
2. Click the **"Delete"** button
3. Confirm the deletion in the popup
4. The skill is removed from:
   - `registry.yml` or `dynamic-registry.tsv`
   - All synced targets
   - CLI output

```bash
# Or via API:
curl -X DELETE http://localhost:9090/api/skills/my-skill
```

## Searching & Filtering

### Search

Type in the search box to filter by:

- Skill ID (exact match)
- Name (contains)
- Description (contains)

Example: "python" will show all Python-related skills

### Filter by Type

Use the dropdown to show:

- **All Types**: Show skills and context-packs
- **Skills Only**: Only skill items
- **Context-Packs Only**: Only context-pack items

## File Storage

### Skills Registry (YAML)

Location: `~/.brain/skills/registry.yml`

Format:

```yaml
skills:
  my-skill:
    name: My Awesome Skill
    version: 1.0.0
    type: internal
    description: Does something awesome
    file: skills/my-skill/SKILL.md
    tags:
      - awesome
      - productivity
    sync-to:
      - cli
      - vscode
    maintained: true
```

Managed by: **SkillsRegistry manager** (creates, updates, deletes YAML entries)

### Context-Packs Registry (TSV)

Location: `~/.brain/skills/dynamic-registry.tsv`

Format (tab-separated):

```text
# skill_id  title  detect_tags  context_path  summary
my-pack  My Context Pack  python,ml  skills/contexts/ml.md  ML-focused context
```

Managed by: **SkillsRegistry manager** (appends, updates, removes rows)

## Synchronization

After any CRUD operation:

1. **File Write** (atomic): Changes are persisted safely (no corruption risk)
2. **Catalog Update** (in-memory): Daemon's internal catalog is updated
3. **Sync Trigger**: Auto-triggered to cascade changes
4. **Target Generation**: `~/.brain/config/skills.json` is regenerated
5. **CLI/IDE Pickup**: All tools read the updated configuration

## Atomic Operations

All writes are atomic (all-or-nothing):

- Temp file created: `registry.yml.tmp`
- Changes written to temp file
- Atomic rename: temp → original
- If any step fails, original file is untouched
- In-memory catalog only updated after successful file write

## Error Handling

Errors are clearly displayed in the UI:

- **Duplicate ID Error** (409 Conflict): A skill with that ID already exists
- **Not Found Error** (404): Trying to update/delete non-existent skill
- **Validation Error** (400): Missing required fields or bad JSON
- **Server Error** (500): Internal daemon error

## Performance Considerations

- **Large Catalogs**: The table shows all skills; very large catalogs (1000+) may be slow
- **Search**: Real-time search across ID, name, and description
- **Sync**: Large syncs (100+ targets) may take a few seconds
- **Network**: For remote daemons, API calls are HTTP (not optimized for high latency)

## Common Workflows

### Add a Python-Specific Context

1. Click "+ New Skill"
2. Set `kind` to "Context-Pack"
3. Use tags: ["python", "ml"]
4. Set description: "Python/ML best practices"
5. Click "Create Skill"

### Update Version Across All

1. Edit the skill
2. Change version field (e.g., 1.0.0 → 2.0.0)
3. Click "Update Skill"
4. CLI immediately shows new version

### Disable a Skill Temporarily

1. Edit the skill
2. Uncheck "Actively Maintained"
3. Click "Update Skill"

(Note: Unchecked items may be filtered in future releases)

## Troubleshooting

| Issue                  | Solution                                                  |
| ---------------------- | --------------------------------------------------------- |
| Skills not appearing   | Click "Refresh" button; check daemon logs                 |
| Sync seems stuck       | Manually click "Sync Config" button in header             |
| Changes not persistent | Verify `~/.brain/skills/` directory has write permissions |
| API returns 404        | Confirm skill ID is typed correctly (case-sensitive)      |
| Form validation fails  | Ensure ID and Name fields are not empty                   |

## API Reference

### List All Skills

```bash
GET http://localhost:9090/api/skills
Response: { "skills": { "id": {...}, ... } }
```

### Get Single Skill

```bash
GET http://localhost:9090/api/skills/my-skill
Response: { "id": "my-skill", "name": "...", ... }
```

### Create Skill

```bash
POST http://localhost:9090/api/skills
Content-Type: application/json
Body: { "id": "...", "name": "...", ...}
Response: 201 Created
```

### Update Skill

```bash
PUT http://localhost:9090/api/skills/my-skill
Content-Type: application/json
Body: { "name": "...", "version": "...", ...}
Response: 200 OK
```

### Delete Skill

```bash
DELETE http://localhost:9090/api/skills/my-skill
Response: 200 OK { "status": "deleted" }
```

### Search Skills

```bash
POST http://localhost:9090/api/skills/search
Content-Type: application/json
Body: { "query": "python" }
Response: { "results": [{...}, {...}] }
```

### Trigger Sync

```bash
POST http://localhost:9090/api/sync
Response: 200 OK
```

## Development Notes

### UI Components

- **SkillForm.tsx**: Reusable form for create/edit operations
- **SkillsList.tsx**: Main dashboard with table, search, and CRUD actions
- **App.tsx**: Integration point and overall layout

### Backend Integration

- All endpoints are in `daemon/cmd/braind/main.go` (`handleSkills()`)
- SkillsRegistry from `daemon/internal/manager/skills.go` handles CRUD
- SyncEngine from `daemon/internal/syncengine/engine.go` cascades changes

### Testing

Run manager tests:

```bash
cd ~/. brain/daemon
go test -v ./internal/manager -run Test
```

All 16 tests should pass:

- 10 existing (load, merge, search, compatibility)
- 6 new CRUD tests (create, update, delete for both skills and context-packs)
