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
var yearsExperience = regexp.MustCompile(`\b(\d+)(?:\s*[-–]\s*(\d+))?\+?\s+years?\b`)
var techYearsExperience = regexp.MustCompile(`(?i)\b(\d+)(?:\s*[-–]\s*(\d+))?\+?\s+years?\s+(?:of\s+)?([a-zA-Z0-9+/.-]+)\b`)

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
		if len(match) > 2 && match[2] != "" { var maxYears int; fmt.Sscanf(match[2], "%d", &maxYears); if maxYears > years { years = maxYears } }
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
		if len(match) > 3 && match[2] != "" { var maxYears int; fmt.Sscanf(match[2], "%d", &maxYears); if maxYears > years { years = maxYears } }
		tech := canonical(match[3])
		if years > out[tech] { out[tech] = years }
	}
	return out
}

func profileTechnologyYears(p config.Profile, technology string) int {
	for key, years := range p.TechnologyYears { if canonical(key) == canonical(technology) { return years } }
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
	"red team", "red-team", "penetration tester", "penetration testing", "pentester", "cybersecurity", "cyber security", "security engineer", "security specialist", "security analyst", "head of security", "iam engineer",
	"qa engineer", "quality assurance", "sdet", "test engineer", "data scientist",
	"executivo de contas", "executiva de contas", "analista comercial", "analista de contas", "analista de operacoes", "analista de operações", "supervisor de operacoes", "supervisor de operações", "especialista de planejamento", "planejamento e performance", "marketplace", "farmer", "closing",
	"machine learning engineer", "ml engineer", "data engineer", "data analyst", "java developer", "java engineer", "kotlin developer", "kotlin engineer", "ruby developer", "ruby on rails", "django engineer", "python developer", "python engineer", ".net developer", "dotnet developer", "c# developer", "devsecops", "security", "iam", "appsec", "application security", "quality assurance", "sdet", "test engineer", "ios", "android", "ux ", "ui ",
}

func IsRelevant(j domain.Job, requestedRoles ...string) bool {
	title := strings.ToLower(strings.TrimSpace(j.Title))
	if len(requestedRoles) > 0 {
		// An explicit role search is authoritative. Evaluate the requested
		// role before generic exclusions so specialized roles can be searched
		// intentionally (for example Data Engineer or Java Developer).
		return matchesRequestedRole(j, requestedRoles)
	}
	if strings.Contains(title, "analista") || strings.Contains(title, "supervisor") || strings.Contains(title, "especialista") {
		technicalTitleSignals := []string{"sistemas", "software", "desenvolvedor", "desenvolvedora", "backend", "api", "dados", "data", "cloud", "infraestrutura", "infrastructure", "devops", "engenheiro", "engenheira", "programador", "programadora"}
		technical := false
		for _, signal := range technicalTitleSignals { if strings.Contains(title, signal) { technical = true; break } }
		if !technical { return false }
	}
	for _, term := range relevantTitleTerms {
		if !strings.Contains(title, term) { continue }
		if term == "software engineer" || term == "software developer" {
			if strings.Contains(title, "software engineer i") || strings.Contains(title, "software engineer ii") || strings.Contains(title, "software engineer 1") || strings.Contains(title, "software engineer 2") { return true }
			all := strings.ToLower(strings.Join([]string{j.Title, j.Description, j.Requirements}, " "))
			for _, signal := range []string{"backend", "back-end", "api", "service", "services", "microservice", "distributed systems", "platform", "cloud", "infrastructure", "integration", "go", "golang", "postgresql", "rest", "grpc", "message queue"} {
				if strings.Contains(all, signal) { return true }
			}
			return false
		}
		return true
	}
	all := strings.ToLower(strings.Join([]string{j.Title, j.Description, j.Requirements}, " "))
	for _, signal := range []string{"backend", "back-end", "api", "apis", "service", "services", "microservice", "distributed systems", "platform", "cloud", "infrastructure", "integration", "go", "golang", "postgresql", "rest", "grpc", "message queue"} {
		if strings.Contains(all, signal) { return true }
	}
	return false
}

