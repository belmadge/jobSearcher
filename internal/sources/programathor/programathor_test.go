package programathor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"jobsearcher/internal/sources"
)

func TestFetchJobs(t *testing.T) {
	page := `<a href="/jobs/123-backend-go-developer">Backend Go DeveloperAcmeRemotoPlenoPJ APIGoPostgreSQL</a>
<a href="/jobs/456-frontend-react-developer">Vencida Frontend React DeveloperAcmeRemotoPlenoPJ</a>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()

	c := NewClient()
	c.URL = srv.URL
	jobs, err := c.FetchJobs(context.Background(), sources.Query{})
	if err != nil { t.Fatal(err) }
	if len(jobs) != 1 { t.Fatalf("expected 1 active job, got %d", len(jobs)) }
	if jobs[0].Title != "Backend Go Developer" { t.Fatalf("unexpected title: %q", jobs[0].Title) }
	if jobs[0].WorkplaceType != "Remote" { t.Fatalf("unexpected workplace: %q", jobs[0].WorkplaceType) }
}
