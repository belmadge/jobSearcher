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
  Weights: config.Weights{Technical:35,Responsibility:20,Seniority:15,Cloud:10,Domain:10,Language:5,AI:5},
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
	j := domain.Job{Title: "Software Engineer I", Description: "Backend APIs with Go, PostgreSQL and AWS"}
	Score(&j, p)
	if j.DomainMatch != 50 { t.Fatalf("unmentioned domain should be neutral: %d", j.DomainMatch) }
	if j.LanguageMatch != 50 { t.Fatalf("unmentioned language should be neutral: %d", j.LanguageMatch) }
	if j.AIMatch != 50 { t.Fatalf("unmentioned AI should be neutral: %d", j.AIMatch) }
	if j.FitScore < 60 { t.Fatalf("compatible backend role should not be pushed below threshold by missing categories: %d", j.FitScore) }
}
