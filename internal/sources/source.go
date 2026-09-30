package sources

import ("context"; "jobsearcher/internal/domain")
type Query struct { Terms []string; Location string; BoardTokens []string; LeverSites []string }
type Board struct { Token string; Company string; Enabled bool }
type BoardProvider interface { Boards(context.Context) ([]Board,error) }
type StaticBoardProvider struct { Configured []Board }
func (p StaticBoardProvider) Boards(ctx context.Context)([]Board,error){if err:=ctx.Err();err!=nil{return nil,err};out:=make([]Board,0,len(p.Configured));for _,b:=range p.Configured{if b.Enabled&&b.Token!=""{out=append(out,b)}};return out,nil}
type JobSource interface{Name()string;FetchJobs(context.Context,Query)([]domain.Job,error)}
type Mock struct{}
func (Mock)Name()string{return "mock"}
func (Mock)FetchJobs(_ context.Context,_ Query)([]domain.Job,error){return []domain.Job{{ID:"1",Source:"mock",Title:"Backend Engineer",Company:"Northstar",URL:"https://example.invalid/jobs/1",Location:"Remote - LATAM",WorkplaceType:"Remote",Seniority:"Mid-level",Description:"Build Go backend APIs using PostgreSQL, AWS, Docker and observability in production.",Requirements:"Go, PostgreSQL, AWS"},{ID:"2",Source:"mock",Title:"Senior Software Engineer",Company:"Example",URL:"https://example.invalid/jobs/2",Location:"Remote - US only",WorkplaceType:"Remote",Description:"Go services and APIs"},{ID:"3",Source:"mock",Title:"API Developer",Company:"Maceió Tech",URL:"https://example.invalid/jobs/3",Location:"Maceió - AL, Brazil",WorkplaceType:"Hybrid",Description:"Develop and support production APIs with Go and PostgreSQL."},{ID:"4",Source:"mock",Title:"Platform Engineer",Company:"Mystery Co",URL:"https://example.invalid/jobs/4",Location:"Remote",WorkplaceType:"Remote",Description:"Cloud infrastructure and Kubernetes"}},nil}
