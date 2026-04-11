package context

// ContextLayer represents one layer in the context bundle
type ContextLayer struct {
	ID            int               // 0-12
	Name          string            // Human-readable name
	Content       string            // Layer content text
	TokenCount    int               // Estimated token count
	Compressible  bool              // Whether this layer can be compressed
	AlwaysInclude bool              // Layers 0-1 must always be included
	Priority      int               // Higher = more important to keep
	Tags          []string          // Labels for filtering
	Metadata      map[string]string // Layer-specific metadata
}

// Layer definitions
const (
	LayerHardPolicy    = 0  // Never compressed, always included
	LayerIdentity      = 1  // Identity & security, always included
	LayerOrgBaseline   = 2  // Org-wide conventions
	LayerUserBaseline  = 3  // Personal preferences
	LayerWorkspace     = 4  // Project architecture
	LayerProject       = 5  // Module structure
	LayerTaskLocal     = 6  // Current goal, active files
	LayerActiveSkills  = 7  // Skill frontmatter, full on trigger
	LayerActiveMCPs    = 8  // Tool definitions
	LayerMemoryStruct  = 9  // Structured memory
	LayerMemorySemantic = 10 // Semantic matches
	LayerMemoryEpisodic = 11 // Recent session events
	LayerRuntimeEphemer = 12 // Current session state

	MaxLayerID = 12
)

// LayerDefinitions returns all 13 layer definitions with metadata
func LayerDefinitions() []ContextLayer {
	return []ContextLayer{
		{
			ID:            LayerHardPolicy,
			Name:          "Hard Policy",
			Content:       "",
			TokenCount:    0,
			Compressible:  false,
			AlwaysInclude: true,
			Priority:      100,
			Tags:          []string{"policy", "system", "mandatory"},
			Metadata:      map[string]string{"compressible": "false", "disclosure": "full"},
		},
		{
			ID:            LayerIdentity,
			Name:          "Identity",
			Content:       "",
			TokenCount:    0,
			Compressible:  false,
			AlwaysInclude: true,
			Priority:      95,
			Tags:          []string{"identity", "security", "mandatory"},
			Metadata:      map[string]string{"compressible": "false", "disclosure": "full"},
		},
		{
			ID:            LayerOrgBaseline,
			Name:          "Organization Baseline",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      80,
			Tags:          []string{"org", "conventions"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerUserBaseline,
			Name:          "User Baseline",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      75,
			Tags:          []string{"user", "preferences"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerWorkspace,
			Name:          "Workspace",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      70,
			Tags:          []string{"workspace", "architecture"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerProject,
			Name:          "Project",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      65,
			Tags:          []string{"project", "modules"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerTaskLocal,
			Name:          "Task Local",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      60,
			Tags:          []string{"task", "current"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "full"},
		},
		{
			ID:            LayerActiveSkills,
			Name:          "Active Skills",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      55,
			Tags:          []string{"skills", "on-demand"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "progressive"},
		},
		{
			ID:            LayerActiveMCPs,
			Name:          "Active MCPs",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      50,
			Tags:          []string{"mcp", "tools", "on-demand"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "progressive"},
		},
		{
			ID:            LayerMemoryStruct,
			Name:          "Memory Structured",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      45,
			Tags:          []string{"memory", "structured"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerMemorySemantic,
			Name:          "Memory Semantic",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      40,
			Tags:          []string{"memory", "semantic"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerMemoryEpisodic,
			Name:          "Memory Episodic",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      35,
			Tags:          []string{"memory", "episodic", "recent"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
		{
			ID:            LayerRuntimeEphemer,
			Name:          "Runtime Ephemeral",
			Content:       "",
			TokenCount:    0,
			Compressible:  true,
			AlwaysInclude: false,
			Priority:      30,
			Tags:          []string{"runtime", "session", "ephemeral"},
			Metadata:      map[string]string{"compressible": "true", "disclosure": "summary"},
		},
	}
}
