package matching

import (
	"fmt"
	"regexp"
	"strings"

	"jobsearcher/internal/config"
	"jobsearcher/internal/domain"
)

var techAliases = map[string][]string{
	"go": {"go", "golang"},
	"postgresql": {"postgresql", "postgres"},
	"ci/cd": {"ci/cd", "continuous integration", "continuous delivery", "continuous integration and continuous delivery"},
	"observability": {"observability", "monitoring", "tracing"},
	"aws": {"aws", "amazon web services"},
	"api": {"api", "apis", "backend api", "backend apis", "rest api", "rest apis"},
}

var requirementTechs = []string{"Go", "PostgreSQL", "SQL", "AWS", "GCP", "Azure", "RabbitMQ", "GraphQL", "Docker", "Kubernetes", "Terraform", "CI/CD", "Datadog", "Observability", "API", "Redis", "Python", "Java"}
var cueSplit = regexp.MustCompile(`[\n.;]+`)
var yearsExperience = regexp.MustCompile(`\b(\d+)(?:\+?\s+years?)\b`)
var techYearsExperience = regexp.MustCompile(`(?i)\b(\d+)\+?\s+years?\s+(?:of\s+)?([a-zA-Z0-9+/.-]+)\b`)

func canonical(term string) string {
	lower := strings.ToLower(strings.TrimSpace(term))
	for key, aliases := range techAliases { for _, alias := range aliases { if lower == alias { return key } } }
	return lower
}

func mentions(text, term string) bool {
	canonicalTerm := canonical(term)
	aliases := techAliases[canonicalTerm]
	if len(aliases) == 0 { aliases = []string{term} }
	text = strings.ToLower(text)
	for _, alias := range aliases {
		pattern := `(?i)(^|[^a-z0-9])` + regexp.QuoteMeta(strings.ToLower(alias)) + `($|[^a-z0-9])`
		if regexp.MustCompile(pattern).MatchString(text) { return true }
	}
	return false
}

func containsCue(sentence string, cues ...string) bool {
	s := strings.ToLower(sentence)
	for _, cue := range cues { if strings.Contains(s, cue) { return true } }
	return false
}

func classifyRequirements(text string) (must, nice []string) {
	for _, sentence := range cueSplit.Split(text, -1) {
		s := strings.ToLower(sentence)
		isMust := containsCue(s, "required", "must have", "must-have", "minimum", "strong experience") || (yearsExperience.MatchString(s) && strings.Contains(s, "experience"))
		isNice := containsCue(s, "nice to have", "nice-to-have", "preferred", "bonus", "plus")
		for _, term := range requirementTechs {
			if !mentions(sentence, term) { continue }
			if isMust { must = appendUnique(must, term) } else if isNice { nice = appendUnique(nice, term) }
		}
	}
	return must, nice
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items { if canonical(existing) == canonical(item) { return items } }
	return append(items, item)
}

func requiredYears(text string) []int {
	matches := yearsExperience.FindAllStringSubmatch(strings.ToLower(text), -1)
	out := []int{}
	for _, match := range matches {
		if len(match) < 2 { continue }
		var years int
		fmt.Sscanf(match[1], "%d", &years)
		if years > 0 { out = appendUniqueInt(out, years) }
	}
	return out
}

func appendUniqueInt(items []int, item int) []int {
	for _, existing := range items { if existing == item { return items } }
	return append(items, item)
}

func requiredTechnologyYears(text string) map[string]int {
	out := map[string]int{}
	for _, match := range techYearsExperience.FindAllStringSubmatch(text, -1) {
		if len(match) < 3 { continue }
		var years int
		fmt.Sscanf(match[1], "%d", &years)
		tech := canonical(match[2])
		if years > out[tech] { out[tech] = years }
	}
	return out
}

func profileTechnologyYears(p config.Profile, technology string) int {
	for key, years := range p.TechnologyYears {
		if canonical(key) == canonical(technology) { return years }
	}
	return 0
}

func clamp(v, low, high int) int { if v < low { return low }; if v > high { return high }; return v }

var relevantTitleTerms = []string{"backend", "software engineer", "software developer", "go developer", "golang", "api engineer", "platform engineer", "platform developer", "cloud engineer", "cloud infrastructure", "infrastructure engineer", "site reliability engineer", "sre", "reliability engineer", "distributed systems", "integration engineer", "integration developer"}
var irrelevantTitleTerms = []string{
	"product manager", "product lead", "product designer", "designer", "sales", "marketing",
	"recruiter", "human resources", "hr ", "payroll", "customer support", "support specialist",
	"business development", "account executive", "account manager", "finance manager", "legal",
	"copywriter", "writer", "technician", "surveyor", "data entry", "office assistant",
	"frontend", "front-end", "front end", "mobile", "ios", "android", "react native",
	"shopify", "salesforce developer", "wordpress developer", "magento", "drupal",
	"red team", "red-team", "penetration tester", "penetration testing", "pentester", "cybersecurity", "cyber security", "security engineer",
	"qa engineer", "quality assurance", "sdet", "test engineer", "data scientist",
	"machine learning engineer", "ml engineer", "data engineer", "data analyst", "ux ", "ui ",
}

