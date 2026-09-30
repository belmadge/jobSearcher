package himalayas

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobsearcher/internal/sources"
)

func TestFetchJobs(t *testing.T) {
	page := `{"jobs":[{"guid":"1","title":"Backend Engineer","companyName":"Acme","employmentType":"Full Time","seniority":["Mid-level"],"locationRestrictions":[{"name":"Brazil","alpha2":"BR"}],"description":"<p>Go APIs and PostgreSQL</p>","pubDate":1720000000000,"applicationLink":"https://example.com/job"}]}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()
	c := NewClient(); c.URL = srv.URL
	jobs, err := c.FetchJobs(context.Background(), sources.Query{Terms: []string{"backend"}})
	if err != nil { t.Fatal(err) }
	if len(jobs) != 1 || jobs[0].Company != "Acme" { t.Fatalf("unexpected jobs: %+v", jobs) }
}
