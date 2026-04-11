package context

import (
	"testing"
)

func TestLayerDefinitions(t *testing.T) {
	t.Run("returns 13 layers", func(t *testing.T) {
		layers := LayerDefinitions()
		if got := len(layers); got != 13 {
			t.Fatalf("expected 13 layers, got %d", got)
		}
	})

	t.Run("layers 0-1 have AlwaysInclude=true", func(t *testing.T) {
		layers := LayerDefinitions()
		for i := 0; i <= 1; i++ {
			if !layers[i].AlwaysInclude {
				t.Errorf("layer %d (%s) should have AlwaysInclude=true", i, layers[i].Name)
			}
		}
	})

	t.Run("layers 2-12 have AlwaysInclude=false", func(t *testing.T) {
		layers := LayerDefinitions()
		for i := 2; i <= 12; i++ {
			if layers[i].AlwaysInclude {
				t.Errorf("layer %d (%s) should have AlwaysInclude=false", i, layers[i].Name)
			}
		}
	})

	t.Run("layer ordering is correct", func(t *testing.T) {
		layers := LayerDefinitions()
		expectedOrder := []int{
			LayerHardPolicy,
			LayerIdentity,
			LayerOrgBaseline,
			LayerUserBaseline,
			LayerWorkspace,
			LayerProject,
			LayerTaskLocal,
			LayerActiveSkills,
			LayerActiveMCPs,
			LayerMemoryStruct,
			LayerMemorySemantic,
			LayerMemoryEpisodic,
			LayerRuntimeEphemer,
		}

		for i, expectedID := range expectedOrder {
			if layers[i].ID != expectedID {
				t.Errorf("layer[%d] ID: expected %d, got %d", i, expectedID, layers[i].ID)
			}
		}
	})

	t.Run("compressible flags are set correctly", func(t *testing.T) {
		tests := []struct {
			layerID      int
			compressible bool
		}{
			{LayerHardPolicy, false},
			{LayerIdentity, false},
			{LayerOrgBaseline, true},
			{LayerUserBaseline, true},
			{LayerWorkspace, true},
			{LayerProject, true},
			{LayerTaskLocal, true},
			{LayerActiveSkills, true},
			{LayerActiveMCPs, true},
			{LayerMemoryStruct, true},
			{LayerMemorySemantic, true},
			{LayerMemoryEpisodic, true},
			{LayerRuntimeEphemer, true},
		}

		layers := LayerDefinitions()
		for _, tt := range tests {
			t.Run(layers[tt.layerID].Name, func(t *testing.T) {
				if layers[tt.layerID].Compressible != tt.compressible {
					t.Errorf("layer %d compressible: expected %v, got %v",
						tt.layerID, tt.compressible, layers[tt.layerID].Compressible)
				}
			})
		}
	})

	t.Run("each layer has a unique ID", func(t *testing.T) {
		layers := LayerDefinitions()
		seen := make(map[int]bool)
		for _, layer := range layers {
			if seen[layer.ID] {
				t.Errorf("duplicate layer ID: %d", layer.ID)
			}
			seen[layer.ID] = true
		}
	})

	t.Run("each layer has non-empty metadata", func(t *testing.T) {
		layers := LayerDefinitions()
		for _, layer := range layers {
			if len(layer.Metadata) == 0 {
				t.Errorf("layer %d (%s) has no metadata", layer.ID, layer.Name)
			}
		}
	})

	t.Run("each layer has at least one tag", func(t *testing.T) {
		layers := LayerDefinitions()
		for _, layer := range layers {
			if len(layer.Tags) == 0 {
				t.Errorf("layer %d (%s) has no tags", layer.ID, layer.Name)
			}
		}
	})

	t.Run("priority decreases from layer 0 to 12", func(t *testing.T) {
		layers := LayerDefinitions()
		for i := 0; i < len(layers)-1; i++ {
			if layers[i].Priority < layers[i+1].Priority {
				t.Errorf("layer %d priority (%d) should be >= layer %d priority (%d)",
					i, layers[i].Priority, i+1, layers[i+1].Priority)
			}
		}
	})
}
