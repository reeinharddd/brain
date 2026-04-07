# PROMPT FOR NEXT AGENT: Documentation Validation System Implementation

## TASK

Implement a **3-layer documentation validation, structure management, and change tracking system** for the Brain repository.

This is a **complete handover** for a new agent session. You have all context needed.

---

## START HERE

Read this document completely, then:

1. Read: `docs/HANDOVER-DOCS-VALIDATION-SYSTEM.md` (full context)
2. Understand: 3-layer architecture (Manifest + Incremental Validator + Changelog)
3. Implement: 5 phases in order
4. Test: All acceptance criteria
5. Verify: Everything works as designed

---

## WHAT YOU'RE BUILDING

### Layer 1: Manifest (`docs-manifest.json`)
- Immutable baseline of expected documentation structure
- Defines all domains, required files, naming rules, validation rules
- Changed only when structure decisions change (rare, deliberate)

### Layer 2: Incremental Validator (`scripts/validate-docs-incremental.sh`)
- Validates ONLY changed files (git diff)
- Checks structure (against manifest) and content (rules)
- Fast (~500ms), not full validation every commit
- Integrates into pre-commit hook

### Layer 3: Changelog (`docs-changelog.jsonl`)
- Records every documentation change
- Format: newline-delimited JSON
- Used by RAG later to inject relevant context
- Immutable append-only log

## ALIGNMENT RULES

To avoid conflicting interpretations during implementation, treat these as hard boundaries:

1. **Scope is current docs only**
   - Work against the current `docs/` structure as the authoritative baseline.
   - Do not perform a repo-wide documentation cleanup in this phase.

2. **Validate by change scope**
   - The validator must focus on staged/changed docs only.
   - Existing legacy content that is not touched should not block this task.

3. **Templates are reference material**
   - `docs/templates/` is part of the repository structure.
   - Treat template files as structure/reference files first; content-policy enforcement for them can be relaxed or skipped if needed for the baseline.

4. **Manifest is the contract**
   - If prose in older docs disagrees with the manifest, the manifest wins for validation purposes.
   - Do not rewrite historical prose unless it is explicitly in scope for a changed file.

5. **Phase 1 is not cleanup**
   - No domain reorg, no folder renames, no broad markdown rewrite.
   - Only create the manifest, incremental validator, changelog, and hook integration.

---

## PHASES (In Order)

### Phase 1: Manifest Generation (1 hour)
**Goal**: Create `docs-manifest.json` documenting current structure

**Do**:
1. Analyze `docs/` directory structure
   - Domains: adr/, architecture/, skills/, testing/, templates/
   - Root files: README.md, INDEX.md, QUICK-START.md, STRUCTURE.md, etc.
2. For each domain:
   - Document purpose
   - List required files
   - Reference template file
   - Define validation rules (naming, frontmatter, sections)
3. Define global rules:
   - Language: English only
   - No emojis
   - Filename format: lowercase-with-hyphens.md
   - Frontmatter fields: type, id, title, version, status, date_created, language, category
4. Set file count targets (current ~78, growth ~100)
5. Commit manifest as immutable baseline

**Validate**:
- ✅ Manifest is valid JSON
- ✅ All 5 domains covered
- ✅ Template references exist
- ✅ File count targets realistic
- ✅ Covers both root files and domains

**Output**: `docs-manifest.json` (committed)

---

### Phase 2: Incremental Validator Script (1.5 hours)
**Goal**: Create `scripts/validate-docs-incremental.sh`

**Algorithm**:
1. Get changed files: `git diff --name-only HEAD docs/`
2. For each changed file:
   a. **Structure check**:
      - Is domain in manifest?
      - Is file in correct path?
      - Does it match template?
      - Return FAIL if structure violation
   b. **Content check** (only if substantive changes):
      - Skip if: only metadata/formatting changed
      - Check if: content text, sections, images changed
      - Validate: frontmatter, language, markdown syntax
      - Return FAIL if content violation
