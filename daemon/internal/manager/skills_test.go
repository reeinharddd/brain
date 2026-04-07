package manager

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestSkillsFromYAML tests loading skills from YAML registry
func TestSkillsFromYAML(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create a test registry.yml
	yamlContent := `skills:
  test-skill-1:
    name: Test Skill 1
    version: 1.0.0
    type: internal
    description: A test skill
    tags:
      - testing
      - demo
    file: skills/test-skill-1/SKILL.md
    sync-to:
      - cli
    maintained: true
  test-skill-2:
    name: Test Skill 2
    version: 2.0.0
    type: external
    description: Another test skill
    tags:
      - testing
    maintained: false
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create an empty TSV to avoid errors
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Verify skills were loaded
	all := registry.GetAll(context.Background())
	if len(all) != 2 {
		t.Errorf("Expected 2 skills, got %d", len(all))
	}

	// Verify skill properties
	skill1 := registry.GetByID(context.Background(), "test-skill-1")
	if skill1 == nil {
		t.Fatal("skill1 not found")
	}
	if skill1.Name != "Test Skill 1" {
		t.Errorf("Expected name 'Test Skill 1', got '%s'", skill1.Name)
	}
	if skill1.Kind != "skill" {
		t.Errorf("Expected kind 'skill', got '%s'", skill1.Kind)
	}
	if skill1.Scope != "global" {
		t.Errorf("Expected scope 'global', got '%s'", skill1.Scope)
	}
	if skill1.Source != "registry.yml" {
		t.Errorf("Expected source 'registry.yml', got '%s'", skill1.Source)
	}

	// Verify backward compatibility: legacy fields still present
	if skill1.Type != "internal" {
		t.Errorf("Expected legacy type 'internal', got '%s'", skill1.Type)
	}
	if len(skill1.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(skill1.Tags))
	}
	if len(skill1.SyncTo) != 1 || skill1.SyncTo[0] != "cli" {
		t.Errorf("Expected sync-to=['cli'], got %v", skill1.SyncTo)
	}
}

// TestContextPacksFromTSV tests loading context-packs from TSV registry
func TestContextPacksFromTSV(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create an empty YAML to avoid errors
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create a test TSV
	tsvContent := `# skill_id	title	detect_tags	context_path	summary
bash-platform	Bash Platform	bash,shell	skills/contexts/bash.md	Bash automation guidance
go-service	Go Service	go	skills/contexts/go.md	Go service best practices
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Verify context-packs were loaded
	all := registry.GetAll(context.Background())
	if len(all) != 2 {
		t.Errorf("Expected 2 context-packs, got %d", len(all))
	}

	// Verify context-pack properties
	pack := registry.GetByID(context.Background(), "bash-platform")
	if pack == nil {
		t.Fatal("bash-platform not found")
	}
	if pack.Name != "Bash Platform" {
		t.Errorf("Expected name 'Bash Platform', got '%s'", pack.Name)
	}
	if pack.Kind != "context-pack" {
		t.Errorf("Expected kind 'context-pack', got '%s'", pack.Kind)
	}
	if pack.Source != "dynamic-registry.tsv" {
		t.Errorf("Expected source 'dynamic-registry.tsv', got '%s'", pack.Source)
	}
	if pack.Path != "skills/contexts/bash.md" {
		t.Errorf("Expected path 'skills/contexts/bash.md', got '%s'", pack.Path)
	}
	if len(pack.Tags) != 2 {
		t.Errorf("Expected 2 tags, got %d: %v", len(pack.Tags), pack.Tags)
	}
	if pack.Tags[0] != "bash" || pack.Tags[1] != "shell" {
		t.Errorf("Expected tags ['bash', 'shell'], got %v", pack.Tags)
	}
}

// TestDuplicateIDDetection tests that duplicate IDs are handled correctly
func TestDuplicateIDDetection(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry.yml with a skill named "duplicate-id"
	yamlContent := `skills:
  duplicate-id:
    name: From YAML
    version: 1.0.0
    type: internal
    description: This is from YAML
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create TSV with same ID
	tsvContent := `# skill_id	title	detect_tags	context_path	summary
duplicate-id	From TSV	test	skills/contexts/test.md	This is from TSV
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	all := registry.GetAll(context.Background())
	if len(all) != 1 {
		t.Errorf("Expected 1 item (YAML wins on duplicate), got %d", len(all))
	}

	item := registry.GetByID(context.Background(), "duplicate-id")
	if item == nil {
		t.Fatal("Item not found")
	}
	// YAML should win (loaded first)
	if item.Source != "registry.yml" {
		t.Errorf("Expected YAML to win on duplicate ID, but source is '%s'", item.Source)
	}
	if item.Name != "From YAML" {
		t.Errorf("Expected 'From YAML', got '%s'", item.Name)
	}
}

