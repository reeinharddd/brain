<!-- markdownlint-disable-file -->

---

id: onboarding-new-developers
created: 2026-04-03

---

# Project Doc: Onboarding New Developers

## Quick Start (< 1 hour)

1. **Clone**: `git clone https://github.com/brain/repo.git ~/.brain`
2. **Setup**:
   ```bash
   cd ~/.brain
   go build ./apps/daemon/cmd/braind      # Build daemon
   go build ./apps/cli/cmd/brain          # Build CLI
   ```
3. **Verify**:
   ```bash
   ./bin/brain --version      # Should print version
   ./bin/braind               # Should start daemon
   ```
4. **Read**: `docs/templates/README.md` (30 min, understand architecture)

## Key Concepts

### 1. Daemon-Centric

Everything routes through daemon at `localhost:9090`. CLI and UI are thin clients.

### 2. SDD Phases

Work follows Explore → Propose → Spec → Design → Tasks → Implement → Verify → Archive.

Never skip phases. Even "small" features follow this.

### 3. Templates

Artifacts (agents, skills, rules) must use templates. Never freestyle.

See: `docs/templates/README.md`

### 4. Go-Only

All orchestration in Go. No bash/Python scripts in ~/.brain/.

## Common Tasks

### "I want to add a new agent"

1. Copy: `docs/templates/functional/agents/TEMPLATE.md`
2. Fill in (15 min)
3. Read: `docs/templates/functional/agents/GUIDE-DO-DONT.md` (10 min)
4. Test via CLI: `brain agents list` (should appear)
5. Commit with message: `feat(agents): add new-agent`

### "I want to fix a bug"

1. Run: `brain test --verbose` (see which test fails)
2. Use: debugging-methodology skill (see `docs/templates/functional/skills/EXAMPLES/`)
3. Follow 5 phases: Reproduce → Narrow → Hypothesize → Test → Fix
4. Add test case to prevent regression
5. Verify: all tests pass + lint clear

### "I want to understand architecture"

1. Read: `docs/templates/README.md` (system overview)
2. Read: `docs/adr/` (architecture decisions)
3. Run: `./bin/braind` (see daemon startup logs)
4. Test: `curl localhost:9090/api/health` (verify daemon runs)

## File Structure

```
~/.brain/
├── apps/            Executable surfaces (daemon, cli, desktop)
├── artifacts/       Canonical managed artifacts
├── core/            Shared product subsystems
├── docs/            Documentation + ADRs + SDDs
├── config/          global configurations
└── tests/           Test suites
```

## Rules to Remember

1. **Go-Only**: No bash/Python in orchestration
2. **Daemon-Centric**: Everything routes through port 9090
3. **Template-First**: Use templates for artifacts
4. **Test-First**: Write tests before code
5. **English-Only**: All code, docs, comments in English
6. **SDD Phases**: Follow phase breakdown, never skip

## Links

- **Architecture**: [README.md](../templates/README.md)
- **Templates**: [docs/templates/INDEX.md](../templates/INDEX.md)
- **Debugging**: [Debugging Skill](../templates/functional/skills/EXAMPLES/debugging-methodology.md)
- **API Docs**: [apps/daemon/cmd/braind/README.md](../../../../apps/daemon/cmd/braind/README.md)

## Getting Help

- **Questions?** → Read [README.md](../templates/README.md)
- **Templates?** → See [INDEX.md](../templates/INDEX.md)
- **Debugging?** → Use debugging-methodology skill (docs/templates/functional/skills/EXAMPLES/)
- **Blocked?** → Ask in #brain-dev Slack channel
- **Stuck?** → Use `/standup` command: `brain standup` (shows current status)

## Success Criteria (You're Good When…)

- ✅ You can start daemon: `./bin/braind`
- ✅ You can run CLI: `./bin/brain --version`
- ✅ You understand SDD phases
- ✅ You can use a template (agents, skills, rules)
- ✅ You've read architecture doc
- ✅ You know where to find answers (templates, ADRs)

**Next**: Pick your first task and follow the workflow!

---

**Created**: 2026-04-03  
**Last Updated**: 2026-04-03  
**Audience**: New developers joining Brain team