func IsRelevant(j domain.Job) bool {
	title := strings.ToLower(strings.TrimSpace(j.Title))
	for _, term := range irrelevantTitleTerms {
		if strings.Contains(title, term) { return false }
	}
	for _, term := range relevantTitleTerms {
		if strings.Contains(title, term) {
			if strings.Contains(term, "software engineer") || strings.Contains(term, "software developer") {
			all := strings.ToLower(strings.Join([]string{j.Title, j.Description, j.Requirements}, " "))
				return mentions(all, "Go") || mentions(all, "Golang") || mentions(all, "Backend") || mentions(all, "API") || mentions(all, "PostgreSQL")
			}
			return true
		}
	}
	all := strings.ToLower(strings.Join([]string{j.Title, j.Description, j.Requirements}, " "))
	return mentions(all, "Go") || mentions(all, "Golang") || mentions(all, "Backend") || mentions(all, "API")
}

// Score uses explicit required language for penalties. Missing optional or unmentioned profile skills do not reduce the score.
func Score(j *domain.Job, p config.Profile) {
	all := strings.Join([]string{j.Title, j.Description, j.Requirements}, " ")
	matched := []string{}
	for _, skill := range p.Technologies { if mentions(all, skill) { matched = appendUnique(matched, skill) } }
	required, nice := classifyRequirements(strings.Join([]string{j.Description, j.Requirements}, " "))
	for _, term := range required {
		inProfile := false
		for _, skill := range p.Technologies { if canonical(skill) == canonical(term) { inProfile = true; break } }
		if inProfile && mentions(all, term) { j.MustHaveMatch = appendUnique(j.MustHaveMatch, term) } else { j.MustHaveMissing = appendUnique(j.MustHaveMissing, term) }
	}
	for _, term := range nice {
		for _, skill := range p.Technologies { if canonical(skill) == canonical(term) && mentions(all, term) { j.NiceToHaveMatch = appendUnique(j.NiceToHaveMatch, term) } }
	}
	j.TechnicalMatch = clamp(40+len(matched)*10-len(j.MustHaveMissing)*20, 0, 100)

	respTerms := []string{"backend", "api", "service", "distributed systems", "integration"}
	j.ResponsibilityMatch = termCoverage(all, respTerms)
	cloudTerms := []string{"aws", "cloud", "kubernetes", "docker", "terraform", "infrastructure"}
	j.CloudMatch = termCoverage(all, cloudTerms)
	j.DomainMatch = termCoverage(all, p.Domains)
	j.LanguageMatch = termCoverage(all, []string{"english"})
	j.AIMatch = termCoverage(all, p.EmergingSkills)
	j.SeniorityMatch, j.SeniorityReason = inferSeniority(*j)
	j.Seniority = seniorityLabel(j.SeniorityMatch, j.SeniorityReason)

	w := p.Weights
	j.FitScore = clamp((j.TechnicalMatch*w.Technical+j.ResponsibilityMatch*w.Responsibility+j.SeniorityMatch*w.Seniority+j.CloudMatch*w.Cloud+j.DomainMatch*w.Domain+j.LanguageMatch*w.Language+j.AIMatch*w.AI)/100, 0, 100)
	if j.SeniorityMatch <= 10 && j.FitScore > 59 { j.FitScore = 59 }
	if j.SeniorityMatch == 55 && j.FitScore > 74 { j.FitScore = 74 }
	jobText := strings.Join([]string{j.Title, j.Description, j.Requirements}, " ")
	for _, years := range requiredYears(jobText) {
		if p.YearsExperience > 0 && years > p.YearsExperience {
			gap := fmt.Sprintf("%d+ years experience", years)
			j.MustHaveMissing = appendUnique(j.MustHaveMissing, gap)
			j.Reasons = append(j.Reasons, fmt.Sprintf("Experience gap: job asks %d+ years; profile has about %d years", years, p.YearsExperience))
		}
	}
	for tech, years := range requiredTechnologyYears(jobText) {
		available := profileTechnologyYears(p, tech)
		if available > 0 && years > available {
			gap := fmt.Sprintf("%d+ years %s", years, tech)
			j.MustHaveMissing = appendUnique(j.MustHaveMissing, gap)
			j.Reasons = append(j.Reasons, fmt.Sprintf("Experience gap: job asks %d+ years %s; profile has about %d years", years, tech, available))
		}
	}
	if len(j.MustHaveMissing) > 0 {
		j.TechnicalMatch = clamp(j.TechnicalMatch-len(j.MustHaveMissing)*10, 0, 100)
		j.FitScore = clamp((j.TechnicalMatch*w.Technical+j.ResponsibilityMatch*w.Responsibility+j.SeniorityMatch*w.Seniority+j.CloudMatch*w.Cloud+j.DomainMatch*w.Domain+j.LanguageMatch*w.Language+j.AIMatch*w.AI)/100, 0, 100)
		if j.SeniorityMatch <= 10 && j.FitScore > 59 { j.FitScore = 59 }
		if j.SeniorityMatch == 55 && j.FitScore > 74 { j.FitScore = 74 }
	}
	j.Gaps = append([]string(nil), j.MustHaveMissing...)
	if len(matched) > 0 { j.Reasons = append(j.Reasons, "Profile technology overlap: "+strings.Join(matched, ", ")) }
	if len(j.MustHaveMatch) > 0 { j.Reasons = append(j.Reasons, "Required skills found: "+strings.Join(j.MustHaveMatch, ", ")) }
	if len(j.MustHaveMissing) > 0 { j.Reasons = append(j.Reasons, "Required skills missing: "+strings.Join(j.MustHaveMissing, ", ")) }
	if len(j.NiceToHaveMatch) > 0 { j.Reasons = append(j.Reasons, "Preferred skills found: "+strings.Join(j.NiceToHaveMatch, ", ")) }
	if j.ResponsibilityMatch >= 40 { j.Reasons = append(j.Reasons, "Backend/API responsibilities are present") }
}