// TestBackwardCompatibility tests that legacy fields are preserved in output
func TestBackwardCompatibility(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with legacy fields
	yamlContent := `skills:
  legacy-skill:
    name: Legacy Skill
    version: 1.0.0
    type: internal
    description: Test legacy compatibility
    file: skills/legacy-skill/SKILL.md
    category: testing
    tags:
      - test
    sync-to:
      - cli
      - vscode
    requires:
      - other-skill
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create an empty TSV
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	item := registry.GetByID(context.Background(), "legacy-skill")
	if item == nil {
		t.Fatal("Item not found")
	}

	// Check all legacy fields are preserved
	if item.Type != "internal" {
		t.Errorf("Expected Type='internal', got '%s'", item.Type)
	}
	if item.File != "skills/legacy-skill/SKILL.md" {
		t.Errorf("Expected File, got '%s'", item.File)
	}
	if item.Path != "skills/legacy-skill/SKILL.md" {
		t.Errorf("Expected Path, got '%s'", item.Path)
	}
	if item.Category != "testing" {
		t.Errorf("Expected Category='testing', got '%s'", item.Category)
	}
	if len(item.SyncTo) != 2 {
		t.Errorf("Expected 2 sync targets, got %d", len(item.SyncTo))
	}
	if len(item.Requires) != 1 {
		t.Errorf("Expected 1 requirement, got %d", len(item.Requires))
	}
}

// TestMergedCatalog tests that YAML and TSV are properly merged
func TestMergedCatalog(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with 2 skills
	yamlContent := `skills:
  skill-1:
    name: Skill 1
    version: 1.0.0
    type: internal
    description: First skill
  skill-2:
    name: Skill 2
    version: 2.0.0
    type: external
    description: Second skill
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create TSV with 2 context-packs
	tsvContent := `# skill_id	title	detect_tags	context_path	summary
pack-1	Pack 1	test	skills/contexts/pack1.md	First pack
pack-2	Pack 2	test	skills/contexts/pack2.md	Second pack
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	all := registry.GetAll(context.Background())
	if len(all) != 4 {
		t.Errorf("Expected 4 items (2 skills + 2 packs), got %d", len(all))
	}

	// Count by kind
	skills := 0
	packs := 0
	for _, item := range all {
		if item.Kind == "skill" {
			skills++
		} else if item.Kind == "context-pack" {
			packs++
		}
	}

	if skills != 2 {
		t.Errorf("Expected 2 skills, got %d", skills)
	}
	if packs != 2 {
		t.Errorf("Expected 2 context-packs, got %d", packs)
	}
}

// TestSearchFunctionality tests search across merged catalog
func TestSearchFunctionality(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with tagged skills
	yamlContent := `skills:
  testing-skill:
    name: Testing Skill
    version: 1.0.0
    type: internal
    description: For testing purposes
    tags:
      - testing
      - quality
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create TSV with tagged context-pack
	tsvContent := `# skill_id	title	detect_tags	context_path	summary
testing-pack	Testing Pack	testing,quality	skills/contexts/test.md	Testing-focused pack
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Search by tag
	results := registry.Search(context.Background(), "testing")
	if len(results) != 2 {
		t.Errorf("Expected 2 results when searching for 'testing', got %d", len(results))
	}

	// Search by description
	results = registry.Search(context.Background(), "purposes")
	if len(results) != 1 {
		t.Errorf("Expected 1 result when searching for 'purposes', got %d", len(results))
	}
}

// TestEmptyRegistries tests behavior when registry files are empty
func TestEmptyRegistries(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create empty registries
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load empty registries: %v", err)
	}

	all := registry.GetAll(context.Background())
	if len(all) != 0 {
		t.Errorf("Expected 0 items from empty registries, got %d", len(all))
	}
}

// TestTSVWithComments tests that TSV comments are properly skipped
func TestTSVWithComments(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create empty YAML
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create TSV with comment lines
	tsvContent := `# This is a comment
# skill_id	title	detect_tags	context_path	summary
# Another comment

valid-pack	Valid Pack	test	skills/contexts/pack.md	Valid entry

# Another comment in middle
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load TSV with comments: %v", err)
	}

	all := registry.GetAll(context.Background())
	if len(all) != 1 {
		t.Errorf("Expected 1 item after filtering comments, got %d", len(all))
	}

	item := registry.GetByID(context.Background(), "valid-pack")
	if item == nil {
		t.Fatal("valid-pack not found")
	}
	if item.Name != "Valid Pack" {
		t.Errorf("Expected 'Valid Pack', got '%s'", item.Name)
	}
}

