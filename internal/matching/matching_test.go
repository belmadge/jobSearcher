package matching

import (
 "strings"
 "testing"
 "jobsearcher/internal/config"
 "jobsearcher/internal/domain"
)

func profile() config.Profile {
 return config.Profile{
  Technologies: []string{"Go","PostgreSQL","SQL","AWS","RabbitMQ","GraphQL","Docker","Kubernetes","Terraform","CI/CD","Datadog","Observability","APIs"},
  Domains: []string{"fintech","production systems","APIs","observability","troubleshooting","cloud infrastructure"},
  EmergingSkills: []string{"AI Agents","AI Skills"},
  YearsExperience: 5,
  TechnologyYears: map[string]int{"Go": 3},
  Weights: config.Weights{Technical:35,Responsibility:15,Seniority:25,Cloud:10,Domain:7,Language:4,AI:4},
 }
}

func TestScoreAndGaps(t *testing.T){
 p:=profile();j:=domain.Job{Title:"Backend Engineer",Seniority:"Mid-level",Description:"Go APIs using PostgreSQL and AWS"};Score(&j,p)
 if j.FitScore<=0{t.Fatal("expected positive fit score")}
 if j.SeniorityMatch!=85{t.Fatalf("expected compatible software engineer level, got %d",j.SeniorityMatch)}
}

func TestAliasesAreEquivalent(t *testing.T){
 p:=profile()
 cases:=[]domain.Job{
  {Title:"Golang Backend Engineer",Description:"Build REST APIs with Postgres and Amazon Web Services"},
  {Title:"Go Backend Engineer",Description:"Build REST APIs with PostgreSQL and AWS"},
 }
 for i,j:=range cases {
  Score(&j,p)
  if j.TechnicalMatch < 70 { t.Fatalf("case %d: expected alias-expanded technical match, got %d; reasons=%v", i, j.TechnicalMatch, j.Reasons) }
 }
}

func TestMissingMustHaveIsPenalizedButUnmentionedNiceToHaveIsNot(t *testing.T){
 p:=profile()
 base:=domain.Job{Title:"Backend Engineer",Description:"Go and PostgreSQL required. Kubernetes is a nice to have."}
 Score(&base,p)
 if len(base.MustHaveMissing)!=0{t.Fatalf("unexpected missing required skills: %v",base.MustHaveMissing)}
 baselineNoNice:=domain.Job{Title:"Backend Engineer",Description:"Go and PostgreSQL required."}
 Score(&baselineNoNice,p)
 if base.FitScore<baselineNoNice.FitScore{t.Fatalf("optional Kubernetes should not reduce score: with=%d without=%d",base.FitScore,baselineNoNice.FitScore)}
 missing:=domain.Job{Title:"Backend Engineer",Description:"Go required. Python required. Kubernetes is a nice to have."}
 Score(&missing,p)
 if len(missing.MustHaveMissing)!=1||missing.MustHaveMissing[0]!="Python"{t.Fatalf("expected Python as missing must-have, got %v",missing.MustHaveMissing)}
 if missing.FitScore>=base.FitScore{t.Fatalf("missing required skill should lower score: base=%d missing=%d",base.FitScore,missing.FitScore)}
}

func TestNiceToHaveMatchIsTracked(t *testing.T){
 p:=profile();j:=domain.Job{Title:"Backend Engineer",Description:"Go required. Terraform is preferred."};Score(&j,p)
 found:=false;for _,v:=range j.NiceToHaveMatch{if v=="Terraform"{found=true}}
 if !found{t.Fatalf("expected Terraform nice-to-have match, got %v",j.NiceToHaveMatch)}
}


func TestIsRelevantRole(t *testing.T) {
	allowed := []domain.Job{
		{Title: "Backend Go Engineer"},
		{Title: "Software Engineer I", Description: "Backend engineering with APIs and Go"},
		{Title: "Software Engineer II", Description: "Backend engineering with APIs and Go"},
		{Title: "Platform Engineer"},
		{Title: "Golang Developer"},
	}
	for _, j := range allowed {
		if !IsRelevant(j) { t.Fatalf("expected relevant role: %s", j.Title) }
	}
	rejected := []domain.Job{
		{Title: "Technical Product Manager"},
		{Title: "Lead Product Designer"},
		{Title: "Business Development Manager"},
		{Title: "Remote Office Assistant"},
	}
	for _, j := range rejected {
		if IsRelevant(j) { t.Fatalf("expected unrelated role: %s", j.Title) }
	}
}

