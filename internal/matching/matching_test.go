package matching

import (
 "testing"
 "jobsearcher/internal/config"
 "jobsearcher/internal/domain"
)

func profile() config.Profile {
 return config.Profile{
  Technologies: []string{"Go","PostgreSQL","SQL","AWS","RabbitMQ","GraphQL","Docker","Kubernetes","Terraform","CI/CD","Datadog","Observability","APIs"},
  Domains: []string{"fintech","production systems","APIs","observability","troubleshooting","cloud infrastructure"},
  EmergingSkills: []string{"AI Agents","AI Skills"},
  Weights: config.Weights{Technical:35,Responsibility:20,Seniority:15,Cloud:10,Domain:10,Language:5,AI:5},
 }
}

func TestScoreAndGaps(t *testing.T){
 p:=profile();j:=domain.Job{Title:"Backend Engineer",Seniority:"Mid-level",Description:"Go APIs using PostgreSQL and AWS"};Score(&j,p)
 if j.FitScore<=0{t.Fatal("expected positive fit score")}
 if j.SeniorityMatch!=85{t.Fatalf("expected compatible software engineer level, got %d",j.SeniorityMatch)}
}

func TestAliasesAreEquivalent(t *testing.T){
 p:=profile();j:=domain.Job{Title:"Golang Backend Engineer",Description:"Build REST APIs with Postgres and Amazon Web Services"}
 Score(&j,p)
 for _,want:=range []string{"Go","PostgreSQL","AWS"}{
  found:=false;for _,r:=range j.Reasons{if containsCue(r, want){found=true}}
  if !found{t.Fatalf("expected alias %s to contribute to match; reasons=%v",want,j.Reasons)}
 }
}

func TestMissingMustHaveIsPenalizedButUnmentionedNiceToHaveIsNot(t *testing.T){
 p:=profile()
 base:=domain.Job{Title:"Backend Engineer",Description:"Go and PostgreSQL required. Kubernetes is a nice to have."}
 Score(&base,p)
 if len(base.MustHaveMissing)!=0{t.Fatalf("unexpected missing required skills: %v",base.MustHaveMissing)}
 if base.FitScore<60{t.Fatalf("optional Kubernetes should not crush score, got %d",base.FitScore)}
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
