package filter

import (
	"jobsearcher/internal/domain"
	"regexp"
	"strings"
)

var words = regexp.MustCompile(`[a-z0-9]+`)
var brazilStateCode = regexp.MustCompile(`(^|[ ,;])-?(ac|al|ap|am|ba|ce|df|es|go|ma|mt|ms|mg|pa|pb|pr|pe|pi|rj|rn|rs|ro|rr|sc|sp|se|to)([ ,;]|$)`)

func norm(s string) string {
	s = strings.ToLower(s)
	s = strings.NewReplacer("\u00e1", "a", "\u00e0", "a", "\u00e2", "a", "\u00e3", "a", "\u00e9", "e", "\u00ea", "e", "\u00ed", "i", "\u00f3", "o", "\u00f4", "o", "\u00f5", "o", "\u00fa", "u", "\u00e7", "c").Replace(s)
	return strings.Join(words.FindAllString(s, -1), " ")
}

func containsAny(s string, values ...string) bool {
	for _, v := range values {
		if strings.Contains(s, norm(v)) {
			return true
		}
	}
	return false
}

// ClassifyLocation accepts any explicitly remote role unless the posting contains
// a clear geographic restriction that excludes Brazil. Physical roles must be in
// Maceio/Alagoas. Ambiguous workplace descriptions remain uncertain.
func ClassifyLocation(j domain.Job) (string, string) {
	w, l := norm(j.WorkplaceType), norm(j.Location)
	d := norm(j.Description)
	combined := l + " " + d
	remote := containsAny(w, "remote", "remoto", "remota") || containsAny(l, "remote", "remoto", "remota") || containsAny(d, "remote work", "remote role", "remote position", "fully remote", "work remotely", "trabalho remoto", "vaga remota", "vaga remoto")
	hybrid := containsAny(w, "hybrid", "hibrido", "híbrido") || containsAny(l, "hybrid", "hibrido", "híbrido") || containsAny(d, "hybrid work", "hybrid role", "hybrid position", "work hybrid", "trabalho hibrido", "trabalho híbrido", "vaga híbrida", "vaga hibrida")
	onsite := containsAny(w, "onsite", "on site", "in person", "presencial") || containsAny(d, "onsite", "on site", "in person", "office-based", "office based", "must work from the office", "presencial", "trabalho presencial", "modelo presencial")
	physicalMaceio := containsAny(combined, "maceio", "alagoas")
	outsideAlagoasState := brazilStateCode.MatchString(l) && !containsAny(l, " al ", "alagoas", "maceio")

	if remote && (hybrid || onsite) {
		return "uncertain_location", "posting contains conflicting remote and physical-workplace signals"
	}

	if remote {
		if containsAny(combined,
			"united states only", "us only", "usa only", "europe only", "uk only", "united kingdom only",
			"canada only", "worldwide except brazil", "worldwide excluding brazil", "excluding brazil",
			"except brazil", "must be located in the united states", "must reside in the united states",
			"must be based in the united states", "candidates must reside in the united states",
			"candidates must be located in the united states", "must be located in the us",
			"must reside in the us", "must be based in the us", "must reside in canada",
			"must be located in europe", "must reside in europe", "must be located in the uk",
			"must reside in the uk") {
			return "rejected_location", "remote role is explicitly restricted to an incompatible region"
		}
		if containsAny(combined, "brazil only", "brasil only", "brazil", "brasil", "latam", "latin america", "south america", "worldwide", "global", "americas") {
			return "approved", "remote scope explicitly includes Brazil or a broader eligible region"
		}
		// A specific remote location in an ineligible country is treated as a geographic restriction.
		// Keep generic "Remote" postings eligible because they do not establish a country restriction.
		if containsAny(l, "canada", "united states", "usa", "united kingdom", "uk", "europe", "australia", "new zealand", "singapore") {
			return "rejected_location", "remote posting is tied to a country or region outside the configured eligible scope"
		}
		return "approved", "remote work is stated and no incompatible geographic restriction was found"
	}

	if physicalMaceio && containsAny(combined, "brazil", "brasil", "alagoas", "maceio") {
		return "approved", "physical location is Maceio/Alagoas"
	}
	if outsideAlagoasState {
		return "rejected_location", "physical location is outside Maceio/Alagoas"
	}
	if hybrid || onsite || physicalMaceio {
		return "rejected_location", "onsite or hybrid physical location is not Maceio/Alagoas"
	}
	return "uncertain_location", "workplace type or physical location could not be determined"
}

func Evaluate(j domain.Job) domain.Job {
	decision, reason := ClassifyLocation(j)
	j.LocationReason = reason
	switch decision {
	case "approved":
		j.LocationEligible = domain.LocationEligible
	case "rejected_location":
		j.LocationEligible = domain.LocationRejected
	default:
		j.LocationEligible = domain.LocationUnknown
	}

	w, l, d := norm(j.WorkplaceType), norm(j.Location), norm(j.Description)
	remote := containsAny(w, "remote", "remoto", "remota") || containsAny(l, "remote", "remoto", "remota") || containsAny(d, "remote work", "remote role", "remote position", "fully remote", "work remotely", "trabalho remoto", "vaga remota", "vaga remoto")
	hybrid := containsAny(w, "hybrid", "hibrido", "híbrido") || containsAny(l, "hybrid", "hibrido", "híbrido") || containsAny(d, "hybrid work", "hybrid role", "hybrid position", "work hybrid", "trabalho hibrido", "trabalho híbrido", "vaga híbrida", "vaga hibrida")
	onsite := containsAny(w, "onsite", "on site", "in person", "presencial") || containsAny(d, "onsite", "on site", "in person", "office-based", "office based", "must work from the office", "presencial", "trabalho presencial", "modelo presencial")
	if remote == hybrid && !onsite {
		j.WorkplaceType = "unknown"
		return j
	}
	if onsite && (remote || hybrid) {
		j.WorkplaceType = "unknown"
		return j
	}
	switch {
	case remote:
		j.WorkplaceType = "remote"
	case hybrid:
		j.WorkplaceType = "hybrid"
	case onsite:
		j.WorkplaceType = "onsite"
	default:
		j.WorkplaceType = "unknown"
	}
	return j
}
