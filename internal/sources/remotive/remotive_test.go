package remotive
import "testing"
func TestNormalize(t *testing.T){j:=normalize(job{ID:42,Title:"Backend Engineer",Company:"Acme",URL:"https://remotive.com/remote-jobs/42",CandidateLocation:"Worldwide",JobType:"full_time",Description:"<p>Build <b>Go</b> APIs</p>",Salary:"$80k"});if j.Source!="remotive"||j.WorkplaceType!="Remote"||j.Location!="Worldwide"{t.Fatalf("unexpected job: %+v",j)};if j.Description!="Build Go APIs"{t.Fatalf("unexpected description %q",j.Description)}}
