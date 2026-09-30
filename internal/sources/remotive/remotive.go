package remotive

import (
 "context";"encoding/json";"fmt";"io";"net/http";"regexp";"strings";"time"
 "jobsearcher/internal/domain";"jobsearcher/internal/sources"
)
const defaultURL="https://remotive.com/api/remote-jobs"
type Client struct{HTTPClient *http.Client;URL string}
func NewClient()*Client{return &Client{HTTPClient:&http.Client{Timeout:20*time.Second},URL:defaultURL}}
func(c *Client)Name()string{return "remotive"}
type response struct{Jobs []job `json:"jobs"`}
type job struct{ID int64 `json:"id"`;URL string `json:"url"`;Title string `json:"title"`;Company string `json:"company_name"`;Category string `json:"category"`;JobType string `json:"job_type"`;PublicationDate string `json:"publication_date"`;CandidateLocation string `json:"candidate_required_location"`;Salary string `json:"salary"`;Description string `json:"description"`}
func(c *Client)FetchJobs(ctx context.Context,q sources.Query)([]domain.Job,error){
 u:=c.URL+"?category=software-dev&limit=100";req,err:=http.NewRequestWithContext(ctx,http.MethodGet,u,nil);if err!=nil{return nil,err};req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","JobSearcher/1.0")
 resp,err:=c.HTTPClient.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();body,err:=io.ReadAll(io.LimitReader(resp.Body,32<<20));if err!=nil{return nil,err};if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("remotive: HTTP %d",resp.StatusCode)}
 var payload response;if err=json.Unmarshal(body,&payload);err!=nil{return nil,err};out:=make([]domain.Job,0,len(payload.Jobs));for _,p:=range payload.Jobs{out=append(out,normalize(p))};return out,nil
}
func normalize(p job)domain.Job{desc:=stripHTML(p.Description);return domain.Job{ID:fmt.Sprintf("remotive:%d",p.ID),Source:"remotive",Title:p.Title,Company:p.Company,URL:p.URL,Location:p.CandidateLocation,WorkplaceType:"Remote",EmploymentType:p.JobType,Description:desc,Salary:p.Salary,PostedAt:p.PublicationDate}}
var tagRE=regexp.MustCompile("<[^>]+>");var spaceRE=regexp.MustCompile("\\s+")
func stripHTML(s string)string{return strings.TrimSpace(spaceRE.ReplaceAllString(tagRE.ReplaceAllString(s," ")," "))}
var _ sources.JobSource=(*Client)(nil)
