package lever

import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "strings"
 "time"
 "jobsearcher/internal/domain"
 "jobsearcher/internal/sources"
)

const defaultBaseURL = "https://api.lever.co/v0/postings"
type Client struct { HTTPClient *http.Client; BaseURL string }
func NewClient() *Client { return &Client{HTTPClient:&http.Client{Timeout:20*time.Second},BaseURL:defaultBaseURL} }
func (c *Client) Name() string { return "lever" }

type posting struct {
 ID string `json:"id"`
 Text string `json:"text"`
 Categories struct { Location string `json:"location"`; Team string `json:"team"`; Commitment string `json:"commitment"` } `json:"categories"`
 DescriptionPlain string `json:"descriptionPlain"`
 WorkplaceType string `json:"workplaceType"`
 AdditionalPlain string `json:"additionalPlain"`
 HostedURL string `json:"hostedUrl"`
 ApplyURL string `json:"applyUrl"`
 CreatedAt int64 `json:"createdAt"`
 UpdatedAt int64 `json:"updatedAt"`
}

func (c *Client) FetchJobs(ctx context.Context,q sources.Query)([]domain.Job,error){
 if len(q.LeverSites)==0{return nil,fmt.Errorf("lever: no sites configured")}
 var jobs []domain.Job;var errs []string
 for _,site:=range q.LeverSites{
  site=strings.TrimSpace(site);if site==""{continue}
  req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.BaseURL+"/"+site+"?mode=json",nil);if err!=nil{errs=append(errs,site+": "+err.Error());continue}
  req.Header.Set("Accept","application/json");resp,err:=c.HTTPClient.Do(req);if err!=nil{errs=append(errs,site+": "+err.Error());continue}
  body,readErr:=io.ReadAll(io.LimitReader(resp.Body,16<<20));resp.Body.Close();if readErr!=nil{errs=append(errs,site+": "+readErr.Error());continue}
  if resp.StatusCode<200||resp.StatusCode>=300{errs=append(errs,fmt.Sprintf("%s: HTTP %d",site,resp.StatusCode));continue}
  var payload []posting;if err:=json.Unmarshal(body,&payload);err!=nil{errs=append(errs,site+": "+err.Error());continue}
  for _,p:=range payload{jobs=append(jobs,normalize(site,p))}
 }
 if len(errs)>0{return jobs,fmt.Errorf("%s",strings.Join(errs,"; "))};return jobs,nil
}
func normalize(site string,p posting)domain.Job{
 desc:=strings.TrimSpace(strings.Join([]string{p.DescriptionPlain,p.AdditionalPlain}," "))
 lower:=strings.ToLower(p.Categories.Location+" "+desc);workplace:=""
 switch strings.ToLower(strings.TrimSpace(p.WorkplaceType)) { case "remote": workplace="Remote"; case "hybrid": workplace="Hybrid"; case "on-site","onsite": workplace="Onsite" }
 if workplace=="" { switch{case strings.Contains(lower,"hybrid"):workplace="Hybrid";case strings.Contains(lower,"remote"):workplace="Remote";case strings.Contains(lower,"onsite"),strings.Contains(lower,"on-site"),strings.Contains(lower,"on site"):workplace="Onsite"} }
 return domain.Job{ID:"lever:"+site+":"+p.ID,Source:"lever",Title:p.Text,Company:site,URL:p.HostedURL,ApplyURL:p.ApplyURL,Location:p.Categories.Location,WorkplaceType:workplace,EmploymentType:p.Categories.Commitment,Description:desc,Requirements:p.Categories.Team,PostedAt:unixTime(p.CreatedAt),UpdatedAt:unixTime(p.UpdatedAt)}
}
func unixTime(ms int64)string{if ms<=0{return ""};return time.Unix(0,ms*int64(time.Millisecond)).UTC().Format(time.RFC3339)}
var _ sources.JobSource=(*Client)(nil)