func requestedRoleIncludesQA(requestedRoles []string) bool {
	for _, requested := range requestedRoles {
		role := normalizeRoleText(requested)
		if role == "qa" || strings.Contains(role, "quality assurance") || strings.Contains(role, "sdet") || strings.Contains(role, "test engineer") {
			return true
		}
	}
	return false
}

func isQATitleExclusion(term string) bool {
	switch term {
	case "qa engineer", "quality assurance", "sdet", "test engineer":
		return true
	default:
		return false
	}
}

func matchesRequestedRole(j domain.Job, requestedRoles []string) bool {
	title := normalizeRoleText(j.Title)
	for _, requested := range requestedRoles {
		role := normalizeRoleText(requested)
		if role == "" { continue }
		if role == "qa" || strings.Contains(role, "quality assurance") {
			if strings.Contains(title, " qa") || strings.HasPrefix(title, "qa ") || strings.Contains(title, "quality assurance") || strings.Contains(title, "sdet") || strings.Contains(title, "test engineer") { return true }
			continue
		}
		if role == "backend developer" || role == "backend engineer" || strings.Contains(role, "backend") {
			if matchesBackendRoleTitle(title) { return true }
			continue
		}
		if role == "software engineer" || role == "software developer" {
			if strings.Contains(title, "software engineer") || strings.Contains(title, "software developer") { return true }
			continue
		}
		if strings.Contains(role, "devops") || role == "sre" || strings.Contains(role, "site reliability") {
			if matchesDevOpsRoleTitle(title) { return true }
			continue
		}
		if strings.Contains(role, "data engineer") {
			if strings.Contains(title, "data engineer") || strings.Contains(title, "analytics engineer") || strings.Contains(title, "data platform engineer") { return true }
			continue
		}
		words := strings.Fields(role)
		matched := 0
		for _, word := range words {
			if word == "developer" { if strings.Contains(title, "developer") || strings.Contains(title, "engineer") { matched++; continue } }
			if word == "engineer" && (strings.Contains(title, "engineer") || strings.Contains(title, "developer")) { matched++; continue }
			if strings.Contains(title, word) { matched++ }
		}
		if matched == len(words) { return true }
	}
	return false
}

func matchesBackendRoleTitle(title string) bool {
	for _, candidate := range []string{"backend", "back end", "api engineer", "api developer", "server side developer", "server side engineer"} {
		if strings.Contains(title, candidate) { return true }
	}
	return false
}

func matchesDevOpsRoleTitle(title string) bool {
	for _, candidate := range []string{"devops", "site reliability", "sre", "platform engineer", "platform developer", "infrastructure engineer"} {
		if strings.Contains(title, candidate) { return true }
	}
	return false
}

