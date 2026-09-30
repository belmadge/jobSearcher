package remoteok

import (
 "context";"encoding/json";"fmt";"io";"net/http";"regexp";"strconv";"strings";"time"
 "jobsearcher/internal/domain";"jobsearcher/internal/sources"
)

const defaultURL="https://remoteok.com/api"
type Client struct{HTTPClient *http.Client; URL string}
func NewClient()*Client{return &Client{HTTPClient:&http.Client{Timeout:20*time.Second},URL:defaultURL}}
func (c *Client)Name()string{return "remoteok"}
type posting struct{ID string `json:"id"`;Company string `json:"company"`;Position string `json:"position"`;Tags []string `json:"tags"`;Description string `json:"description"`;Location string `json:"location"`;ApplyURL string `json:"apply_url"`;URL string `json:"url"`;SalaryMin float64 `json:"salary_min"`;SalaryMax float64 `json:"salary_max"`;Date string `json:"date"`}
func(c *Client)FetchJobs(ctx context.Context,q sources.Query)([]domain.Job,error){
 req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.URL,nil);if err!=nil{return nil,err};req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","JobSearcher/1.0")
 resp,err:=c.HTTPClient.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();body,err:=io.ReadAll(io.LimitReader(resp.Body,32<<20));if err!=nil{return nil,err};if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("remoteok: HTTP %d",resp.StatusCode)}
 var raw []json.RawMessage;if err=json.Unmarshal(body,&raw);err!=nil{return nil,err}
 jobs:=[]domain.Job{};for _,item:=range raw{var p posting;if json.Unmarshal(item,&p)!=nil||p.ID==""||p.Position==""{continue};jobs=append(jobs,normalize(p))}
 return jobs,nil
}
func normalize(p posting)domain.Job{desc:=stripHTML(p.Description);req:=strings.Join(p.Tags,", ");salary:="";if p.SalaryMin>0||p.SalaryMax>0{salary=strconv.FormatFloat(p.SalaryMin,'f',0,64)+" - "+strconv.FormatFloat(p.SalaryMax,'f',0,64)}
 lower:=strings.ToLower(p.Location+" "+desc);workplace:="Remote";if strings.Contains(lower,"hybrid"){workplace="Hybrid"}else if strings.Contains(lower,"onsite")||strings.Contains(lower,"on-site")||strings.Contains(lower,"on site"){workplace="Onsite"}
 return domain.Job{ID:"remoteok:"+p.ID,Source:"remoteok",Title:p.Position,Company:strings.TrimSpace(p.Company),URL:p.URL,ApplyURL:p.ApplyURL,Location:p.Location,WorkplaceType:workplace,Description:desc,Requirements:req,Salary:salary,PostedAt:p.Date}
}
var tagRE=regexp.MustCompile("<[^>]+>")
var spaceRE=regexp.MustCompile("\\s+")
func stripHTML(s string)string{return strings.TrimSpace(spaceRE.ReplaceAllString(tagRE.ReplaceAllString(s," "), " "))}
var _ sources.JobSource=(*Client)(nil)
