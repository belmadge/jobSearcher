package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"jobsearcher/internal/domain"
	"jobsearcher/internal/sources"
)

func TestIsFreshUsesUpdatedDate(t *testing.T) {
	fresh := domain.Job{UpdatedAt: time.Now().Add(-24 * time.Hour).Format(time.RFC3339)}
	if !isFresh(fresh, 7) {
		t.Fatal("expected recently updated job to be fresh")
	}
	stale := domain.Job{UpdatedAt: time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339)}
	if isFresh(stale, 7) {
		t.Fatal("expected old job to be stale")
	}
}

func TestIsFreshFallsBackToPostedDate(t *testing.T) {
	job := domain.Job{PostedAt: time.Now().Add(-2 * 24 * time.Hour).Format("2006-01-02")}
	if !isFresh(job, 7) {
		t.Fatal("expected recent posted job to be fresh")
	}
}

func TestIsFreshKeepsUnknownDates(t *testing.T) {
	job := domain.Job{}
	if !isFresh(job, 7) {
		t.Fatal("jobs without parseable dates should not be discarded")
	}
}

func TestIsFreshUsesNewestOfUpdatedAndPostedDates(t *testing.T) {
	job := domain.Job{
		UpdatedAt: time.Now().Add(-10 * 24 * time.Hour).Format(time.RFC3339),
		PostedAt:  time.Now().Add(-2 * 24 * time.Hour).Format("2006-01-02"),
	}
	if !isFresh(job, 7) {
		t.Fatal("expected recent posted date to keep the job fresh")
	}
}

func TestLimitJobsPerSourceKeepsRankedOrder(t *testing.T) {
	jobs := []domain.Job{
		{Source: "greenhouse", FitScore: 95, Title: "Best"},
		{Source: "greenhouse", FitScore: 90, Title: "Second"},
		{Source: "lever", FitScore: 88, Title: "Other"},
		{Source: "greenhouse", FitScore: 70, Title: "Discarded"},
	}
	got := limitJobsPerSource(jobs, 2)
	if len(got) != 3 {
		t.Fatalf("expected 3 jobs after per-source limit, got %d", len(got))
	}
	if got[0].Title != "Best" || got[1].Title != "Second" || got[2].Title != "Other" {
		t.Fatalf("expected limit to preserve ranked order, got %+v", got)
	}
}

func TestLimitJobsPerSourceDoesNotLimitWhenDisabled(t *testing.T) {
	jobs := []domain.Job{
		{Source: "greenhouse", Title: "A"},
		{Source: "greenhouse", Title: "B"},
	}
	got := limitJobsPerSource(jobs, 0)
	if len(got) != len(jobs) {
		t.Fatalf("expected disabled limit to preserve all jobs, got %d", len(got))
	}
}

func TestRankAndDedupeJobsKeepsBestDuplicate(t *testing.T) {
	jobs := []domain.Job{
		{Source: "remoteok", URL: "https://example.com/jobs/42/", FitScore: 60, RoleMatch: 60, Title: "Backend Engineer", Company: "Acme"},
		{Source: "greenhouse", URL: "https://example.com/jobs/42?utm_source=feed", FitScore: 90, RoleMatch: 95, Title: "Backend Engineer", Company: "Acme"},
		{Source: "lever", URL: "https://example.com/jobs/99", FitScore: 80, RoleMatch: 80, Title: "Other Backend Engineer", Company: "Acme"},
	}
	got := rankAndDedupeJobs(jobs)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique jobs, got %d", len(got))
	}
	if got[0].FitScore != 90 {
		t.Fatalf("expected highest-scoring duplicate to survive, got %+v", got[0])
	}
	if got[1].FitScore != 80 {
		t.Fatalf("expected second unique job to remain ranked, got %+v", got[1])
	}
}

func TestWebJobEligibleRequiresExplicitGlobalRemoteCompatibility(t *testing.T) {
	job := domain.Job{
		WorkplaceType:    "remote",
		LocationEligible: domain.LocationEligible,
		FitScore:         80,
		SeniorityMatch:   85,
	}
	if !webJobEligible(job, 55) {
		t.Fatal("expected eligible remote job to pass")
	}
}

func TestWebJobEligibleRejectsNonRemoteWorkplace(t *testing.T) {
	job := domain.Job{
		WorkplaceType:    "hybrid",
		LocationEligible: domain.LocationEligible,
		FitScore:         90,
		SeniorityMatch:   90,
	}
	if webJobEligible(job, 55) {
		t.Fatal("expected hybrid job to be rejected")
	}
}

func TestWebJobEligibleRejectsUnknownLocation(t *testing.T) {
	job := domain.Job{
		WorkplaceType:    "remote",
		LocationEligible: domain.LocationUnknown,
		FitScore:         90,
		SeniorityMatch:   90,
	}
	if webJobEligible(job, 55) {
		t.Fatal("expected geographically uncertain job to be rejected")
	}
}