func TestSeniorAndStaffGuardrails(t *testing.T) {
	p := profile()
	staff := domain.Job{Title: "Staff Backend Engineer", Description: "Go PostgreSQL AWS Kubernetes Docker Terraform APIs"}
	Score(&staff, p)
	if staff.FitScore >= 60 { t.Fatalf("staff role should not enter the shortlist: got %d", staff.FitScore) }
	senior := domain.Job{Title: "Senior Backend Engineer", Description: "Go PostgreSQL AWS Kubernetes Docker Terraform APIs"}
	Score(&senior, p)
	if senior.FitScore >= 75 { t.Fatalf("senior role should remain below high-compatibility threshold: got %d", senior.FitScore) }
}


func TestExperienceRequirements(t *testing.T) {
	p := profile()
	ok := domain.Job{Title: "Software Engineer II", Description: "Go required. 3+ years of Go experience. 5+ years of software engineering experience."}
	Score(&ok, p)
	if ok.SeniorityMatch != 78 { t.Fatalf("Software Engineer II should be secondary target: %d", ok.SeniorityMatch) }
	if len(ok.MustHaveMissing) != 0 { t.Fatalf("expected no experience gaps: %v", ok.MustHaveMissing) }

	gap := domain.Job{Title: "Senior Go Engineer", Description: "Go required. 5+ years of Go experience. 7+ years of software engineering experience."}
	Score(&gap, p)
	if len(gap.MustHaveMissing) == 0 { t.Fatal("expected experience gaps") }
	if gap.FitScore >= ok.FitScore { t.Fatalf("experience gap should lower fit: gap=%d ok=%d", gap.FitScore, ok.FitScore) }
}


func TestEngineerLevelPreference(t *testing.T) {
	p := profile()
	i := domain.Job{Title: "Software Engineer I", Description: "Backend APIs with Go"}
	Score(&i, p)
	if i.SeniorityMatch != 95 || !strings.Contains(i.SeniorityReason, "primary target") { t.Fatalf("Engineer I should be primary target: %d (%s)", i.SeniorityMatch, i.SeniorityReason) }

	second := domain.Job{Title: "Software Engineer II", Description: "Backend APIs with Go"}
	Score(&second, p)
	if second.SeniorityMatch != 78 { t.Fatalf("Engineer II should be secondary target: %d", second.SeniorityMatch) }
	if i.SeniorityMatch <= second.SeniorityMatch { t.Fatal("Engineer I should rank above Engineer II by seniority fit") }
	if i.FitScore <= second.FitScore { t.Fatalf("Engineer I should rank above Engineer II by fit: I=%d II=%d", i.FitScore, second.FitScore) }
}

func TestUnrelatedEngineeringRolesAreRejected(t *testing.T) {
	roles := []string{"Frontend Engineer", "Mobile Engineer", "QA Engineer", "Data Scientist", "Machine Learning Engineer", "Data Analyst"}
	for _, title := range roles {
		if IsRelevant(domain.Job{Title: title, Description: "Software engineer role"}) {
			t.Fatalf("expected unrelated role to be rejected: %s", title)
		}
	}
}

func TestGenericSoftwareEngineerNeedsBackendSignal(t *testing.T) {
	if IsRelevant(domain.Job{Title: "Software Engineer", Description: "Build iOS applications with Swift"}) {
		t.Fatal("generic software engineer without backend signal should be rejected")
	}
	if !IsRelevant(domain.Job{Title: "Software Engineer", Description: "Build backend services and APIs with Go"}) {
		t.Fatal("backend software engineer should be relevant")
	}
}


func TestSpecializedPlatformRolesAreRejected(t *testing.T) {
	roles := []domain.Job{
		{Title: "Senior Shopify Developer", Description: "Build Shopify stores, themes and integrations with APIs"},
		{Title: "Salesforce Developer", Description: "Develop Salesforce applications and integrations"},
		{Title: "WordPress Developer", Description: "Build WordPress sites and plugins"},
	}
	for _, j := range roles {
		if IsRelevant(j) {
			t.Fatalf("expected specialized platform role to be rejected: %s", j.Title)
		}
	}
}

