package environment

import "testing"

func TestNormalize(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "development", raw: "development", want: Development},
		{name: "staging", raw: "staging", want: Staging},
		{name: "production", raw: "production", want: Production},
		{name: "empty defaults to production", raw: "", want: Production},
		{name: "unknown defaults to production", raw: "something-else", want: Production},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Normalize(tt.raw); got != tt.want {
				t.Fatalf("Normalize(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestAllowsVisibility(t *testing.T) {
	if !AllowsVisibility(ProdSafe, Production) {
		t.Fatal("prod-safe components should be allowed in production")
	}

	if AllowsVisibility(DevOnly, Production) {
		t.Fatal("dev-only components must not be allowed in production")
	}

	if !AllowsVisibility(DevOnly, Development) {
		t.Fatal("dev-only components should be allowed in development")
	}
}
