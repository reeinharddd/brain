# Brain Refactoring - Execution Plan

**Status**: Phase 1 Complete ✅  
**Date**: March 31, 2026

---

## What We've Done (Phase 1)

✅ **Created modular lib/ structure** (`scripts/lib/`)
- `common.sh` - Shared utilities (6.2K)
- `colors.sh` - ANSI colors (1.8K)
- `logging.sh` - Consistent logging (2.3K)
- `docker.sh` - Docker helpers (3.7K)
- `assert.sh` - Assertions (2.3K)

✅ **Created comprehensive README.NEW.md** (fully documented with:
- Quick start (5 min)
- Architecture overview
- All commands
- Multi-IDE usage
- Troubleshooting
- etc.)

✅ **Created audit documents**
- `REFACTOR_PLAN.md` (goals, risks, strategy)
- This execution plan

---

## What Needs Done (Phase 2-5)

### Phase 2: Consolidate Documentation (30 min)

**Current state**: 85 markdown files across repo

**Action**:
1. Move README.NEW.md → README.md (replace current)
2. Move detailed docs to `docs/guides/` (cloud-sync, memory, etc.)
3. Delete duplicates:
   - ❌ IMPLEMENTATION_COMPLETE.md
   - ❌ MIGRATION_GUIDE.md
   - ❌ STATUS.md
   - ❌ QUICKSTART.md

**Result**: Single README + organized docs/

### Phase 3: Fix Core Scripts Quality (1 hour)

**Scripts to refactor**:
1. `brain-cli.sh` - Add lib imports, improve error handling
2. `init.sh` - Clean up logging, add checks
3. `doctor.sh` - Use consistent patterns from lib/
4. `consolidate-memory.sh` - Same

**Action**: 
- Add `source "${BRAIN_DIR}/scripts/lib/common.sh"` to each
- Replace custom log/color code with lib functions
- Run `shellcheck` and fix warnings

### Phase 4: Consolidate Setup Scripts (1 hour)

**Problem**: 4 setup scripts with overlapping logic
- `install.sh` (343 lines)
- `brain-setup.sh` (75 lines)
- `setup-persistent.sh` (315 lines)
- `autostart-setup.sh` (94 lines)

**Action**:
1. Keep `install.sh` as main (most complete)
2. Fold `brain-setup.sh` logic into `install.sh`
3. Consolidate `setup-persistent.sh` into `install.sh`
4. Remove `autostart-setup.sh` (functionality → `brain-cli.sh`)

**Result**: Single `install.sh` handles all setup

### Phase 5: Organize Utility Scripts (1 hour)

**Problem**: 33 scripts scattered with unclear purposes

**Action**:
1. Move test scripts to `tests/scripts/`:
   - test-docker-mcp.sh
   - test-stdio-mcp.sh
   - test-memory.sh
   - smoke-real-env.sh

2. Move utility scripts to `scripts/lib/` or `scripts/utils/`:
   - detect-stack.sh
   - render-skill-context.sh
   - memory-namespace.sh

3. Keep only high-level in `scripts/`:
   - brain-cli.sh
   - init.sh
   - install.sh
   - deploy.sh
   - update.sh

4. Create `scripts/README.md` documenting each script's purpose

---

## Order of Execution

### RIGHT NOW (Priority 1 - Critical Path)

Do these in order, one per session:

1. **Replace README.md** with README.NEW.md contents
   - Time: 5 min
   - Risk: Low
   - Impact: Immediate clarity for new users

2. **Delete duplicate docs** (IMPLEMENTATION_COMPLETE.md, etc.)
   - Time: 2 min
   - Risk: None (content moved to README)
   - Impact: Reduced confusion

3. **Create docs/guides/** directory structure
   - Time: 10 min
   - Risk: None
   - Impact: Better organization

### NEXT (Priority 2 - Quality)

4. **Update brain-cli.sh** to use `lib/common.sh`
   - Time: 30 min
   - Risk: Medium (test thoroughly)
   - Impact: Code consistency, easier maintenance

5. **Update doctor.sh, init.sh** to use lib
   - Time: 20 min each
   - Risk: Medium
   - Impact: Code consistency

6. **Consolidate setup scripts**
   - Time: 1 hour
   - Risk: High (test all paths)
   - Impact: Simpler onboarding

### LATER (Priority 3 - Organization)

7. Move utility scripts
8. Create test framework
9. Full refactor sweep

---

## How to Execute Phase 2 (Delete Duplicates)

The first concrete step:

```bash
cd ~/.brain

# 1. Backup current README
cp README.md README.md.bak

# 2. Replace with new comprehensive one
cp README.NEW.md README.md

# 3. Delete old documentation files
rm -f IMPLEMENTATION_COMPLETE.md MIGRATION_GUIDE.md STATUS.md QUICKSTART.md

# 4. Verify
ls -lh README*.md IMPLEMENTATION* MIGRATION* STATUS* QUICKSTART* 2>&1 | grep -v "No such"

# 5. Commit
git add -A
git commit -m "brain: consolidate documentation into single comprehensive README"
```

---

## How to Execute Phase 3 (brain-cli.sh Quality)

Example of what we'll do:

```bash
# Before (current brain-cli.sh)
RED='\033[0;31m'
GREEN='\033[0;32m'
log_info() { echo "[INFO] $*"; }

# After (with lib)
source "${BRAIN_DIR}/scripts/lib/common.sh"  # Includes all libs

# Now can use:
log_info "Message"
color_error "Error"
docker_is_running
# etc.
```

**Test after each change**:
```bash
bash ~/.brain/scripts/brain-cli.sh status
brain health
```

---

## Success Metrics

After completing all phases:

- [ ] 1 README (not 5)
- [ ] 8 lib modules (consolidates 15+ script snippets)
- [ ] 4 core scripts (not 33)
- [ ] < 5% code duplication (down from ~15%)
- [ ] All scripts pass: `shellcheck`
- [ ] All commands still work
- [ ] Setup time < 2 min

---

## Risk Mitigation

**Before each major change**:
1. Create git tag: `git tag refactor-phase-X-start`
2. Commit incrementally (not all at once)
3. Test after each change: `brain status`, `brain health`
4. If broken, rollback: `git reset --hard refactor-phase-X-start`

**Safety net**: Everything is version controlled. Easy to revert.

---

## Timeline

- **Phase 2** (Delete duplicates): 10 min TODAY
- **Phase 3** (Quality): 2-3 hours over 3 sessions
- **Phase 4** (Consolidate setup): 1 hour
- **Phase 5** (Organize utils): 1 hour

**Total**: ~5 hours of focused work spread over 1-2 weeks

---

## Next Action

**Tell me which phase to start with**:

1. **Phase 2 NOW** (delete duplicates, replace README)
2. **Phase 3 NEXT** (refactor core scripts)
3. **Wait** (gather more information first)

I recommend: **Phase 2 NOW** (5 min, zero risk, immediate clarity)

Then **Phase 3 NEXT WEEK** (higher risk, requires testing)

What do you want to do?
