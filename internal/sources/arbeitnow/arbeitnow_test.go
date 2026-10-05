package arbeitnow
import "testing"
func TestNormalizeRemote(t *testing.T){j:=normalize(posting{Slug:"abc",Title:"Go Engineer",Company:"Acme",URL:"https://example.com/abc",Location:"Remote",Remote:true});if j.Source!="arbeitnow"||j.WorkplaceType!="Remote"{t.Fatalf("unexpected job: %+v",j)}}
