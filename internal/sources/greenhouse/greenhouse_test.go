package greenhouse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"jobsearcher/internal/sources"
)

func TestFetchAndNormalize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/boards/acme/jobs" || r.URL.Query().Get("content") != "true" {
			t.Errorf("unexpected request %s", r.URL.String())
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jobs":[{"id":42,"title":"Go Backend Engineer","updated_at":"2026-09-29T10:00:00Z","location":{"name":"Remote - Brazil"},"absolute_url":"https://boards.greenhouse.io/acme/jobs/42","content":"<p>Build Go APIs &amp; services</p>"}]}`))
	}))
	defer server.Close()
	src := NewClientWithHTTP(server.Client(), server.URL+"/v1/boards")
	jobs, err := src.FetchJobs(context.Background(), sources.Query{Terms: []string{"Go"}, BoardTokens: []string{"acme"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}
	j := jobs[0]
	if j.ID != "greenhouse:acme:42" || j.Source != "greenhouse" || j.Company != "acme" {
		t.Fatalf("unstable identity/metadata: %+v", j)
	}
	if j.WorkplaceType != "Remote" || j.Location != "Remote - Brazil" || !strings.Contains(j.Description, "Build Go APIs & services") {
		t.Fatalf("normalization mismatch: %+v", j)
	}
}

func TestEmptyBoard(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"jobs":[]}`)) }))
	defer s.Close()
	jobs, err := NewClientWithHTTP(s.Client(), s.URL).FetchJobs(context.Background(), sources.Query{BoardTokens: []string{"empty"}})
	if err != nil || len(jobs) != 0 {
		t.Fatalf("jobs=%v err=%v", jobs, err)
	}
}

func TestRejectsInvalidBoardToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { t.Fatal("invalid token reached HTTP server") }))
	defer server.Close()
	_, err := NewClientWithHTTP(server.Client(), server.URL).FetchJobs(context.Background(), sources.Query{BoardTokens: []string{"../other"}})
	if err == nil || !strings.Contains(err.Error(), "invalid board token") {
		t.Fatalf("expected invalid-token error, got %v", err)
	}
}

func TestNormalizeKeepsWorkplaceUncertainWhenUnspecified(t *testing.T) {
	var dto jobDTO
	dto.ID = 1
	dto.Location.Name = "Maceio, Brazil"
	j := normalize("acme", dto)
	if j.WorkplaceType != "" {
		t.Fatalf("inferred workplace type %q without evidence", j.WorkplaceType)
	}
}

func TestHTTPAndInvalidResponseErrors(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
		want       string
	}{
		{"not found", `{"error":"missing"}`, http.StatusNotFound, "HTTP 404"},
		{"rate limited", `{}`, http.StatusTooManyRequests, "HTTP 429"},
		{"invalid json", `not-json`, http.StatusOK, "decode response"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer s.Close()
			_, err := NewClientWithHTTP(s.Client(), s.URL).FetchJobs(context.Background(), sources.Query{BoardTokens: []string{"board"}})
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v, want %q", err, tc.want)
			}
		})
	}
}

func TestPartialFailureAndCanceledContext(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/bad/") {
			http.Error(w, "missing", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"jobs":[{"id":7,"title":"Engineer"}]}`))
	}))
	defer s.Close()
	jobs, err := NewClientWithHTTP(s.Client(), s.URL).FetchJobs(context.Background(), sources.Query{BoardTokens: []string{"bad", "good"}})
	if err == nil || len(jobs) != 1 {
		t.Fatalf("partial result jobs=%d err=%v", len(jobs), err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = NewClientWithHTTP(s.Client(), s.URL).FetchJobs(ctx, sources.Query{BoardTokens: []string{"good"}})
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected canceled-context error, got %v", err)
	}
}

func TestRequestTimeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(30 * time.Millisecond)
		_, _ = w.Write([]byte(`{"jobs":[]}`))
	}))
	defer s.Close()
	client := NewClientWithHTTP(&http.Client{Timeout: 2 * time.Millisecond}, s.URL)
	_, err := client.FetchJobs(context.Background(), sources.Query{BoardTokens: []string{"slow"}})
	if err == nil || !strings.Contains(err.Error(), "Client.Timeout exceeded") {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
