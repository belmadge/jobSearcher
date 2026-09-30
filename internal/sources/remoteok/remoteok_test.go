package remoteok

import "testing"

func TestNormalize(t *testing.T){j:=normalize(posting{ID:"1",Company:"Acme",Position:"Backend Engineer",Tags:[]string{"golang","aws"},Description:"<p>Build <strong>Go</strong> APIs</p>",Location:"Remote",URL:"https://remoteok.com/jobs/1",SalaryMin:80000,SalaryMax:100000});if j.Source!="remoteok"||j.WorkplaceType!="Remote"||j.Company!="Acme"{t.Fatalf("unexpected job: %+v",j)};if j.Salary!="80000 - 100000"{t.Fatalf("unexpected salary %q",j.Salary)};if j.Description!="Build Go APIs"{t.Fatalf("unexpected description %q",j.Description)}}
