package gupy

import (
 "context"
 "encoding/json"
 "fmt"
 "io"
 "net/http"
 "net/url"
 "strings"
 "time"
 "jobsearcher/internal/domain"
 "jobsearcher/internal/sources"
)

const defaultURL = "https://employability-portal.gupy.io/api/v1/jobs"
type Client struct { HTTPClient *http.Client; URL string }
func NewClient() *Client { return &Client{HTTPClient:&http.Client{Timeout:30*time.Second},URL:defaultURL} }
func (c *Client) Name() string { return "gupy" }
type response struct { Data []posting `json:"data"`; Jobs []posting `json:"jobs"` }
type posting struct { ID string `json:"id"`; Title string `json:"title"`; Name string `json:"name"`; Company json.RawMessage `json:"company"`; CareerPageName string `json:"careerPageName"`; City string `json:"city"`; State string `json:"state"`; AddressCountry string `json:"addressCountry"`; WorkplaceType string `json:"workplaceType"`; EmploymentType string `json:"employmentType"`; Description string `json:"description"`; Requirements json.RawMessage `json:"requirements"`; JobURL string `json:"jobUrl"`; ApplyURL string `json:"applyUrl"`; PublishedAt string `json:"publishedDate"`; PublishedAt2 string `json:"publishedAt"` }

func (c *Client) FetchJobs(ctx context.Context,q sources.Query)([]domain.Job,error) { terms:=uniqueTerms(q.Terms); if len(terms)==0 { terms=[]string{""} }; out:=[]domain.Job{}; for _,term:=range terms { jobs,err:=c.fetchTerm(ctx,term); if err!=nil{return out,fmt.Errorf("gupy term %q: %w",term,err)}; out=append(out,jobs...) }; return out,nil }
func (c *Client) fetchTerm(ctx context.Context,term string)([]domain.Job,error) { u,err:=url.Parse(c.URL);if err!=nil{return nil,err}; q:=u.Query();q.Set("jobName",term);q.Set("offset","0");q.Set("limit","100");u.RawQuery=q.Encode();req,err:=http.NewRequestWithContext(ctx,http.MethodGet,u.String(),nil);if err!=nil{return nil,err};req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","JobSearcher/1.0");resp,err:=c.HTTPClient.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();body,err:=io.ReadAll(io.LimitReader(resp.Body,32<<20));if err!=nil{return nil,err};if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("HTTP %d",resp.StatusCode)};var raw response;if err:=json.Unmarshal(body,&raw);err!=nil{return nil,err};postings:=raw.Data;if len(postings)==0{postings=raw.Jobs};out:=make([]domain.Job,0,len(postings));for _,p:=range postings{if p.ID==""{continue};out=append(out,normalize(p))};return out,nil }
func normalize(p posting)domain.Job{title:=strings.TrimSpace(p.Title);if title==""{title=strings.TrimSpace(p.Name)};company:=strings.TrimSpace(p.CareerPageName);if company==""{company=rawText(p.Company)};location:=strings.Trim(strings.TrimSpace(strings.Join([]string{p.City,p.State,p.AddressCountry},", ")),", ");u:=strings.TrimSpace(p.JobURL);if u==""{u=strings.TrimSpace(p.ApplyURL)};posted:=strings.TrimSpace(p.PublishedAt);if posted==""{posted=strings.TrimSpace(p.PublishedAt2)};return domain.Job{ID:"gupy:"+p.ID,Source:"gupy",Title:title,Company:company,URL:u,ApplyURL:strings.TrimSpace(p.ApplyURL),Location:location,WorkplaceType:normalizeWorkplace(p.WorkplaceType),EmploymentType:strings.TrimSpace(p.EmploymentType),Description:stripHTML(p.Description),Requirements:rawTextHTML(p.Requirements),PostedAt:posted}}
func normalizeWorkplace(v string)string{s:=strings.ToLower(strings.TrimSpace(v));switch s{case "remote","remoto":return "Remote";case "hybrid","hibrido","híbrido":return "Hybrid";case "on-site","onsite","on site","presencial":return "Onsite"};return ""}
func uniqueTerms(in []string)[]string{seen:=map[string]bool{};out:=[]string{};for _,v:=range in{s:=strings.TrimSpace(v);if s==""||seen[strings.ToLower(s)]{continue};seen[strings.ToLower(s)]=true;out=append(out,s)};return out}

func rawText(v json.RawMessage)string{if len(v)==0{return ""};var s string;if json.Unmarshal(v,&s)==nil{return strings.TrimSpace(s)};var obj struct{Name string `json:"name"`;DisplayName string `json:"display_name"`};if json.Unmarshal(v,&obj)==nil{if obj.Name!=""{return strings.TrimSpace(obj.Name)};return strings.TrimSpace(obj.DisplayName)};return ""}
func rawTextHTML(v json.RawMessage)string{return stripHTML(rawText(v))}
func stripHTML(s string)string{out:=make([]rune,0,len(s));tag:=false;for _,ch:=range s{if ch==60{tag=true;continue};if ch==62{tag=false;continue};if !tag{out=append(out,ch)}};return strings.Join(strings.Fields(string(out))," ")}
var _ sources.JobSource=(*Client)(nil)
