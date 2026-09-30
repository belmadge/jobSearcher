package matching

import (
	"jobsearcher/internal/config"
	"jobsearcher/internal/domain"
	"testing"
)

func TestScoreAndGaps(t *testing.T) {
	p := config.Profile{Technologies: []string{"Go", "PostgreSQL", "GCP"}, Weights: config.Weights{Technical: 35, Responsibility: 20, Seniority: 15, Cloud: 10, Domain: 10, Language: 5, AI: 5}, Domains: []string{"fintech"}, EmergingSkills: []string{"AI Agents"}}
	j := domain.Job{Title: "Backend Engineer", Seniority: "Mid-level", Description: "Go APIs using PostgreSQL and AWS"}
	Score(&j, p)
	if j.FitScore <= 0 {
		t.Fatal("expected positive fit score")
	}
	if len(j.Gaps) != 1 || j.Gaps[0] != "GCP" {
		t.Fatalf("expected GCP to be reported as a gap; got %v", j.Gaps)
	}
	if j.SeniorityMatch != 85 {
		t.Fatalf("expected compatible default software engineer level, got %d", j.SeniorityMatch)
	}
}
