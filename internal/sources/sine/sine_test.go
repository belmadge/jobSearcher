package sine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jobsearcher/internal/sources"
)

func TestParseListingsExtractsSINEVacancy(t *testing.T) {
	html := `<html><body><div>VENDEDOR - ANTARES</div><div>Ver mais...</div><div>#7756</div><div>Escolaridade mínima: Nível Médio completo</div><div>Bairro da vaga: Antares</div><div>Data limite para se candidatar: 31/12/2026</div></body></html>`
	jobs := ParseListings(html)
	if len(jobs) != 1 { t.Fatalf("got %d jobs, want 1", len(jobs)) }
	if !strings.Contains(strings.ToLower(jobs[0].Title), "vendedor") { t.Fatalf("unexpected title: %q", jobs[0].Title) }
	if jobs[0].Source != "sine-maceio" || jobs[0].WorkplaceType != "Presencial" { t.Fatalf("unexpected source/workplace: %#v", jobs[0]) }
	if !strings.Contains(jobs[0].URL, "/talentos/") { t.Fatalf("unexpected URL: %q", jobs[0].URL) }
}

func TestFetchJobsUsesContextAndFiltersTerms(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<div>VENDEDOR - ANTARES</div><div>#7756</div><div>Bairro da vaga: Antares</div><div>ATENDENTE - CENTRO</div><div>#7757</div><div>Bairro da vaga: Centro</div>`))
	}))
	defer server.Close()
	client := NewClient()
	client.HTTPClient = server.Client()
	client.URL = server.URL
	jobs, err := client.FetchJobs(context.Background(), sources.Query{Terms: []string{"vendedora"}})
	if err != nil { t.Fatal(err) }
	if len(jobs) != 1 || !strings.Contains(strings.ToLower(jobs[0].Title), "vendedor") { t.Fatalf("unexpected jobs: %#v", jobs) }
}
