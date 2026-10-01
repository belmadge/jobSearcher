package main

import (
	"testing"
	"time"

	"jobsearcher/internal/domain"
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
