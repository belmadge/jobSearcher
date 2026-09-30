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
