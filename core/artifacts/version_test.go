package artifacts

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []int
		wantErr bool
	}{
		{"simple", "1.2.3", []int{1, 2, 3}, false},
		{"two components", "2.0", []int{2, 0}, false},
		{"single component", "5", []int{5}, false},
		{"with pre-release", "1.0.0-beta", []int{1, 0, 0}, false},
		{"with build metadata", "1.0.0+build.123", []int{1, 0, 0}, false},
		{"with pre-release and build", "1.2.3-rc.1+meta", []int{1, 2, 3}, false},
		{"leading zeros", "01.02.03", []int{1, 2, 3}, false},
		{"empty", "", nil, true},
		{"non-numeric", "a.b.c", nil, true},
		{"empty component", "1..2", nil, true},
		{"just hyphen", "-", nil, true},
		{"zero version", "0.0.0", []int{0, 0, 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseVersion(%q) = %v, want %v", tt.input, got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseVersion(%q) = %v, want %v", tt.input, got, tt.want)
					return
				}
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{"equal", "1.2.3", "1.2.3", 0},
		{"a less than b", "1.0.0", "2.0.0", -1},
		{"a greater than b", "2.0.0", "1.0.0", 1},
		{"minor version", "1.1.0", "1.2.0", -1},
		{"patch version", "1.2.0", "1.2.1", -1},
		{"different lengths", "1.2", "1.2.0", 0},
		{"shorter is less", "1.0", "1.0.1", -1},
		{"shorter is more", "1.1", "1.0.9", 1},
		{"zero vs one", "0.0.0", "0.0.1", -1},
		{"pre-release equal base", "1.0.0-beta", "1.0.0", 0},
		{"large numbers", "10.0.0", "9.9.9", 1},
		{"equal strings", "latest", "latest", 0},
		{"string comparison", "alpha", "beta", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareVersions(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestMatchVersion(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		want       bool
	}{
		// Exact match
		{"exact match", "1.2.3", "1.2.3", true},
		{"exact mismatch", "1.2.3", "1.2.4", false},

		// Latest
		{"latest", "1.2.3", "latest", true},
		{"empty constraint", "1.2.3", "", true},

		// >=
		{">= true", "1.5.0", ">=1.0.0", true},
		{">= boundary", "1.0.0", ">=1.0.0", true},
		{">= false", "0.9.0", ">=1.0.0", false},

		// <=
		{"<= true", "0.9.0", "<=1.0.0", true},
		{"<= boundary", "1.0.0", "<=1.0.0", true},
		{"<= false", "1.1.0", "<=1.0.0", false},

		// >
		{"> true", "1.1.0", ">1.0.0", true},
		{"> boundary", "1.0.0", ">1.0.0", false},
		{"> false", "0.9.0", ">1.0.0", false},

		// <
		{"< true", "0.9.0", "<1.0.0", true},
		{"< boundary", "1.0.0", "<1.0.0", false},
		{"< false", "1.1.0", "<1.0.0", false},

		// Compound
		{"compound true", "1.5.0", ">=1.0.0,<2.0.0", true},
		{"compound boundary low", "1.0.0", ">=1.0.0,<2.0.0", true},
		{"compound boundary high", "2.0.0", ">=1.0.0,<2.0.0", false},
		{"compound below range", "0.9.0", ">=1.0.0,<2.0.0", false},
		{"compound above range", "2.1.0", ">=1.0.0,<2.0.0", false},

		// Spaces in constraint
		{"compound with spaces", "1.5.0", ">=1.0.0, <2.0.0", true},

		// = operator
		{"= match", "1.2.3", "=1.2.3", true},
		{"= no match", "1.2.3", "=1.2.4", false},

		// != operator
		{"!= true", "1.2.3", "!=1.2.4", true},
		{"!= false", "1.2.3", "!=1.2.3", false},

		// Version padding
		{"padded equal", "1.2", "1.2.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchVersion(tt.version, tt.constraint)
			if got != tt.want {
				t.Errorf("MatchVersion(%q, %q) = %v, want %v", tt.version, tt.constraint, got, tt.want)
			}
		})
	}
}

func TestSelectBestVersion(t *testing.T) {
	tests := []struct {
		name       string
		versions   []string
		constraint string
		want       string
	}{
		{
			name:       "select highest",
			versions:   []string{"1.0.0", "1.5.0", "2.0.0", "1.2.0"},
			constraint: ">=1.0.0",
			want:       "2.0.0",
		},
		{
			name:       "constraint filters",
			versions:   []string{"1.0.0", "1.5.0", "2.0.0"},
			constraint: ">=1.0.0,<2.0.0",
			want:       "1.5.0",
		},
		{
			name:       "no match",
			versions:   []string{"1.0.0", "1.1.0"},
			constraint: ">=2.0.0",
			want:       "",
		},
		{
			name:       "empty versions",
			versions:   []string{},
			constraint: "latest",
			want:       "",
		},
		{
			name:       "latest picks highest",
			versions:   []string{"0.1.0", "1.0.0", "0.9.0"},
			constraint: "latest",
			want:       "1.0.0",
		},
		{
			name:       "exact match",
			versions:   []string{"1.0.0", "2.0.0", "3.0.0"},
			constraint: "2.0.0",
			want:       "2.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectBestVersion(tt.versions, tt.constraint)
			if got != tt.want {
				t.Errorf("SelectBestVersion(%v, %q) = %q, want %q", tt.versions, tt.constraint, got, tt.want)
			}
		})
	}
}
