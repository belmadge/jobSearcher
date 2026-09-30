package filter

import (
	"jobsearcher/internal/domain"
	"testing"
)

func TestClassifyLocation(t *testing.T) {
	cases := []struct{ name, work, loc, description, want string }{
		{"remote eligible broad", "Remote", "Worldwide", "", "approved"},
		{"remote Brazil", "Remote", "Remote - Brazil", "", "approved"},
		{"remote LATAM", "Remote", "Remote - LATAM", "", "approved"},
		{"US only", "Remote", "Remote - US only", "", "rejected_location"},
		{"Europe only", "Remote", "Remote - Europe only", "", "rejected_location"},
		{"onsite Maceio", "Onsite", "Maceio - AL, Brazil", "", "approved"},
		{"onsite accented Maceio", "Onsite", "Macei\u00f3 - AL, Brazil", "", "approved"},
		{"onsite SP", "Onsite", "S\u00e3o Paulo, Brazil", "", "rejected_location"},
		{"hybrid Maceio", "Hybrid", "Maceio, Alagoas", "", "approved"},
		{"hybrid Recife", "Hybrid", "Recife, Brazil", "", "rejected_location"},
		{"remote without geographic scope", "Remote", "Remote", "", "approved"},
		{"unknown", "", "Somewhere", "", "uncertain_location"},
		{"Maceio physical location", "", "Maceio, Brazil", "", "approved"},
		{"onsite Maceio stated in description", "", "Maceio, Brazil", "This is an on-site role", "approved"},
		{"remote Brazil only from description", "Remote", "Remote", "Remote - Brazil only", "approved"},
		{"US residence restriction in description", "Remote", "Remote - LATAM", "Must be located in the United States", "rejected_location"},
		{"LATAM residence in description", "Remote", "Remote", "Candidates must reside in LATAM", "approved"},
		{"timezone-only remote", "Remote", "Remote", "Remote within CET timezone", "approved"},
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