func normalizeRoleText(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("-", " ", "_", " ", "/", " ", "(", " ", ")", " ", ",", " ").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

// Score applies the existing technology, responsibility and seniority scoring.
type MatchBucket string

const (
	StrongMatch      MatchBucket = "strong"
	CompatibleMatch  MatchBucket = "compatible"
	PossibleMatch    MatchBucket = "possible"
	LowMatch         MatchBucket = "low"
)

func matchBucket(score int) MatchBucket {
	switch {
	case score >= 75:
		return StrongMatch
	case score >= 60:
		return CompatibleMatch
	case score >= 40:
		return PossibleMatch
	default:
		return LowMatch
	}
}

func Score(j *domain.Job, p config.Profile) {
	ScoreForRoles(j, p, nil)
}

func ScoreForRoles(j *domain.Job, p config.Profile, requestedRoles []string) {
	roleMatch := roleMatchScore(j, requestedRoles)
	j.RoleMatch = roleMatch

	all := strings.Join([]string{j.Title, j.Description, j.Requirements}, " ")
	matched := []string{}
	for _, skill := range p.Technologies { if mentions(all, skill) { matched = appendUnique(matched, skill) } }
	required, nice := classifyRequirements(strings.Join([]string{j.Description, j.Requirements}, " "))
	for _, term := range required {
		inProfile := false
		for _, skill := range p.Technologies { if canonical(skill) == canonical(term) { inProfile = true; break } }
		if inProfile && mentions(all, term) { j.MustHaveMatch = appendUnique(j.MustHaveMatch, term) } else { j.MustHaveMissing = appendUnique(j.MustHaveMissing, term) }
	}
	for _, term := range nice { for _, skill := range p.Technologies { if canonical(skill) == canonical(term) && mentions(all, term) { j.NiceToHaveMatch = appendUnique(j.NiceToHaveMatch, term) } } }
	skillMatch := skillMatchScore(all, p.Technologies)
	j.SkillMatch = skillMatch
	j.RelatedSkillMatch = relatedSkillMatchScore(all, p.Technologies)
	baseTechnical := clamp(40+len(matched)*10-len(j.MustHaveMissing)*20, 0, 100)
	if mentions(all, "Go") { baseTechnical = clamp(baseTechnical+5, 0, 100) }
	if len(requestedRoles) == 0 {
		j.TechnicalMatch = baseTechnical
	} else {
		j.TechnicalMatch = blendRoleAndSkills(roleMatch, blendSkillScores(skillMatch, j.RelatedSkillMatch), baseTechnical)
	}
	j.ResponsibilityMatch = responsibilityMatchScore(all, requestedRoles, p.Technologies)
	cloudTerms := []string{"aws", "cloud", "kubernetes", "docker", "terraform", "infrastructure"}
	j.CloudMatch = categoryCoverage(all, cloudTerms)
	j.DomainMatch = neutralCoverage(all, p.Domains)
	j.LanguageMatch = neutralCoverage(all, []string{"english"})
	j.AIMatch = neutralCoverage(all, p.EmergingSkills)
	j.SeniorityMatch, j.SeniorityReason = inferSeniority(*j)
	if p.TargetSeniority != "" {
		j.SeniorityMatch, j.SeniorityReason = seniorityFitForTarget(*j, p.TargetSeniority)
	}
	j.Seniority = seniorityLabel(j.SeniorityMatch, j.SeniorityReason)
	w := p.Weights
	j.FitScore = clamp((j.TechnicalMatch*w.Technical+j.ResponsibilityMatch*w.Responsibility+j.SeniorityMatch*w.Seniority+j.CloudMatch*w.Cloud+j.DomainMatch*w.Domain+j.LanguageMatch*w.Language+j.AIMatch*w.AI)/100, 0, 100)
	if j.SeniorityMatch <= 10 && j.FitScore > 59 { j.FitScore = 59 }
	if j.SeniorityMatch == 55 && j.FitScore > 74 { j.FitScore = 74 }
	jobText := strings.Join([]string{j.Title, j.Description, j.Requirements}, " ")
	j.ExperienceMatch = experienceMatchScore(jobText, p)
	for _, years := range requiredYears(jobText) {
		if p.YearsExperience > 0 && years > p.YearsExperience { gap := fmt.Sprintf("%d+ years experience", years); j.MustHaveMissing = appendUnique(j.MustHaveMissing, gap); j.Reasons = append(j.Reasons, fmt.Sprintf("Experience gap: job asks %d+ years; profile has about %d years", years, p.YearsExperience)) }
	}
	for tech, years := range requiredTechnologyYears(jobText) {
		available := profileTechnologyYears(p, tech)
		if available > 0 && years > available { gap := fmt.Sprintf("%d+ years %s", years, tech); j.MustHaveMissing = appendUnique(j.MustHaveMissing, gap); j.Reasons = append(j.Reasons, fmt.Sprintf("Experience gap: job asks %d+ years %s; profile has about %d years", years, tech, available)) }
	}
	if j.ExperienceMatch < 100 && (len(requiredYears(jobText)) > 0 || len(requiredTechnologyYears(jobText)) > 0) {
		j.TechnicalMatch = clamp((j.TechnicalMatch*80+j.ExperienceMatch*20)/100, 0, 100)
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
	// Keep the bucket synchronized with the final FitScore after all penalties are applied.
	j.MatchBucket = string(matchBucket(j.FitScore))
	j.MatchHighlights = buildMatchHighlights(*j, matched)
	if len(j.NiceToHaveMatch) > 0 { j.Reasons = append(j.Reasons, "Preferred skills found: "+strings.Join(j.NiceToHaveMatch, ", ")) }
	if j.ResponsibilityMatch >= 40 { j.Reasons = append(j.Reasons, "Backend/API responsibilities are present") }
}


func responsibilityMatchScore(text string, requestedRoles, skills []string) int {
	terms := []string{}
	for _, role := range requestedRoles {
		r := normalizeRoleText(role)
		switch {
		case strings.Contains(r, "backend"), strings.Contains(r, "software"):
			terms = appendUnique(terms, "api"); terms = appendUnique(terms, "service"); terms = appendUnique(terms, "backend"); terms = appendUnique(terms, "microservice"); terms = appendUnique(terms, "distributed systems")
		case r == "qa" || strings.Contains(r, "quality assurance") || strings.Contains(r, "test"):
			terms = appendUnique(terms, "testing"); terms = appendUnique(terms, "automation"); terms = appendUnique(terms, "quality"); terms = appendUnique(terms, "test")
		case strings.Contains(r, "devops"), strings.Contains(r, "sre"), strings.Contains(r, "platform"):
			terms = appendUnique(terms, "deployment"); terms = appendUnique(terms, "infrastructure"); terms = appendUnique(terms, "ci/cd"); terms = appendUnique(terms, "monitoring"); terms = appendUnique(terms, "reliability")
		case strings.Contains(r, "data"):
			terms = appendUnique(terms, "data"); terms = appendUnique(terms, "pipeline"); terms = appendUnique(terms, "etl"); terms = appendUnique(terms, "warehouse")
		}
	}
	if len(terms) == 0 && len(skills) > 0 { terms = []string{"api", "service", "backend"} }
	if len(terms) == 0 { return 50 }
	return categoryCoverage(text, terms)
}

func seniorityFitForTarget(j domain.Job, target string) (int, string) {
	target = normalizeRoleText(target)
	jobScore, reason := inferSeniority(j)
	switch {
	case strings.Contains(target, "junior"), strings.Contains(target, "entry"):
		if jobScore >= 90 { return 100, "Junior/entry target; seniority is directly aligned" }
		if jobScore >= 80 { return 85, "Junior/entry target; mid-level role is compatible" }
		if jobScore == 55 { return 55, "Junior/entry target; senior role is above target" }
		return 20, "Junior/entry target; leadership/staff role is well above target"
	case strings.Contains(target, "mid"), strings.Contains(target, "pleno"), strings.Contains(target, "intermediate"):
		if jobScore >= 90 { return 90, "Mid-level target; junior/entry role is compatible" }
		if jobScore >= 80 { return 100, "Mid-level target; role is directly aligned" }
		if jobScore == 55 { return 75, "Mid-level target; senior role is above target" }
		return 30, "Mid-level target; leadership/staff role is above target"
	case strings.Contains(target, "senior"):
		if jobScore == 55 { return 100, "Senior target; role is directly aligned" }
		if jobScore == 78 { return 90, "Senior target; Engineer II role is compatible" }
		if jobScore >= 80 { return 80, "Senior target; role is below target" }
		return 70, reason
	case strings.Contains(target, "staff"), strings.Contains(target, "lead"):
		if jobScore <= 10 { return 100, "Staff/lead target; leadership level is directly aligned" }
		if jobScore == 55 { return 85, "Staff/lead target; senior role is below target" }
		return 60, "Staff/lead target; role is below target"
	}
	return jobScore, reason
}

func experienceMatchScore(text string, p config.Profile) int {
	years := requiredYears(text)
	techYears := requiredTechnologyYears(text)
	if len(years) == 0 && len(techYears) == 0 { return 70 }
	if p.YearsExperience <= 0 { return 50 }
	score := 100
	for _, required := range years {
		if required > p.YearsExperience { score = minScore(score, experienceGapScore(required-p.YearsExperience)) }
	}
	for tech, required := range techYears {
		available := profileTechnologyYears(p, tech)
		if available > 0 && required > available { score = minScore(score, experienceGapScore(required-available)) }
	}
	return score
}

func experienceGapScore(gap int) int {
	switch { case gap <= 0: return 100; case gap == 1: return 80; case gap == 2: return 60; case gap == 3: return 40; default: return 20 }
}

func minScore(a, b int) int { if a < b { return a }; return b }

func roleMatchScore(j *domain.Job, requestedRoles []string) int {
	if len(requestedRoles) == 0 { return 50 }
	title := normalizeRoleText(j.Title)
	best := 0
	for _, requested := range requestedRoles {
		role := normalizeRoleText(requested)
		if role == "" { continue }
		if title == role { if 100 > best { best = 100 }; continue }
		if strings.Contains(title, role) { if 95 > best { best = 95 }; continue }
		if rolesEquivalent(role, title) { if 85 > best { best = 85 }; continue }
		words := strings.Fields(role)
		if len(words) == 0 { continue }
		hits := 0
		for _, word := range words {
			if word == "developer" || word == "engineer" {
				if strings.Contains(title, "developer") || strings.Contains(title, "engineer") { hits++; continue }
			}
			if strings.Contains(title, word) { hits++ }
		}
		if hits > 0 {
			score := 35 + (65 * hits / len(words))
			if score > best { best = score }
		}
	}
	return best
}

func rolesEquivalent(requested, title string) bool {
	if requested == "qa" || strings.Contains(requested, "quality assurance") {
		return strings.Contains(title, "qa") || strings.Contains(title, "quality assurance") || strings.Contains(title, "sdet") || strings.Contains(title, "test engineer")
	}
	if strings.Contains(requested, "backend") {
		return strings.Contains(title, "backend") || strings.Contains(title, "back end") || strings.Contains(title, "back-end") || strings.Contains(title, "api engineer") || strings.Contains(title, "api developer")
	}
	if requested == "software engineer" || requested == "software developer" {
		return strings.Contains(title, "software engineer") || strings.Contains(title, "software developer") || strings.Contains(title, "backend engineer") || strings.Contains(title, "backend developer")
	}
	return false
}


var relatedTechnologyGroups = [][]string{
	{"aws", "gcp", "azure"},
	{"postgresql", "mysql", "mariadb", "oracle"},
	{"docker", "kubernetes"},
	{"rabbitmq", "kafka", "nats", "activemq"},
	{"graphql", "rest", "grpc"},
	{"redis", "memcached"},
	{"terraform", "pulumi", "cloudformation", "ansible"},
	{"datadog", "new relic", "grafana", "prometheus"},
	{"go", "java", "python", "ruby", "c#"},
}

func relatedSkillMatchScore(text string, skills []string) int {
	if len(skills) == 0 { return 50 }
	relatedGroups := 0
	coveredGroups := 0
	for _, skill := range skills {
		canonicalSkill := canonical(skill)
		for _, group := range relatedTechnologyGroups {
			inGroup := false
			for _, member := range group {
				if canonical(member) == canonicalSkill { inGroup = true; break }
			}
			if !inGroup { continue }
			relatedGroups++
			foundRelated := false
			for _, member := range group {
				if canonical(member) == canonicalSkill { continue }
				if mentions(text, member) { foundRelated = true; break }
			}
			if foundRelated { coveredGroups++ }
			break
		}
	}
	if relatedGroups == 0 { return 50 }
	return coveredGroups * 100 / relatedGroups
}

func blendSkillScores(exact, related int) int {
	if related == 50 { return exact }
	return clamp((exact*80+related*20)/100, 0, 100)
}

func skillMatchScore(text string, skills []string) int {
	if len(skills) == 0 { return 50 }
	hits := 0
	for _, skill := range skills {
		if mentions(text, skill) { hits++ }
	}
	return hits * 100 / len(skills)
}

func blendRoleAndSkills(roleMatch, skillMatch, baseTechnical int) int {
	if roleMatch == 0 { roleMatch = 50 }
	blended := (roleMatch*40 + skillMatch*40 + baseTechnical*20) / 100
	return clamp(blended, 0, 100)
}

func inferSeniority(j domain.Job) (int, string) {
	level := strings.ToLower(strings.TrimSpace(j.Title + " " + j.Seniority))
	switch {
	case containsAny(level, "staff", "principal", "director", "manager", "team lead", " tech lead", "lead "): return 10, "Leadership/staff-level title; outside the target I/II range"
	case strings.Contains(level, "senior"): return 55, "Senior-level title; considered above the primary I/II target"
	case strings.Contains(level, "software engineer ii") || strings.Contains(level, "software engineer 2") || strings.Contains(level, "engineer ii") || strings.Contains(level, "engineer 2"): return 78, "Explicit Engineer II level; secondary target"
	case strings.Contains(level, "software engineer i") || strings.Contains(level, "software engineer 1") || strings.Contains(level, "engineer i") || strings.Contains(level, "engineer 1"): return 95, "Explicit Engineer I level; primary target"
	case containsAny(level, "junior", "entry level", "entry-level", "associate", "new grad", "graduate"): return 90, "Entry/junior/associate level; close to the primary target"
	case containsAny(level, "mid-level", "mid level", "intermediate", "pleno"): return 85, "Mid-level title; compatible with the target range"
	}
	if strings.Contains(level, "software engineer") || strings.Contains(level, "software developer") { return 85, "Generic Software Engineer title; backend-compatible and no higher level stated" }
	return 60, "Seniority not explicit"
}

func seniorityLabel(score int, reason string) string {
	lower := strings.ToLower(reason)
	switch {
	case strings.Contains(lower, "engineer ii"): return "Engineer II"
	case strings.Contains(lower, "engineer i"): return "Engineer I"
	case strings.Contains(lower, "senior-level"): return "Senior"
	case strings.Contains(lower, "leadership/staff-level"): return "Staff/Lead"
	case score >= 90: return "Junior/Associate"
	case score >= 80: return "Mid-level"
	default: return "Not explicit"
	}
}

func containsAny(text string, terms ...string) bool { for _, term := range terms { if strings.Contains(text, term) { return true } }; return false }
func termCoverage(text string, terms []string) int { if len(terms) == 0 { return 0 }; hits:=0; for _,term:=range terms {if mentions(text,term){hits++}}; return hits*100/len(terms) }
func categoryCoverage(text string, terms []string) int { if len(terms)==0{return 50}; coverage:=termCoverage(text,terms);if coverage==0{return 50};return 50+coverage/2 }
func neutralCoverage(text string, terms []string) int { if len(terms)==0{return 50};for _,term:=range terms{if mentions(text,term){return termCoverage(text,terms)}};return 50 }


func buildMatchHighlights(j domain.Job, matched []string) []string {
	highlights := []string{}
	if j.RoleMatch >= 85 {
		highlights = append(highlights, fmt.Sprintf("Cargo muito alinhado (%d%%)", j.RoleMatch))
	} else if j.RoleMatch >= 60 {
		highlights = append(highlights, fmt.Sprintf("Cargo compatível (%d%%)", j.RoleMatch))
	}
	if len(matched) > 0 {
		highlights = append(highlights, "Skills encontradas: "+strings.Join(matched, ", "))
	}
	if j.RelatedSkillMatch >= 50 && j.SkillMatch < 100 {
		highlights = append(highlights, fmt.Sprintf("Tecnologias relacionadas ajudam no match (%d%%)", j.RelatedSkillMatch))
	}
	if j.ResponsibilityMatch >= 75 {
		highlights = append(highlights, fmt.Sprintf("Responsabilidades bem alinhadas (%d%%)", j.ResponsibilityMatch))
	}
	if j.SeniorityMatch >= 85 {
		highlights = append(highlights, "Senioridade alinhada ao perfil")
	} else if j.SeniorityMatch < 60 {
		highlights = append(highlights, "Senioridade é um ponto de atenção")
	}
	if j.ExperienceMatch < 100 {
		highlights = append(highlights, fmt.Sprintf("Experiência é um ponto de atenção (%d%%)", j.ExperienceMatch))
	}
	if len(j.MustHaveMissing) > 0 {
		highlights = append(highlights, "Requisitos ausentes: "+strings.Join(j.MustHaveMissing, ", "))
	}
	if len(highlights) == 0 {
		highlights = append(highlights, "Compatibilidade baseada nos critérios disponíveis da vaga")
	}
	return highlights
}
