package adzuna
import "testing"
func TestNormalize(t *testing.T){j:=normalize(posting{ID:"1",Title:"Backend Engineer",RedirectURL:"https://example.com/1",Company:struct{DisplayName string `json:"display_name"`}{DisplayName:"Acme"}});if j.Source!="adzuna"||j.Company!="Acme"||j.Title!="Backend Engineer"{t.Fatalf("unexpected job: %+v",j)}}
