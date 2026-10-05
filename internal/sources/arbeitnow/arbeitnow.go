package arbeitnow
import ("context";"encoding/json";"fmt";"io";"net/http";"regexp";"strings";"time";"jobsearcher/internal/domain";"jobsearcher/internal/sources")
const defaultURL="https://www.arbeitnow.com/api/job-board-api"
type Client struct{HTTPClient *http.Client;URL string}
func NewClient()*Client{return &Client{HTTPClient:&http.Client{Timeout:25*time.Second},URL:defaultURL}}
func(c *Client)Name()string{return "arbeitnow"}
type response struct{Data []posting `json:"data"`}
type posting struct{Slug string `json:"slug"`;Company string `json:"company_name"`;Title string `json:"title"`;Description string `json:"description"`;Remote bool `json:"remote"`;URL string `json:"url"`;Location string `json:"location"`;Tags []string `json:"tags"`;Created int64 `json:"created_at"`}
func(c *Client)FetchJobs(ctx context.Context,_ sources.Query)([]domain.Job,error){req,err:=http.NewRequestWithContext(ctx,http.MethodGet,c.URL,nil);if err!=nil{return nil,err};req.Header.Set("Accept","application/json");req.Header.Set("User-Agent","JobSearcher/1.0");resp,err:=c.HTTPClient.Do(req);if err!=nil{return nil,err};defer resp.Body.Close();body,err:=io.ReadAll(io.LimitReader(resp.Body,32<<20));if err!=nil{return nil,err};if resp.StatusCode<200||resp.StatusCode>=300{return nil,fmt.Errorf("arbeitnow: HTTP %d",resp.StatusCode)};var raw response;if err:=json.Unmarshal(body,&raw);err!=nil{return nil,err};out:=make([]domain.Job,0,len(raw.Data));for _,p:=range raw.Data{if p.Title==""||p.URL==""{continue};out=append(out,normalize(p))};return out,nil}
func normalize(p posting)domain.Job{tags:=strings.Join(p.Tags,", ");work:="";if p.Remote{work="Remote"};return domain.Job{ID:"arbeitnow:"+p.Slug,Source:"arbeitnow",Title:strings.TrimSpace(p.Title),Company:strings.TrimSpace(p.Company),URL:strings.TrimSpace(p.URL),Location:strings.TrimSpace(p.Location),WorkplaceType:work,Description:stripHTML(p.Description),Requirements:tags,PostedAt:time.Unix(p.Created,0).UTC().Format(time.RFC3339)}}
var tagRE=regexp.MustCompile("<[^>]+>");var spaceRE=regexp.MustCompile("\s+")
func stripHTML(s string)string{return strings.TrimSpace(spaceRE.ReplaceAllString(tagRE.ReplaceAllString(s," ")," "))}
var _ sources.JobSource=(*Client)(nil)
