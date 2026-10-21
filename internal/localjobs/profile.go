package localjobs

import (
	"regexp"
	"strings"

	"jobsearcher/internal/domain"
)

type Profile struct {
	Title      string
	Category   string
	Experience string
	Education  string
	City       string
	State      string
	Country    string
}

var wordPattern = regexp.MustCompile(`[a-z0-9]+`)

func normalize(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a",
		"é", "e", "ê", "e", "í", "i", "ó", "o",
		"ô", "o", "õ", "o", "ú", "u", "ü", "u", "ç", "c",
	).Replace(value)
	return strings.Join(wordPattern.FindAllString(value, -1), " ")
}

// DefaultProfile returns the documented example profile. Callers should keep
// request values in memory and must not persist them automatically.
func DefaultProfile() Profile {
	return Profile{
		Title: "Vendedora", Category: "Vendas e comércio",
		Experience: "any", Education: "any",
		City: "Maceió", State: "AL", Country: "Brasil",
	}
}

// SearchTerms expands a local job title into discovery terms, not guaranteed
// equivalents. The requested title remains the primary matching signal.
func SearchTerms(p Profile) []string {
	terms := []string{p.Title}
	switch normalize(p.Title) {
	case "vendedora", "vendedor", "vendas":
		terms = append(terms, "vendedor", "consultora de vendas", "consultor de vendas", "atendente de loja", "vendedora de loja")
	case "recepcionista":
		terms = append(terms, "recepcionista de clinica", "recepcionista hospitalar", "atendente de clinica")
	case "atendente", "atendimento ao publico":
		terms = append(terms, "atendente de loja", "operador de atendimento", "recepcionista")
	}
	return unique(terms)
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := normalize(value)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, strings.TrimSpace(value))
	}
	return out
}

// Eligible accepts only clearly local, explicitly in-person jobs. Unknown
// location or workplace type is intentionally rejected.
func Eligible(j domain.Job, p Profile) bool {
	if !isMaceio(j.Location, p) || !isOnsite(j.WorkplaceType, j.Location, j.Description) {
		return false
	}
	return true
}

func isMaceio(location string, p Profile) bool {
	text := normalize(location)
	city := normalize(p.City)
	if city == "" {
		city = "maceio"
	}
	if !strings.Contains(text, city) {
		return false
	}
	// Avoid accepting a different city merely because Maceió is mentioned
	// in descriptive text; this check uses the structured location field only.
	if strings.Contains(text, "arapiraca") || strings.Contains(text, "recife") ||
		strings.Contains(text, "remote") || strings.Contains(text, "remoto") {
		return false
	}
	state := normalize(p.State)
	if state == "" {
		state = "al"
	}
	return strings.Contains(text, "alagoas") || regexp.MustCompile(`(^| )`+regexp.QuoteMeta(state)+`($| )`).MatchString(text)
}

func isOnsite(workplace, location, description string) bool {
	place := normalize(workplace)
	all := normalize(workplace + " " + location + " " + description)
	if strings.Contains(all, "hibrid") || strings.Contains(all, "hybrid") ||
		strings.Contains(all, "remote") || strings.Contains(all, "remoto") ||
		strings.Contains(all, "home office") {
		return false
	}
	if strings.Contains(place, "presencial") || strings.Contains(place, "on site") ||
		strings.Contains(place, "onsite") || strings.Contains(place, "in person") {
		return true
	}
	return false
}

// MatchScore is a simple non-technical score for local jobs. It does not use
// programming skills or the technology module's weighted scoring model.
func MatchScore(j domain.Job, p Profile) int {
	title := normalize(j.Title)
	requested := normalize(p.Title)
	if title == "" || requested == "" {
		return 0
	}
	if title == requested {
		return 100
	}
	if strings.Contains(title, requested) || strings.Contains(requested, title) {
		return 90
	}
	for _, term := range SearchTerms(p) {
		n := normalize(term)
		if n != requested && (strings.Contains(title, n) || strings.Contains(n, title)) {
			return 75
		}
	}
	category := normalize(p.Category)
	if category != "" && strings.Contains(normalize(j.Title+" "+j.Description+" "+j.Requirements), category) {
		return 60
	}
	return 35
}