func TestWebJobEligibleRejectsBelowMinimumScore(t *testing.T) {
	job := domain.Job{
		WorkplaceType:    "remote",
		LocationEligible: domain.LocationEligible,
		FitScore:         54,
		SeniorityMatch:   90,
	}
	if webJobEligible(job, 55) {
		t.Fatal("expected low-fit job to be rejected")
	}
}

func TestWebJobEligibleRejectsOutOfTargetSeniority(t *testing.T) {
	job := domain.Job{
		WorkplaceType:    "remote",
		LocationEligible: domain.LocationEligible,
		FitScore:         90,
		SeniorityMatch:   20,
	}
	if webJobEligible(job, 55) {
		t.Fatal("expected out-of-target seniority job to be rejected")
	}
}


func TestMergeSourceResultsKeepsSuccessfulSourcesWhenOneFails(t *testing.T) {
	jobs, err := mergeSourceResults(
		sourceResult{jobs: []domain.Job{{Source: "remoteok", Title: "Backend Engineer", URL: "https://example.com/remote"}}},
		sourceResult{err: errors.New("remotive unavailable")},
		sourceResult{jobs: []domain.Job{{Source: "himalayas", Title: "Go Engineer", URL: "https://example.com/go"}}},
	)
	if err == nil {
		t.Fatal("expected partial failure to be reported")
	}
	if len(jobs) != 2 {
		t.Fatalf("expected successful source jobs to be preserved, got %d", len(jobs))
	}
}

func TestMergeSourceResultsFailsWhenAllSourcesFail(t *testing.T) {
	jobs, err := mergeSourceResults(
		sourceResult{err: errors.New("remoteok unavailable")},
		sourceResult{err: errors.New("remotive unavailable")},
	)
	if err == nil {
		t.Fatal("expected error when all sources fail")
	}
	if jobs != nil {
		t.Fatalf("expected no jobs when every source fails, got %d", len(jobs))
	}
}

func TestMergeSourceResultsReturnsJobsWithoutErrorWhenAllSucceed(t *testing.T) {
	jobs, err := mergeSourceResults(
		sourceResult{jobs: []domain.Job{{Source: "remoteok", Title: "A", URL: "https://example.com/a"}}},
		sourceResult{jobs: []domain.Job{{Source: "remotive", Title: "B", URL: "https://example.com/b"}}},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestSanitizeJobsRejectsIncompleteAndFillsIdentity(t *testing.T) {
	jobs := []domain.Job{
		{Source: "remoteok", Title: "Backend Engineer", ApplyURL: "https://example.com/job/1"},
		{Source: "remoteok", URL: "https://example.com/job/2"},
		{Source: "", Title: "Missing Source", URL: "https://example.com/job/3"},
		{Source: "remoteok", Title: "Duplicate", URL: "https://example.com/job/1"},
	}
	got := sanitizeJobs(jobs)
	if len(got) != 1 {
		t.Fatalf("expected one valid unique job, got %d: %+v", len(got), got)
	}
	if got[0].URL != got[0].ApplyURL || got[0].ID == "" || got[0].CanonicalURL == "" {
		t.Fatalf("expected normalized URL and generated identity, got %+v", got[0])
	}
}

func TestSanitizeJobsPreservesExplicitIdentity(t *testing.T) {
	job := domain.Job{
		ID: "remoteok:42",
		Source: "remoteok",
		Title: "Backend Engineer",
		URL: "https://example.com/jobs/42",
		ApplyURL: "https://example.com/jobs/42/apply",
		CanonicalURL: "https://example.com/jobs/42",
	}
	got := sanitizeJobs([]domain.Job{job})
	if len(got) != 1 || got[0].ID != job.ID || got[0].ApplyURL != job.ApplyURL {
		t.Fatalf("expected explicit identity to remain unchanged, got %+v", got)
	}
}

func TestFetchSourcesMockStillReturnsNormalizedJobs(t *testing.T) {
	jobs, err := fetchSources(context.Background(), "mock", sources.Query{})
	if err != nil {
		t.Fatalf("expected mock source to succeed, got %v", err)
	}
	if len(jobs) == 0 {
		t.Fatal("expected mock source jobs")
	}
	for _, job := range jobs {
		if job.Source == "" || job.Title == "" || job.URL == "" || job.ID == "" {
			t.Fatalf("expected normalized mock job, got %+v", job)
		}
	}
}

func TestValidWebSeniority(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"junior", true},
		{"mid", true},
		{"senior", true},
		{"staff", true},
		{"", false},
		{"Senior", false},
		{"lead", false},
	}
	for _, tt := range tests {
		if got := validWebSeniority(tt.value); got != tt.want {
			t.Fatalf("validWebSeniority(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}
