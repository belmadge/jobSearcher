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
	s = strings.NewReplacer("á", "a", "à", "a", "â", "a", "ã", "a", "é", "e", "ê", "e", "í", "i", "ó", "o", "ô", "o", "õ", "o", "ú", "u", "ç", "c").Replace(s)
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

// ClassifyLocation accepts only roles that are explicitly remote.
// Geographic restrictions are evaluated separately so a remote role tied
// exclusively to an incompatible country or region is rejected.
func ClassifyLocation(j domain.Job) (string, string) {
	w, l := norm(j.WorkplaceType), norm(j.Location)
	d := norm(j.Description)
	combined := l + " " + d

	remote := containsAny(w, "remote", "remoto", "remota") ||
		containsAny(l, "remote", "remoto", "remota") ||
		containsAny(d, "remote work", "remote role", "remote position", "fully remote", "work remotely", "trabalho remoto", "vaga remota", "vaga remoto")

	hybrid := containsAny(w, "hybrid", "hibrido", "híbrido") ||
		containsAny(l, "hybrid", "hibrido", "híbrido") ||
		containsAny(d, "hybrid work", "hybrid role", "hybrid position", "work hybrid", "trabalho hibrido", "trabalho híbrido", "vaga híbrida", "vaga hibrida")

	onsite := containsAny(w, "onsite", "on site", "in person", "presencial") ||
		containsAny(l, "onsite", "on site", "in person", "presencial") ||
		containsAny(d, "onsite", "on site", "in person", "office-based", "office based", "must work from the office", "presencial", "trabalho presencial", "modelo presencial")

	explicitRemote := containsAny(w, "remote", "remoto", "remota") || containsAny(l, "remote", "remoto", "remota")
	explicitHybrid := containsAny(w, "hybrid", "hibrido", "híbrido") || containsAny(l, "hybrid", "hibrido", "híbrido")
	explicitOnsite := containsAny(w, "onsite", "on site", "in person", "presencial") || containsAny(l, "onsite", "on site", "in person", "presencial")

	if explicitRemote {
		hybrid, onsite = false, false
	}
	if explicitHybrid {
		remote, onsite = false, false
	}
	if explicitOnsite {
		remote, hybrid = false, false
	}

	if remote && (hybrid || onsite) {
		return "uncertain_location", "posting contains conflicting remote and physical-workplace signals"
	}

	if hybrid {
		return "rejected_location", "hybrid roles are outside the global remote-only search"
	}
	if onsite {
		return "rejected_location", "onsite roles are outside the global remote-only search"
	}
	if !remote {
		return "uncertain_location", "workplace type could not be confirmed as remote"
	}

	if containsAny(combined,
		"united states only", "us only", "usa only", "europe only", "uk only", "united kingdom only",
		"canada only", "australia only", "new zealand only", "singapore only", "india only",
		"worldwide except brazil", "worldwide excluding brazil", "excluding brazil", "except brazil",
		"must be located in the united states", "must reside in the united states",
		"must be based in the united states", "candidates must reside in the united states",
		"candidates must be located in the united states", "must be located in the us",
		"must reside in the us", "must be based in the us", "must reside in canada",
		"must be located in canada", "must reside in europe", "must be located in europe",
		"must reside in the uk", "must be located in the uk") {
		return "rejected_location", "remote role is explicitly restricted to an incompatible region"
	}

	if containsAny(combined,
		"brazil only", "brasil only", "brazil", "brasil", "latam", "latin america",
		"south america", "worldwide", "global", "americas") {
		return "approved", "remote scope explicitly includes Brazil or a broader eligible region"
	}

	// A country/region in the location field is treated as a geographic restriction.
	if containsAny(l,
		"canada", "united states", "usa", "united kingdom", "uk", "europe",
		"australia", "new zealand", "singapore", "india", "indonesia", "philippines",
		"malaysia", "japan", "china", "hong kong", "taiwan", "south korea", "germany",
		"france", "spain", "italy", "netherlands", "belgium", "ireland", "portugal",
		"poland", "sweden", "norway", "denmark", "switzerland", "israel",
		"south africa", "nigeria", "emea", "apac") {
		return "rejected_location", "remote posting is tied to a country or region outside the global Brazil-compatible scope"
	}

	if strings.Contains(l, "remote") || strings.Contains(l, "remoto") || strings.Contains(w, "remote") || strings.Contains(w, "remoto") {
		return "approved", "remote work is explicitly stated without an incompatible country restriction"
	}

	return "uncertain_location", "remote work is stated but the geographic scope could not be determined"
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
	remote := containsAny(w, "remote", "remoto", "remota") ||
		containsAny(l, "remote", "remoto", "remota") ||
		containsAny(d, "remote work", "remote role", "remote position", "fully remote", "work remotely", "trabalho remoto", "vaga remota", "vaga remoto")
	hybrid := containsAny(w, "hybrid", "hibrido", "híbrido") ||
		containsAny(l, "hybrid", "hibrido", "híbrido") ||
		containsAny(d, "hybrid work", "hybrid role", "hybrid position", "work hybrid", "trabalho hibrido", "trabalho híbrido", "vaga híbrida", "vaga hibrida")
	onsite := containsAny(w, "onsite", "on site", "in person", "presencial") ||
		containsAny(l, "onsite", "on site", "in person", "presencial") ||
		containsAny(d, "onsite", "on site", "in person", "office-based", "office based", "must work from the office", "presencial", "trabalho presencial", "modelo presencial")

	explicitRemote := containsAny(w, "remote", "remoto", "remota") || containsAny(l, "remote", "remoto", "remota")
	explicitHybrid := containsAny(w, "hybrid", "hibrido", "híbrido") || containsAny(l, "hybrid", "hibrido", "híbrido")
	explicitOnsite := containsAny(w, "onsite", "on site", "in person", "presencial") || containsAny(l, "onsite", "on site", "in person", "presencial")

	if explicitRemote {
		hybrid, onsite = false, false
	}
	if explicitHybrid {
		remote, onsite = false, false
	}
	if explicitOnsite {
		remote, hybrid = false, false
	}

	switch {
	case remote && !hybrid && !onsite:
		j.WorkplaceType = "remote"
	case hybrid && !remote && !onsite:
		j.WorkplaceType = "hybrid"
	case onsite && !remote && !hybrid:
		j.WorkplaceType = "onsite"
	default:
		j.WorkplaceType = "unknown"
	}
	return j
}
