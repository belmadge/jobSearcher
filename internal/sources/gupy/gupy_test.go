package gupy
import ("testing";"encoding/json")
func TestNormalizeRemote(t *testing.T){p:=posting{ID:"1",Title:"Backend Developer",CareerPageName:"Acme",City:"São Paulo",State:"São Paulo",WorkplaceType:"remote",JobURL:"https://example.com/job/1",Description:"<p>Go</p>"};j:=normalize(p);if j.Source!="gupy"||j.WorkplaceType!="Remote"||j.Company!="Acme"||j.Title!="Backend Developer"{t.Fatalf("unexpected job: %+v",j)}}
func TestNormalizeCompanyObject(t *testing.T){p:=posting{ID:"2",Title:"Engineer",CareerPageName:"",Company:json.RawMessage(`{"name":"Acme"}`),JobURL:"https://example.com/job/2"};j:=normalize(p);if j.Company!="Acme"{t.Fatalf("expected Acme, got %q",j.Company)}}