func TestSpecializedSecurityRolesAreRejected(t *testing.T) {
	roles := []string{
		"Red Team Specialist",
		"Cybersecurity Engineer",
		"Penetration Tester",
		"Security Engineer",
	}
	for _, title := range roles {
		if IsRelevant(domain.Job{Title: title, Description: "Go, AWS, Kubernetes and APIs"}) {
			t.Fatalf("expected specialized security role to be rejected: %s", title)
		}
	}
}


func TestUnmentionedCategoriesAreNeutral(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Software Engineer I", Description: "Backend work with Go, PostgreSQL and AWS"}
	Score(&j, p)
	if j.DomainMatch != 50 { t.Fatalf("unmentioned domain should be neutral: %d", j.DomainMatch) }
	if j.LanguageMatch != 50 { t.Fatalf("unmentioned language should be neutral: %d", j.LanguageMatch) }
	if j.AIMatch != 50 { t.Fatalf("unmentioned AI should be neutral: %d", j.AIMatch) }
	if j.FitScore < 60 { t.Fatalf("compatible backend role should not be pushed below threshold by missing categories: %d", j.FitScore) }
}

func TestYearsRequirementDoesNotInventSeniority(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Software Engineer", Description: "Backend APIs with Go. 2+ years of experience."}
	Score(&j, p)
	if j.SeniorityMatch != 85 {
		t.Fatalf("years requirement must not invent Engineer I seniority: got %d (%s)", j.SeniorityMatch, j.SeniorityReason)
	}
	if j.Seniority == "Engineer I" {
		t.Fatal("generic role with years requirement must not be labeled Engineer I")
	}
}

func TestAdditionalFalsePositiveRolesAreRejected(t *testing.T) {
	for _, title := range []string{"Java Developer", "Security Specialist", "Security Analyst", "Head of Security", "IAM Engineer"} {
		if IsRelevant(domain.Job{Title: title, Description: "Go AWS Kubernetes backend APIs"}) {
			t.Fatalf("expected unrelated specialized role to be rejected: %s", title)
		}
	}
}


func TestRequestedRoleChangesResultSet(t *testing.T) {
	qa := domain.Job{Title: "QA Engineer", Description: "Quality assurance and automated testing"}
	backend := domain.Job{Title: "Backend Developer", Description: "Backend APIs with Go"}
	if IsRelevant(qa, "Backend Developer") { t.Fatal("QA must not match a Backend Developer search") }
	if IsRelevant(backend, "QA") { t.Fatal("Backend Developer must not match a QA search") }
	if !IsRelevant(qa, "QA") { t.Fatal("QA must match a QA search") }
	if !IsRelevant(backend, "Backend Developer") { t.Fatal("Backend Developer must match a Backend Developer search") }
}

func TestRequestedRoleAliases(t *testing.T) {
	cases := []struct{ requested, title string }{
		{"QA", "Quality Assurance Engineer"},
		{"QA", "SDET"},
		{"Backend Developer", "Backend Engineer"},
		{"Backend Engineer", "Backend Developer"},
		{"Software Engineer", "Software Developer"},
	}
	for _, tc := range cases {
		if !IsRelevant(domain.Job{Title: tc.title}, tc.requested) {
			t.Fatalf("%q should match %q", tc.title, tc.requested)
		}
	}

	moreCases := []struct{ requested, title string }{
		{"Backend Developer", "API Engineer"},
		{"Backend Engineer", "Server-side Engineer"},
		{"DevOps Engineer", "Site Reliability Engineer"},
		{"SRE", "Platform Engineer"},
		{"Data Engineer", "Analytics Engineer"},
	}
	for _, tc := range moreCases {
		if !IsRelevant(domain.Job{Title: tc.title}, tc.requested) {
			t.Fatalf("%q should match %q", tc.title, tc.requested)
		}
	}
}





