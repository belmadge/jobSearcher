package matching

import (
	"testing"
	"jobsearcher/internal/config"
)

func TestBuildSearchProfileExpandsBackendRole(t *testing.T) {
	p := BuildSearchProfile(config.Profile{Titles: []string{"Backend Developer"}, Technologies: []string{"Go","AWS"}})
	for _, want := range []string{"Backend Developer","Backend Engineer","API Engineer"} {
		if !containsNormalized(p.ExpandedRoles,want) { t.Fatalf("expected role %q in %v",want,p.ExpandedRoles) }
	}
}

func TestBuildSearchProfileExpandsQA(t *testing.T) {
	p := BuildSearchProfile(config.Profile{Titles: []string{"QA"}})
	for _, want := range []string{"QA","QA Engineer","Quality Assurance","SDET","Test Automation Engineer"} {
		if !containsNormalized(p.ExpandedRoles,want) { t.Fatalf("expected QA expansion %q in %v",want,p.ExpandedRoles) }
	}
}

func TestBuildSearchProfileExpandsSkills(t *testing.T) {
	p := BuildSearchProfile(config.Profile{Technologies: []string{"Golang","AWS","K8s","Postgres"}})
	for _, want := range []string{"Golang","Go","Amazon Web Services","Kubernetes","PostgreSQL"} {
		if !containsNormalized(p.ExpandedSkills,want) { t.Fatalf("expected skill expansion %q in %v",want,p.ExpandedSkills) }
	}
}

func TestBuildSearchProfilePreservesCustomTerms(t *testing.T) {
	p := BuildSearchProfile(config.Profile{Titles: []string{"Payment Platform Engineer"}, Technologies: []string{"OpenTelemetry"}})
	if !containsNormalized(p.ExpandedRoles,"Payment Platform Engineer") { t.Fatal("custom role was lost") }
	if !containsNormalized(p.ExpandedSkills,"OpenTelemetry") { t.Fatal("custom skill was lost") }
}

func containsNormalized(values []string,want string) bool {
	for _, value := range values { if normalizeRoleText(value)==normalizeRoleText(want) { return true } }
	return false
}


func TestBuildSearchProfileBroadensRetrievalByRoleFamily(t *testing.T) {
	cases := []struct {
		role string
		want []string
	}{
		{"Backend Developer", []string{"Software Engineer", "Software Developer", "API Engineer"}},
		{"QA", []string{"SDET", "Test Automation Engineer", "Test Automation"}},
		{"DevOps", []string{"SRE", "Platform Engineer", "Infrastructure Engineer"}},
	}
	for _, tc := range cases {
		p := BuildSearchProfile(config.Profile{Titles: []string{tc.role}})
		for _, want := range tc.want {
			if !containsNormalized(p.ExpandedRoles, want) {
				t.Fatalf("%q: expected broad discovery term %q in %v", tc.role, want, p.ExpandedRoles)
			}
		}
	}
}

func TestDiscoveryTermsPrioritizeRolesAndStayBounded(t *testing.T) {
	profile := BuildSearchProfile(config.Profile{
		Titles:       []string{"Backend Developer"},
		Technologies: []string{"Go", "PostgreSQL", "AWS", "Docker", "Kubernetes"},
	})

	if len(profile.DiscoveryTerms) == 0 {
		t.Fatal("expected discovery terms")
	}
	if len(profile.DiscoveryTerms) > maxDiscoveryTerms {
		t.Fatalf("discovery terms exceeded budget: got %d, want <= %d", len(profile.DiscoveryTerms), maxDiscoveryTerms)
	}
	if profile.DiscoveryTerms[0] != "Backend Developer" {
		t.Fatalf("explicit role should be the first discovery term, got %q", profile.DiscoveryTerms[0])
	}
}

func TestDiscoveryTermsDeduplicateNormalizedValues(t *testing.T) {
	profile := BuildSearchProfile(config.Profile{
		Titles:       []string{"Backend Developer", "backend developer"},
		Technologies: []string{"Go", "golang"},
	})

	seen := map[string]bool{}
	for _, term := range profile.DiscoveryTerms {
		key := normalizeRoleText(term)
		if seen[key] {
			t.Fatalf("duplicate discovery term: %q", term)
		}
		seen[key] = true
	}
}
