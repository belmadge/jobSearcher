package dedupe

import (
	"testing"
	"jobsearcher/internal/domain"
)

func TestCanonicalURLDropsTrackingParameters(t *testing.T) {
	got := CanonicalURL("HTTPS://WWW.Example.com:443/jobs/123/?utm_source=mail&ref=abc&gh_jid=123#apply")
	want := "https://example.com/jobs/123?gh_jid=123"
	if got != want { t.Fatalf("got %q, want %q", got, want) }
}

func TestFallbackKeyNormalizesCompanyTitleAndLocation(t *testing.T) {
	a := FallbackKey(domain.Job{Company: "Empresa Ágil", Title: "Backend Engineer", Location: "Maceió - AL"})
	b := FallbackKey(domain.Job{Company: "Empresa Agil", Title: "Backend Engineer", Location: "Maceio AL"})
	if a != b { t.Fatalf("fallback keys differ: %q vs %q", a, b) }
}