func TestExplicitRequestedRoleOverridesGenericExclusion(t *testing.T) {
	cases := []struct {
		requested string
		title     string
	}{
		{"Data Engineer", "Data Engineer"},
		{"Java Developer", "Java Developer"},
		{"QA", "QA Engineer"},
	}
	for _, tc := range cases {
		if !IsRelevant(domain.Job{Title: tc.title}, tc.requested) {
			t.Fatalf("%q should be searchable when explicitly requested as %q", tc.title, tc.requested)
		}
	}
}

func TestGenericSearchStillRejectsSpecializedRoles(t *testing.T) {
	for _, title := range []string{"Data Engineer", "Java Developer", "QA Engineer"} {
		if IsRelevant(domain.Job{Title: title}) {
			t.Fatalf("generic search should still reject specialized role: %s", title)
		}
	}
}
func TestRoleAndSkillMatchScores(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Go", "PostgreSQL", "AWS", "Docker"}
	exact := domain.Job{Title: "Backend Developer", Description: "Go PostgreSQL AWS Docker"}
	alias := domain.Job{Title: "API Engineer", Description: "Go PostgreSQL"}
	ScoreForRoles(&exact, p, []string{"Backend Developer"})
	ScoreForRoles(&alias, p, []string{"Backend Developer"})

	if exact.RoleMatch != 100 {
		t.Fatalf("exact role should score 100, got %d", exact.RoleMatch)
	}
	if alias.RoleMatch != 85 {
		t.Fatalf("backend alias role should score 85, got %d", alias.RoleMatch)
	}
	if exact.SkillMatch != 100 {
		t.Fatalf("all requested skills should score 100, got %d", exact.SkillMatch)
	}
	if alias.SkillMatch <= 0 || alias.SkillMatch >= exact.SkillMatch {
		t.Fatalf("partial skill overlap should fall below exact match: exact=%d alias=%d", exact.SkillMatch, alias.SkillMatch)
	}
	if exact.FitScore <= alias.FitScore {
		t.Fatalf("exact role/skill match should rank above partial match: exact=%d alias=%d", exact.FitScore, alias.FitScore)
	}
}

func TestRoleMatchDoesNotRequireExactTitle(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Backend Engineer", Description: "Go APIs with PostgreSQL"}
	ScoreForRoles(&j, p, []string{"Backend Developer"})
	if j.RoleMatch < 80 {
		t.Fatalf("equivalent backend title should score highly without exact title: %d", j.RoleMatch)
	}
	if j.FitScore < 55 {
		t.Fatalf("compatible role should remain above minimum fit: %d", j.FitScore)
	}
}


func TestResponsibilityMatchUsesRequestedRole(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Backend Developer", Description: "Build backend services, REST APIs and microservices"}
	ScoreForRoles(&j, p, []string{"Backend Developer"})
	if j.ResponsibilityMatch < 70 {
		t.Fatalf("backend responsibilities should score meaningfully: %d", j.ResponsibilityMatch)
	}
}

func TestExperienceMatchIsGraduated(t *testing.T) {
	p := profile()
	ok := domain.Job{Title: "Backend Engineer", Description: "Requires 5+ years of software engineering experience"}
	Score(&ok, p)
	if ok.ExperienceMatch != 100 {
		t.Fatalf("matching experience should score 100, got %d", ok.ExperienceMatch)
	}
	gap := domain.Job{Title: "Backend Engineer", Description: "Requires 7+ years of software engineering experience"}
	Score(&gap, p)
	if gap.ExperienceMatch != 60 {
		t.Fatalf("two-year experience gap should score 60, got %d", gap.ExperienceMatch)
	}
	if gap.ExperienceMatch >= ok.ExperienceMatch {
		t.Fatal("experience gap should reduce experience match")
	}
}

func TestTargetSeniorityChangesPreference(t *testing.T) {
	p := profile()
	p.TargetSeniority = "senior"
	j := domain.Job{Title: "Senior Backend Engineer", Description: "Backend APIs with Go"}
	Score(&j, p, )
	if j.SeniorityMatch != 100 {
		t.Fatalf("senior target should align with senior role: %d", j.SeniorityMatch)
	}

	p.TargetSeniority = "junior"
	Score(&j, p)
	if j.SeniorityMatch != 55 {
		t.Fatalf("junior target should score senior role below target: %d", j.SeniorityMatch)
	}
}