3. Generate changelog entry:
   - Timestamp, commit hash, action, domain, file, checksum
   - Append to `docs-changelog.jsonl`
4. Report results and exit

**Acceptance**:
- ✅ Detects git-staged changes
- ✅ Validates structure against manifest
- ✅ Validates content rules
- ✅ Appends to changelog
- ✅ Works as git hook
- ✅ Fast (<1 second)

**Output**: `scripts/validate-docs-incremental.sh` (executable)

---

### Phase 3: Git Hook Integration (30 min)
**Goal**: Update `.git/hooks/pre-commit` to use incremental validator

**Do**:
1. Current hook has:
   - Part 1: Skills validation
   - Part 2: Old documentation validation (full)
2. Update Part 2 to:
   - Call `validate-docs-incremental.sh` instead
   - Pass staged files
   - Block commit if fails
   - Show clear errors

**Acceptance**:
- ✅ Hook calls incremental validator
- ✅ Blocks bad commits
- ✅ Allows formatting-only changes
- ✅ Clear error messages

**Output**: Updated `.git/hooks/pre-commit` (executable)

---

### Phase 4: Initialize Changelog (30 min)
**Goal**: Create `docs-changelog.jsonl` from git history

**Do**:
1. Parse git log for docs/ modifications
2. Generate JSON entry for each commit:
   - timestamp (ISO8601)
   - commit hash
   - action (created|modified|deleted)
   - domain (inferred from path)
   - file
   - changes (inferred from diff)
3. Write one JSON object per line
4. Commit as baseline

**Acceptance**:
- ✅ Changelog exists
- ✅ Valid newline-delimited JSON
- ✅ Can be read programmatically
- ✅ Covers git history

**Output**: `docs-changelog.jsonl` (committed)

---

### Phase 5: Verification & Testing (1 hour)
**Goal**: Validate entire system works correctly

**Test Cases** (implement each):

1. **Compliant file**:
   - Create file with proper frontmatter, English, correct naming
   - Expected: PASS
   
2. **Missing frontmatter**:
   - Create file without `---` header
   - Expected: FAIL

3. **New domain not in manifest**:
   - Create `docs/new-domain/file.md`
   - Expected: FAIL (domain not authorized)

4. **Spanish content**:
   - Create file with Spanish words ("porque", "usuario")
   - Expected: WARN

