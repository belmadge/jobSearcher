package himalayas

import (
	"context"
	"errors"
	"encoding/json"
	"bytes"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"jobsearcher/internal/domain"
	"jobsearcher/internal/sources"
)

const defaultURL = "https://himalayas.app/jobs/api/search"

type Client struct { HTTPClient *http.Client; URL string }

func NewClient() *Client { return &Client{HTTPClient: &http.Client{Timeout: 20 * time.Second}, URL: defaultURL} }
func (c *Client) Name() string { return "himalayas" }

type response struct { Jobs []job `json:"jobs"` }
type restriction struct { Name string `json:"name"`; Alpha2 string `json:"alpha2"` }

func (r *restriction) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) { return nil }
	if data[0] == '"' {
		return json.Unmarshal(data, &r.Name)
	}
	type alias restriction
	var v alias
	if err := json.Unmarshal(data, &v); err != nil { return err }
	*r = restriction(v)
	return nil
}
type job struct {
	GUID string `json:"guid"`
	Title string `json:"title"`
	Excerpt string `json:"excerpt"`
	CompanyName string `json:"companyName"`
	EmploymentType string `json:"employmentType"`
	MinSalary *float64 `json:"minSalary"`
	MaxSalary *float64 `json:"maxSalary"`
	SalaryPeriod string `json:"salaryPeriod"`
	Seniority []string `json:"seniority"`
	Currency string `json:"currency"`
	LocationRestrictions []restriction `json:"locationRestrictions"`
	Description string `json:"description"`
	PubDate int64 `json:"pubDate"`
	ExpiryDate int64 `json:"expiryDate"`
	ApplicationLink string `json:"applicationLink"`
}

var tagRE = regexp.MustCompile("(?s)<[^>]+>")
var spaceRE = regexp.MustCompile("\\s+")

func (c *Client) FetchJobs(ctx context.Context, q sources.Query) ([]domain.Job, error) {
	terms := uniqueTerms(q.Terms)
	if len(terms) == 0 { terms = []string{"software engineer", "backend engineer", "golang", "platform engineer"} }
	seen := map[string]bool{}
	out := make([]domain.Job, 0)
	var errs []error
	for _, term := range terms {
		u, _ := url.Parse(c.URL)
		params := u.Query()
		params.Set("q", term)
		params.Set("sort", "recent")
		params.Set("page", "1")
		u.RawQuery = params.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("term %q: %w", term, err))
			continue
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", "JobSearcher/1.0")
		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			errs = append(errs, fmt.Errorf("term %q: %w", term, err))
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
		resp.Body.Close()
		if readErr != nil {
			errs = append(errs, fmt.Errorf("term %q: %w", term, readErr))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			errs = append(errs, fmt.Errorf("term %q: HTTP %d", term, resp.StatusCode))
			continue
		}
		var payload response
		if err := json.Unmarshal(body, &payload); err != nil {
			errs = append(errs, fmt.Errorf("term %q: %w", term, err))
			continue
		}
		for _, p := range payload.Jobs {
			id := p.GUID
			if id == "" { id = p.ApplicationLink + "|" + p.Title + "|" + p.CompanyName }
			if seen[id] { continue }
			seen[id] = true
			out = append(out, normalize(p))
		}
	}
	return out, errors.Join(errs...)
}

func uniqueTerms(terms []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		key := strings.ToLower(term)
		if term == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, term)
	}
	return out
}

func normalize(p job) domain.Job {
	desc := stripHTML(p.Description)
	if desc == "" { desc = p.Excerpt }
	location := "Remote"
	if len(p.LocationRestrictions) > 0 {
		names := make([]string, 0, len(p.LocationRestrictions))
		for _, r := range p.LocationRestrictions {
			if r.Name != "" { names = append(names, r.Name) } else if r.Alpha2 != "" { names = append(names, r.Alpha2) }
		}
		if len(names) > 0 { location = "Remote - " + strings.Join(names, ", ") }
	}
	salary := ""
	if p.MinSalary != nil || p.MaxSalary != nil {
		parts := []string{}
		if p.MinSalary != nil { parts = append(parts, fmt.Sprintf("%.0f", *p.MinSalary)) }
		if p.MaxSalary != nil { parts = append(parts, fmt.Sprintf("%.0f", *p.MaxSalary)) }
		salary = p.Currency + " " + strings.Join(parts, " - ") + " / " + p.SalaryPeriod
	}
	posted := ""
	if p.PubDate > 0 {
		// Himalayas has historically returned Unix seconds for some records and
		// milliseconds for others. Detect the unit to avoid 1970-era dates.
		var postedAt time.Time
		if p.PubDate < 1_000_000_000_000 { postedAt = time.Unix(p.PubDate, 0).UTC() } else { postedAt = time.UnixMilli(p.PubDate).UTC() }
		posted = postedAt.Format(time.RFC3339)
	}
	return domain.Job{ID:"himalayas:"+p.GUID,Source:"himalayas",Title:p.Title,Company:p.CompanyName,URL:p.ApplicationLink,ApplyURL:p.ApplicationLink,CanonicalURL:p.ApplicationLink,Location:location,WorkplaceType:"Remote",EmploymentType:p.EmploymentType,Seniority:strings.Join(p.Seniority,", "),Description:desc,Salary:salary,PostedAt:posted}
}

func stripHTML(s string) string {
	return strings.TrimSpace(spaceRE.ReplaceAllString(html.UnescapeString(tagRE.ReplaceAllString(s, " ")), " "))
}

var _ sources.JobSource = (*Client)(nil)