func TestRelatedTechnologyMatch(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Docker", "PostgreSQL", "AWS"}
	j := domain.Job{Title: "Backend Engineer", Description: "Build services with Kubernetes and MySQL"}
	Score(&j, p)
	if j.SkillMatch != 0 {
		t.Fatalf("unmentioned exact skills should not receive exact credit: %d", j.SkillMatch)
	}
	if j.RelatedSkillMatch <= 0 {
		t.Fatalf("related technologies should provide secondary credit: %d", j.RelatedSkillMatch)
	}
	if j.RelatedSkillMatch >= 100 {
		t.Fatalf("related technologies should not equal full exact skill match: %d", j.RelatedSkillMatch)
	}
}

func TestExactSkillRemainsStrongerThanRelated(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Docker"}
	exact := domain.Job{Title: "Backend Engineer", Description: "Docker"}
	related := domain.Job{Title: "Backend Engineer", Description: "Kubernetes"}
	Score(&exact, p)
	Score(&related, p)
	if exact.SkillMatch != 100 {
		t.Fatalf("exact Docker match should be 100: %d", exact.SkillMatch)
	}
	if related.RelatedSkillMatch != 100 {
		t.Fatalf("Kubernetes should be related to Docker: %d", related.RelatedSkillMatch)
	}
	if exact.TechnicalMatch <= related.TechnicalMatch {
		t.Fatalf("exact technology should remain stronger: exact=%d related=%d", exact.TechnicalMatch, related.TechnicalMatch)
	}
}


func TestMatchBucketUsesFinalFitScore(t *testing.T) {
	cases := []struct {
		score int
		want MatchBucket
	}{
		{75, StrongMatch},
		{90, StrongMatch},
		{74, CompatibleMatch},
		{60, CompatibleMatch},
		{59, PossibleMatch},
		{40, PossibleMatch},
		{39, LowMatch},
		{0, LowMatch},
	}
	for _, tc := range cases {
		if got := matchBucket(tc.score); got != tc.want {
			t.Fatalf("score %d: expected %q, got %q", tc.score, tc.want, got)
		}
	}
}

func TestScoreSetsMatchBucket(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Backend Engineer", Description: "Go PostgreSQL AWS Docker APIs"}
	Score(&j, p)
	if j.MatchBucket == "" {
		t.Fatal("expected match bucket to be populated")
	}
	if string(matchBucket(j.FitScore)) != j.MatchBucket {
		t.Fatalf("bucket must reflect final fit score: score=%d bucket=%q", j.FitScore, j.MatchBucket)
	}
}


func TestBetterMatchUsesFitScoreFirst(t *testing.T) {
	high := domain.Job{FitScore: 80, RoleMatch: 50}
	low := domain.Job{FitScore: 79, RoleMatch: 100}
	if !BetterMatch(high, low) {
		t.Fatal("higher FitScore must remain the primary ranking signal")
	}
}

func TestBetterMatchBreaksFitScoreTiesByRoleAndSkills(t *testing.T) {
	role := domain.Job{FitScore: 75, RoleMatch: 95, SkillMatch: 40}
	skill := domain.Job{FitScore: 75, RoleMatch: 80, SkillMatch: 100}
	if !BetterMatch(role, skill) {
		t.Fatal("RoleMatch should break a FitScore tie before SkillMatch")
	}
}

func TestBetterMatchUsesDeterministicFallback(t *testing.T) {
	a := domain.Job{FitScore: 70, RoleMatch: 80, SkillMatch: 80, TechnicalMatch: 80, Title: "Backend Engineer", Company: "A"}
	b := domain.Job{FitScore: 70, RoleMatch: 80, SkillMatch: 80, TechnicalMatch: 80, Title: "Backend Engineer", Company: "B"}
	if !BetterMatch(a, b) {
		t.Fatal("company should provide a deterministic final tie-break")
	}
}


