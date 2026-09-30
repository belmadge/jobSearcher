package lever

import (
 "context"
 "net/http"
 "net/http/httptest"
 "testing"

 "jobsearcher/internal/sources"
)

func TestFetchJobsNormalizesPublicPostings(t *testing.T) {
 srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path!="/acme" || r.URL.Query().Get("mode")!="json" { t.Fatalf("unexpected request: %s",r.URL.String()) }
  w.Header().Set("Content-Type","application/json")
  w.WriteHeader(http.StatusOK)
  _,_=w.Write([]byte(`[{"id":"p1","text":"Backend Engineer","categories":{"location":"Remote - Brazil","team":"Engineering","commitment":"Full-time"},"descriptionPlain":"Build Go APIs","additionalPlain":"Work with PostgreSQL","hostedUrl":"https://jobs.lever.co/acme/p1","applyUrl":"https://jobs.lever.co/acme/p1/apply","createdAt":1760000000000,"updatedAt":1760001000000}]`))
 }))
 defer srv.Close()
 c:=NewClient();c.HTTPClient=srv.Client();c.BaseURL=srv.URL
 jobs,err:=c.FetchJobs(context.Background(),sources.Query{LeverSites:[]string{"acme"}})
 if err!=nil{t.Fatal(err)}
 if len(jobs)!=1{t.Fatalf("got %d jobs",len(jobs))}
 if jobs[0].Source!="lever"||jobs[0].Title!="Backend Engineer"||jobs[0].Location!="Remote - Brazil"{t.Fatalf("unexpected job: %+v",jobs[0])}
 if jobs[0].ApplyURL==""||jobs[0].PostedAt==""{t.Fatalf("expected URLs and timestamp: %+v",jobs[0])}
}

func TestFetchJobsKeepsSuccessfulSitesOnPartialFailure(t *testing.T) {
 srv:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request){
  if r.URL.Path=="/bad" { http.Error(w,"boom",http.StatusInternalServerError);return }
  w.Header().Set("Content-Type","application/json");_,_=w.Write([]byte(`[{"id":"p2","text":"Go Developer","categories":{"location":"Remote - LATAM"}}]`))
 }))
 defer srv.Close()
 c:=NewClient();c.HTTPClient=srv.Client();c.BaseURL=srv.URL
 jobs,err:=c.FetchJobs(context.Background(),Query{LeverSites:[]string{"bad","good"}})
 if err==nil{t.Fatal("expected partial failure error")}
 if len(jobs)!=1||jobs[0].ID!="lever:good:p2"{t.Fatalf("unexpected jobs: %+v",jobs)}
}
