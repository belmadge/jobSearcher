package filter

import (
	"jobsearcher/internal/domain"
	"testing"
)

func TestClassifyLocation(t *testing.T) {
	cases := []struct{ name, work, loc, description, want string }{
		{"remote eligible broad", "Remote", "Worldwide", "", "approved"},
		{"remote Brazil", "Remote", "Remote - Brazil", "", "approved"},
		{"remote Portuguese with office base", "", "Remoto; IFOOD - BASE OSASCO, Osasco - SP", "", "approved"},
		{"remote Portuguese Brazil with office base", "", "Remoto; IFOOD - BASE OSASCO, Osasco - SP", "Trabalho remoto no Brasil", "approved"},
		{"remote LATAM", "Remote", "Remote - LATAM", "", "approved"},
		{"hybrid Portuguese Maceio", "Híbrido", "Maceió, Alagoas", "", "rejected_location"},
		{"US only", "Remote", "Remote - US only", "", "rejected_location"},
		{"Europe only", "Remote", "Remote - Europe only", "", "rejected_location"},
		{"onsite Maceio", "Onsite", "Maceio - AL, Brazil", "", "rejected_location"},
		{"onsite accented Maceio", "Onsite", "Maceió - AL, Brazil", "", "rejected_location"},
		{"onsite SP", "Onsite", "São Paulo, Brazil", "", "rejected_location"},
		{"hybrid Maceio", "Hybrid", "Maceio, Alagoas", "", "rejected_location"},
		{"hybrid Recife", "Hybrid", "Recife, Brazil", "", "rejected_location"},
		{"remote without geographic scope", "Remote", "Remote", "", "approved"},
		{"remote Canada restriction", "Remote", "Canada", "", "rejected_location"},
		{"remote Europe restriction", "Remote", "Europe", "", "rejected_location"},
		{"unknown", "", "Somewhere", "", "uncertain_location"},
		{"Maceio physical location", "", "Maceio, Brazil", "", "uncertain_location"},
		{"onsite Maceio stated in description", "", "Maceio, Brazil", "This is an on-site role", "rejected_location"},
		{"remote Brazil only from description", "Remote", "Remote", "Remote - Brazil only", "approved"},
		{"US residence restriction in description", "Remote", "Remote - LATAM", "Must be located in the United States", "rejected_location"},
		{"LATAM residence in description", "Remote", "Remote", "Candidates must reside in LATAM", "approved"},
		{"timezone-only remote", "Remote", "Remote", "Remote within CET timezone", "approved"},
		{"remote EMEA", "Remote", "Remote - EMEA", "", "rejected_location"},
		{"remote APAC", "Remote", "Remote - APAC", "", "rejected_location"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := ClassifyLocation(domain.Job{WorkplaceType: tc.work, Location: tc.loc, Description: tc.description})
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestRemoteCloudDoesNotBecomeHybrid(t *testing.T) {
	j := domain.Job{WorkplaceType: "Remote", Location: "Brazil", Description: "Work with hybrid cloud infrastructure and remote teams."}
	got, _ := ClassifyLocation(j)
	if got != "approved" {
		t.Fatalf("expected remote cloud role to remain eligible, got %q", got)
	}
}

func TestEvaluateKeepsExplicitRemote(t *testing.T) {
	j := Evaluate(domain.Job{WorkplaceType: "Remote", Location: "Remote", Description: "Hybrid cloud infrastructure and remote collaboration."})
	if j.LocationEligible != domain.LocationEligible {
		t.Fatalf("expected eligible location, got %q (%s)", j.LocationEligible, j.LocationReason)
	}
	if j.WorkplaceType != "remote" {
		t.Fatalf("expected workplace type remote, got %q", j.WorkplaceType)
	}
}

func TestEvaluateRejectsHybridAndOnsiteEvenInBrazil(t *testing.T) {
	cases := []domain.Job{
		{WorkplaceType: "Hybrid", Location: "Maceió, Alagoas"},
		{WorkplaceType: "Onsite", Location: "Maceió, Alagoas"},
		{WorkplaceType: "", Location: "Maceió, Alagoas", Description: "This is an on-site role."},
	}
	for _, job := range cases {
		got := Evaluate(job)
		if got.LocationEligible != domain.LocationRejected {
			t.Fatalf("expected non-remote job to be rejected, got %q (%s)", got.LocationEligible, got.LocationReason)
		}
	}
}