func TestMatchHighlightsExplainMainSignals(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Backend Engineer", Description: "Go APIs with PostgreSQL. 7+ years of experience."}
	ScoreForRoles(&j, p, []string{"Backend Developer"})
	if len(j.MatchHighlights) == 0 {
		t.Fatal("expected match highlights")
	}
	foundRole, foundSkill, foundExperience := false, false, false
	for _, h := range j.MatchHighlights {
		if strings.Contains(h, "Cargo") { foundRole = true }
		if strings.Contains(h, "Skills") { foundSkill = true }
		if strings.Contains(h, "Experiência") { foundExperience = true }
	}
	if !foundRole || !foundSkill || !foundExperience {
		t.Fatalf("expected role, skill and experience explanations: %v", j.MatchHighlights)
	}
}

func TestMatchHighlightsReportMissingRequirements(t *testing.T) {
	p := profile()
	j := domain.Job{Title: "Backend Engineer", Description: "Go required. Python required."}
	Score(&j, p)
	found := false
	for _, h := range j.MatchHighlights {
		if strings.Contains(h, "Python") { found = true }
	}
	if !found {
		t.Fatalf("expected missing requirement in highlights: %v", j.MatchHighlights)
	}
}

func TestRankingPrefersCompleteRoleAndSkillMatch(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Go", "PostgreSQL", "AWS", "Docker"}

	complete := domain.Job{
		Title:       "Backend Engineer",
		Description: "Build backend services and APIs with Go, PostgreSQL, AWS and Docker.",
	}
	partial := domain.Job{
		Title:       "Backend Engineer",
		Description: "Build backend services and APIs with Go. Kubernetes is used instead of Docker.",
	}

	ScoreForRoles(&complete, p, []string{"Backend Developer"})
	ScoreForRoles(&partial, p, []string{"Backend Developer"})

	if !BetterMatch(complete, partial) {
		t.Fatalf("complete technical match should rank above partial match: complete=%d partial=%d", complete.FitScore, partial.FitScore)
	}
}

func TestRankingPenalizesMissingRequiredSkill(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Go", "PostgreSQL", "AWS"}

	complete := domain.Job{
		Title:       "Backend Engineer",
		Description: "Go, PostgreSQL and AWS are required.",
	}
	missing := domain.Job{
		Title:       "Backend Engineer",
		Description: "Go and PostgreSQL are required. Python is required.",
	}

	ScoreForRoles(&complete, p, []string{"Backend Developer"})
	ScoreForRoles(&missing, p, []string{"Backend Developer"})

	if len(missing.MustHaveMissing) == 0 {
		t.Fatal("expected missing required skill to be recorded")
	}
	if !BetterMatch(complete, missing) {
		t.Fatalf("job missing a required skill should not outrank a complete match: complete=%d missing=%d", complete.FitScore, missing.FitScore)
	}
}

func TestRankingKeepsExplicitSeniorityFitRelevant(t *testing.T) {
	p := profile()
	p.TargetSeniority = "senior"

	senior := domain.Job{
		Title:       "Senior Backend Engineer",
		Description: "Backend APIs with Go and PostgreSQL.",
	}
	junior := domain.Job{
		Title:       "Junior Backend Engineer",
		Description: "Backend APIs with Go and PostgreSQL.",
	}

	ScoreForRoles(&senior, p, []string{"Backend Engineer"})
	ScoreForRoles(&junior, p, []string{"Backend Engineer"})

	if senior.SeniorityMatch <= junior.SeniorityMatch {
		t.Fatalf("target seniority should prefer the aligned role: senior=%d junior=%d", senior.SeniorityMatch, junior.SeniorityMatch)
	}
	if !BetterMatch(senior, junior) {
		t.Fatalf("role aligned with requested seniority should rank above the less-aligned role: senior=%d junior=%d", senior.FitScore, junior.FitScore)
	}
}

func TestRankingDoesNotLetRelatedTechnologyBeatExactSkillMatch(t *testing.T) {
	p := profile()
	p.Technologies = []string{"Docker"}

	exact := domain.Job{
		Title:       "Backend Engineer",
		Description: "Backend services using Docker.",
	}
	related := domain.Job{
		Title:       "Backend Engineer",
		Description: "Backend services using Kubernetes.",
	}

	ScoreForRoles(&exact, p, []string{"Backend Developer"})
	ScoreForRoles(&related, p, []string{"Backend Developer"})

	if !BetterMatch(exact, related) {
		t.Fatalf("exact technology match should rank above related technology: exact=%d related=%d", exact.FitScore, related.FitScore)
	}
}
