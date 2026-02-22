package service

import (
	"testing"

	"ragtime-backend/internal/retrieval"
)

func TestResolveProfileAndWeight_ExplicitProfiles(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    float64
	}{
		{name: "exact", profile: retrieval.RetrievalProfileExact, want: 0.2},
		{name: "balanced", profile: retrieval.RetrievalProfileBalanced, want: 0.5},
		{name: "semantic", profile: retrieval.RetrievalProfileSemantic, want: 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, _ := resolveProfileAndWeight(retrieval.Request{RetrievalProfile: tt.profile, Query: "test"})
			if got != tt.want {
				t.Fatalf("resolveProfileAndWeight() weight = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveProfileAndWeight_SemanticWeightOverride(t *testing.T) {
	_, got, signals := resolveProfileAndWeight(retrieval.Request{
		RetrievalProfile:  retrieval.RetrievalProfileExact,
		SemanticWeight:    0.66,
		SemanticWeightSet: true,
		Query:             "where is activation logic",
	})

	if got != 0.66 {
		t.Fatalf("resolveProfileAndWeight() weight = %v, want 0.66", got)
	}
	if len(signals) != 1 || signals[0] != "semantic_weight_override" {
		t.Fatalf("resolveProfileAndWeight() signals = %v, want semantic override", signals)
	}
}

func TestClassifyAutoProfile(t *testing.T) {
	profile, _ := classifyAutoProfile(`error E_CONN_TIMEOUT in src/retrieval/service.go`)
	if profile != retrieval.RetrievalProfileExact {
		t.Fatalf("classifyAutoProfile() = %q, want %q", profile, retrieval.RetrievalProfileExact)
	}

	profile, _ = classifyAutoProfile("how does chunk activation preserve old active versions during failure")
	if profile != retrieval.RetrievalProfileSemantic {
		t.Fatalf("classifyAutoProfile() = %q, want %q", profile, retrieval.RetrievalProfileSemantic)
	}
}
