package themuse
import "testing"
func TestNormalizeRemote(t *testing.T){p:=posting{ID:1,Name:"Backend Engineer",Company:struct{Name string `json:"name"`}{Name:"Acme"},Locations:[]struct{Name string `json:"name"`}{{Name:"Flexible / Remote"}},Refs:struct{LandingPage string `json:"landing_page"`}{LandingPage:"https://example.com/1"}};j:=normalize(p);if j.Source!="themuse"||j.WorkplaceType!="Remote"{t.Fatalf("unexpected job: %+v",j)}}
