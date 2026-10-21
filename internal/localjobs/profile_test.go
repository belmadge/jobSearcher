package localjobs

import (
	"testing"

	"jobsearcher/internal/domain"
)

func TestDefaultProfile(t *testing.T) {
	p := DefaultProfile()
	if p.Title != "Vendedora" || p.Category != "Vendas e comércio" ||
		p.Experience != "any" || p.Education != "any" ||
		p.City != "Maceió" || p.State != "AL" || p.Country != "Brasil" {
		t.Fatalf("unexpected defaults: %+v", p)
	}
}

func TestSearchTermsForSales(t *testing.T) {
	got := SearchTerms(DefaultProfile())
	if len(got) < 4 {
		t.Fatalf("expected sales aliases, got %v", got)
	}
}

func TestEligibleRequiresMaceioAndExplicitOnsite(t *testing.T) {
	p := DefaultProfile()
	cases := []struct {
		name string
		job  domain.Job
		want bool
	}{
		{"Maceió presencial", domain.Job{Location:"Maceió, AL, Brasil", WorkplaceType:"Presencial", Description:"Atendimento em loja"}, true},
		{"Maceio without accent", domain.Job{Location:"Maceio - Alagoas", WorkplaceType:"On-site"}, true},
		{"other city", domain.Job{Location:"Arapiraca, AL", WorkplaceType:"Presencial"}, false},
		{"unknown city", domain.Job{Location:"Alagoas", WorkplaceType:"Presencial"}, false},
		{"remote", domain.Job{Location:"Maceió, AL", WorkplaceType:"Remote"}, false},
		{"hybrid", domain.Job{Location:"Maceió, AL", WorkplaceType:"Presencial", Description:"Modelo híbrido"}, false},
		{"unknown mode", domain.Job{Location:"Maceió, AL"}, false},
		{"remote mentioned", domain.Job{Location:"Maceió, AL", WorkplaceType:"On-site", Description:"Não remoto; trabalho presencial"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Eligible(tc.job, p); got != tc.want {
				t.Fatalf("Eligible() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestMatchScoreUsesLocalJobTitle(t *testing.T) {
	p := DefaultProfile()
	exact := MatchScore(domain.Job{Title:"Vendedora"}, p)
	related := MatchScore(domain.Job{Title:"Consultora de vendas"}, p)
	if exact != 100 || related < 75 || related >= exact {
		t.Fatalf("unexpected scores exact=%d related=%d", exact, related)
	}
}