// TestGetByTarget tests filtering by sync target
func TestGetByTarget(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with different sync targets
	yamlContent := `skills:
  cli-skill:
    name: CLI Skill
    version: 1.0.0
    type: internal
    description: For CLI
    sync-to:
      - cli
  multi-skill:
    name: Multi Skill
    version: 1.0.0
    type: internal
    description: For multiple targets
    sync-to:
      - cli
      - vscode
      - cursor
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	// Create empty TSV
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Get skills for specific target
	cliSkills := registry.GetByTarget(context.Background(), "cli")
	if len(cliSkills) != 2 {
		t.Errorf("Expected 2 skills for 'cli' target, got %d", len(cliSkills))
	}

	vsxodeSkills := registry.GetByTarget(context.Background(), "vscode")
	if len(vsxodeSkills) != 1 {
		t.Errorf("Expected 1 skill for 'vscode' target, got %d", len(vsxodeSkills))
	}
}

// TestStatus tests the GetStatus method
func TestStatus(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with 3 items
	yamlContent := `skills:
  skill-1:
    name: Skill 1
    version: 1.0.0
    type: internal
    description: First
  skill-2:
    name: Skill 2
    version: 1.0.0
    type: internal
    description: Second
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	tsvContent := `# skill_id	title	detect_tags	context_path	summary
pack-1	Pack 1	test	skills/contexts/pack1.md	First pack
`
	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(tsvContent), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	status := registry.GetStatus(context.Background())
	count, ok := status["count"].(int)
	if !ok {
		t.Errorf("Expected count to be int, got %T", status["count"])
	}
	if count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

// TestCreateSkill tests creating a new skill via CRUD
func TestCreateSkill(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory with empty files
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	// Load empty registry
	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create a new skill
	newSkill := &CatalogItem{
		ID:          "new-test-skill",
		Name:        "New Test Skill",
		Kind:        "skill",
		Scope:       "global",
		Description: "A newly created skill",
		Version:     "1.0.0",
		Type:        "internal",
		File:        "skills/new-test-skill/SKILL.md",
		Tags:        []string{"test"},
		Maintained:  true,
	}

	if err := registry.CreateItem(context.Background(), newSkill); err != nil {
		t.Fatalf("Failed to create skill: %v", err)
	}

	// Verify skill exists in catalog
	retrieved := registry.GetByID(context.Background(), "new-test-skill")
	if retrieved == nil {
		t.Fatal("Created skill not found in catalog")
	}
	if retrieved.Name != "New Test Skill" {
		t.Errorf("Expected name 'New Test Skill', got '%s'", retrieved.Name)
	}

	// Verify YAML file was updated
	yamlData, err := os.ReadFile(filepath.Join(skillsDir, "registry.yml"))
	if err != nil {
		t.Fatalf("Failed to read updated registry.yml: %v", err)
	}

	var rawData map[string]interface{}
	if err := yaml.Unmarshal(yamlData, &rawData); err != nil {
		t.Fatalf("Failed to parse updated YAML: %v", err)
	}

	skillsMap, ok := rawData["skills"].(map[string]interface{})
	if !ok {
		t.Fatal("Skills section not found in updated YAML")
	}

	if _, exists := skillsMap["new-test-skill"]; !exists {
		t.Fatal("New skill not persisted in YAML file")
	}
}

// TestUpdateSkill tests updating an existing skill
func TestUpdateSkill(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with one skill
	yamlContent := `skills:
  existing-skill:
    name: Original Name
    version: 1.0.0
    type: internal
    description: Original description
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Update the skill
	updated := &CatalogItem{
		ID:          "existing-skill",
		Name:        "Updated Name",
		Description: "Updated description",
		Version:     "2.0.0",
	}

	if err := registry.UpdateItem(context.Background(), "existing-skill", updated); err != nil {
		t.Fatalf("Failed to update skill: %v", err)
	}

	// Verify updated
	retrieved := registry.GetByID(context.Background(), "existing-skill")
	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", retrieved.Name)
	}
	if retrieved.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", retrieved.Version)
	}
}

// TestDeleteSkill tests deleting a skill
func TestDeleteSkill(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	// Create registry with one skill
	yamlContent := `skills:
  to-delete:
    name: Delete Me
    version: 1.0.0
    type: internal
    description: Will be deleted
`
	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Verify skill exists
	if registry.GetByID(context.Background(), "to-delete") == nil {
		t.Fatal("Skill not found before deletion")
	}

	// Delete the skill
	if err := registry.DeleteItem(context.Background(), "to-delete"); err != nil {
		t.Fatalf("Failed to delete skill: %v", err)
	}

	// Verify skill removed
	if registry.GetByID(context.Background(), "to-delete") != nil {
		t.Fatal("Skill still exists after deletion")
	}

	// Verify catalog count
	all := registry.GetAll(context.Background())
	if len(all) != 0 {
		t.Errorf("Expected 0 items after deletion, got %d", len(all))
	}
}

// TestCreateContextPack tests creating a new context-pack via CRUD
func TestCreateContextPack(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create a new context-pack
	newPack := &CatalogItem{
		ID:          "new-context-pack",
		Name:        "New Context Pack",
		Kind:        "context-pack",
		Scope:       "global",
		Description: "A newly created context pack",
		Path:        "skills/contexts/new-context.md",
		Tags:        []string{"test", "context"},
		Maintained:  true,
	}

	if err := registry.CreateItem(context.Background(), newPack); err != nil {
		t.Fatalf("Failed to create context-pack: %v", err)
	}

	// Verify context-pack exists
	retrieved := registry.GetByID(context.Background(), "new-context-pack")
	if retrieved == nil {
		t.Fatal("Created context-pack not found in catalog")
	}
	if retrieved.Kind != "context-pack" {
		t.Errorf("Expected kind 'context-pack', got '%s'", retrieved.Kind)
	}

	// Verify TSV file was updated
	tsvData, err := os.ReadFile(filepath.Join(skillsDir, "dynamic-registry.tsv"))
	if err != nil {
		t.Fatalf("Failed to read updated TSV: %v", err)
	}

	if !strings.Contains(string(tsvData), "new-context-pack") {
		t.Fatal("New context-pack not persisted in TSV file")
	}
}

// TestDuplicateCreateFails tests that creating duplicate ID fails
func TestDuplicateCreateFails(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Create first skill
	skill1 := &CatalogItem{
		ID:   "duplicate-id",
		Name: "First Skill",
		Kind: "skill",
	}

	if err := registry.CreateItem(context.Background(), skill1); err != nil {
		t.Fatalf("Failed to create first skill: %v", err)
	}

	// Try to create skill with same ID
	skill2 := &CatalogItem{
		ID:   "duplicate-id",
		Name: "Second Skill",
		Kind: "skill",
	}

	if err := registry.CreateItem(context.Background(), skill2); err == nil {
		t.Fatal("Expected error when creating duplicate ID, but got none")
	}
}

// TestUpdateNonExistentFails tests that updating non-existent skill fails
func TestUpdateNonExistentFails(t *testing.T) {
	brainRoot := t.TempDir()

	// Create skills directory
	skillsDir := filepath.Join(brainRoot, "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatalf("Failed to create skills dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "registry.yml"), []byte("skills: {}"), 0644); err != nil {
		t.Fatalf("Failed to write registry.yml: %v", err)
	}

	if err := os.WriteFile(filepath.Join(skillsDir, "dynamic-registry.tsv"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write dynamic-registry.tsv: %v", err)
	}

	logCh := make(chan string, 100)
	registry := NewSkillsRegistry(brainRoot, logCh)

	if err := registry.Load(context.Background()); err != nil {
		t.Fatalf("Failed to load registry: %v", err)
	}

	// Try to update non-existent skill
	phantom := &CatalogItem{
		ID:   "non-existent",
		Name: "Ghost Skill",
	}

	if err := registry.UpdateItem(context.Background(), "non-existent", phantom); err == nil {
		t.Fatal("Expected error when updating non-existent skill, but got none")
	}
}