func inferSeniority(j domain.Job) (int, string) {
	level := strings.ToLower(strings.TrimSpace(j.Title + " " + j.Seniority))
	text := strings.ToLower(strings.Join([]string{j.Title, j.Seniority, j.Description, j.Requirements}, " "))

	switch {
	case containsAny(level, "staff", "principal", "director", "manager", "team lead", " tech lead", "lead "):
		return 10, "Leadership/staff-level title; outside the target I/II range"
	case strings.Contains(level, "senior"):
		return 55, "Senior-level title; considered above the primary I/II target"
	case strings.Contains(level, "software engineer ii") || strings.Contains(level, "software engineer 2") || strings.Contains(level, "engineer ii") || strings.Contains(level, "engineer 2"):
		return 78, "Explicit Engineer II level; secondary target"
	case strings.Contains(level, "software engineer i") || strings.Contains(level, "software engineer 1") || strings.Contains(level, "engineer i") || strings.Contains(level, "engineer 1"):
		return 95, "Explicit Engineer I level; primary target"
	case containsAny(level, "junior", "entry level", "entry-level", "associate", "new grad", "graduate"):
		return 90, "Entry/junior/associate level; close to the primary target"
	case containsAny(level, "mid-level", "mid level", "intermediate", "pleno"):
		return 85, "Mid-level title; compatible with the target range"
	}

	for _, years := range requiredYears(text) {
		switch {
		case years <= 2:
			return 92, fmt.Sprintf("Description asks for %d+ years; consistent with Engineer I", years)
		case years <= 4:
			return 80, fmt.Sprintf("Description asks for %d+ years; consistent with Engineer II/early mid-level", years)
		case years <= 5:
			return 68, fmt.Sprintf("Description asks for %d+ years; seniority is above the primary I target", years)
		default:
			return 55, fmt.Sprintf("Description asks for %d+ years; treated as senior-level scope", years)
		}
	}
	if strings.Contains(level, "software engineer") || strings.Contains(level, "software developer") {
		return 85, "Generic Software Engineer title; backend-compatible and no higher level stated"
	}
	return 60, "Seniority not explicit"
}

func seniorityLabel(score int, reason string) string {
	lower := strings.ToLower(reason)
	switch {
	case strings.Contains(lower, "engineer ii level") || strings.Contains(lower, "engineer ii"):
		return "Engineer II"
	case strings.Contains(lower, "engineer i level") || strings.Contains(lower, "engineer i"):
		return "Engineer I"
	case strings.Contains(lower, "senior-level"), strings.Contains(lower, "senior-level title"):
		return "Senior"
	case strings.Contains(lower, "leadership/staff-level"):
		return "Staff/Lead"
	case score >= 90:
		return "Junior/Associate"
	case score >= 80:
		return "Mid-level"
	default:
		return "Not explicit"
	}
}

func containsAny(text string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}

func termCoverage(text string, terms []string) int {
	if len(terms) == 0 { return 0 }
	hits := 0
	for _, term := range terms { if mentions(text, term) { hits++ } }
	return hits * 100 / len(terms)
}