5. **Unpaired code block**:
   - Create file with ` ``` ` without closing
   - Expected: WARN

6. **Formatting-only change**:
   - Modify file: change spacing/indentation only
   - Expected: PASS (content validation skipped)

7. **Structure matches manifest**:
   - Run: `validate-docs-incremental.sh docs/`
   - Expected: All files either match manifest or documented as templates
   - Expected: No orphan files or missing required files

**Validation Criteria**:
- ✅ All 7 test cases pass
- ✅ Changelog grows with commits
- ✅ No false positives
- ✅ Git hook prevents bad commits
- ✅ Performance < 1 second

---

## TECHNICAL REQUIREMENTS

### Manifest JSON Schema
```json
{
  "version": "1.0.0",
  "last_updated": "ISO8601_DATE",
  "root_files": {
    "filename.md": {
      "type": "navigation|guide|reference",
      "required": true|false
    }
  },
  "domains": {
    "domain_name": {
      "purpose": "string",
      "required": true|false,
      "template": "template-filename.md",
      "rules": {
        "naming": "pattern",
        "frontmatter_required": true|false,
        "language": "en",
        "min_sections": ["Section1", "Section2"]
      }
    }
  },
  "global_rules": {
    "language": "en_only",
    "no_emojis": true,
    "filename_format": "lowercase-with-hyphens.md",
    "frontmatter_fields": ["type", "id", "title", "version", "status", "date_created", "language", "category"]
  }
}
```

### Changelog Format
```jsonl
{"timestamp":"2026-04-03T17:45:00Z","commit":"abc123","action":"created","domain":"architecture","file":"design.md","checksum":"sha256:xyz"}
```

---

## IMPORTANT NOTES

### What Already Exists (Don't Change)
- ✅ Domain structure (adr/, architecture/, skills/, testing/)
- ✅ Validation rules (English-only, no emojis, frontmatter)
- ✅ Git hook system (enhance, don't replace)
- ✅ Existing docs (leave as-is, will be validated going forward)

### What Is Explicitly Out of Scope For Phase 1
- ❌ Rewriting legacy documentation prose to match the new validator
- ❌ Renaming folders or reclassifying domains
- ❌ Global content cleanup across untouched files
- ❌ Making templates pass the same strict content checks as publishable docs if that blocks the baseline manifest

### What You're Adding (New)
- ✅ Manifest (baseline)
- ✅ Incremental validator (new script)
- ✅ Changelog (new log file)
- ✅ Git hook enhancement (not from scratch)

### What NOT to Do
- ❌ Don't change validation rules (already good)
- ❌ Don't reorganize domains (already organized)
- ❌ Don't validate all existing docs (do it only for new/changed)
- ❌ Don't break backward compatibility

---

## CONTEXT FILE

**Full context**: `docs/HANDOVER-DOCS-VALIDATION-SYSTEM.md`
- Read it completely before starting
- Reference it during implementation
- Use it to answer questions

---

## DELIVERABLES

### After Completion

**Files Created**:
- ✅ `docs-manifest.json` - Structure & rules (committed)
- ✅ `scripts/validate-docs-incremental.sh` - Incremental validator (committed, executable)
- ✅ `docs-changelog.jsonl` - Change log baseline (committed)

**Files Modified**:
- ✅ `.git/hooks/pre-commit` - Enhanced for incremental validation
- ✅ `docs/README.md` - Reference to manifest
- ✅ `docs/STRUCTURE.md` - Reference to manifest

**Tests**:
- ✅ All 7 test cases pass
- ✅ Changelog grows correctly
- ✅ No false positives/negatives

---

## SUCCESS CRITERIA

✅ **Done When**:
- Manifest covers all 5 domains completely
- Incremental validator works on changed files only
- Git hook prevents non-compliant commits
- Changelog bootstrapped and functional
- All tests pass
- Documentation updated
- Performance validated (<1 second)
- Ready for RAG integration

## ALIGNMENT CHECK

Before implementing, confirm these assumptions are true:

- The manifest baseline describes the current repo structure, not an ideal future state.
- Validation is incremental and change-scoped.
- Templates remain available as reference files, even if they require separate handling.
- Legacy wording in untouched docs does not block this phase.

---

## QUESTIONS TO ANSWER WHILE IMPLEMENTING

1. **Manifest**: What should file count targets be? Current ~78, growth to ~100?
2. **Validator**: How to detect "substantive" changes vs formatting? (Skip full validation for spacing)
3. **Changelog**: Should it include "changes" field (what changed in the file)?
4. **Testing**: Should tests be automated scripts or manual?
5. **Performance**: OK with <1 second for normal commits?

---

## WHEN YOU'RE DONE

1. Mark all 5 phases as ✅ complete
2. Run full test suite
3. Verify git hook integration
4. Commit everything
5. Update memory: Document what was implemented
6. Report: "System ready for RAG integration"

---

## NEXT PHASE (Not This Session)

Once this is done, RAG MCP can:
- Read `docs-changelog.jsonl`
- Inject "Recent docs changes" into context
- Know which documentation is relevant
- (That's a future task)

---

## HELP

If stuck:
1. Check `docs/HANDOVER-DOCS-VALIDATION-SYSTEM.md` (full specs)
2. Check `docs/DOCUMENTATION-ENFORCEMENT-GUIDE.md` (validation rules)
3. Look at Phase description above
4. Ask yourself: "What does the manifest need to cover?"
5. Ask yourself: "How does git diff work?"

---

**Ready? Start with Phase 1. Good luck!**
