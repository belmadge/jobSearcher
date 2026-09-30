package programathor

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"jobsearcher/internal/domain"
	"jobsearcher/internal/sources"
)

const defaultURL = "https://programathor.com.br/jobs-api"

type Client struct {
	HTTPClient *http.Client
	URL        string
}

func NewClient() *Client {
	return &Client{HTTPClient: &http.Client{Timeout: 20 * time.Second}, URL: defaultURL}
}

func (c *Client) Name() string { return "programathor" }

var jobLinkRE = regexp.MustCompile(`(?is)<a[^>]+href="(/jobs/[^"]+)"[^>]*>(.*?)</a>`)
var tagRE = regexp.MustCompile(`(?s)<[^>]+>`)
var spaceRE = regexp.MustCompile(`\s+`)
var staleRE = regexp.MustCompile(`(?i)\b(vencida|expired)\b`)

func (c *Client) FetchJobs(ctx context.Context, q sources.Query) ([]domain.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL, nil)
	if err != nil { return nil, err }
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "JobSearcher/1.0")
	resp, err := c.HTTPClient.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil { return nil, err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("programathor: HTTP %d", resp.StatusCode) }

	out := make([]domain.Job, 0)
	seen := map[string]bool{}
	for _, match := range jobLinkRE.FindAllStringSubmatch(string(body), -1) {
		if len(match) < 3 { continue }
		path := html.UnescapeString(match[1])
		text := clean(match[2])
		if text == "" || staleRE.MatchString(text) { continue }
		if seen[path] { continue }
		seen[path] = true

		title := titleFromPath(path)
		if title == "" { continue }
		fullURL := "https://programathor.com.br" + path
		location, workplace := inferLocation(text)
		out = append(out, domain.Job{
			ID: "programathor:" + path,
			Source: "programathor",
			Title: title,
			Company: inferCompany(text, title),
			URL: fullURL,
			CanonicalURL: fullURL,
			Location: location,
			WorkplaceType: workplace,
			EmploymentType: inferEmployment(text),
			Seniority: inferSeniority(text),
			Description: text,
			Requirements: text,
			Salary: inferSalary(text),
		})
	}
	return out, nil
}

func clean(s string) string {
	s = html.UnescapeString(tagRE.ReplaceAllString(s, " "))
	return strings.TrimSpace(spaceRE.ReplaceAllString(s, " "))
}

func titleFromPath(path string) string {
	u, err := url.PathUnescape(path)
	if err != nil { return "" }
	parts := strings.Split(strings.Trim(u, "/"), "/")
	if len(parts) == 0 { return "" }
	slug := parts[len(parts)-1]
	if i := strings.Index(slug, "-"); i >= 0 {
		if _, err := strconv.Atoi(slug[:i]); err == nil { slug = slug[i+1:] }
	}
	words := strings.Fields(strings.ReplaceAll(slug, "-", " "))
	for i := range words {
		if len(words[i]) > 0 { words[i] = strings.ToUpper(words[i][:1]) + words[i][1:] }
	}
	return strings.Join(words, " ")
}

func inferLocation(text string) (string, string) {
	l := strings.ToLower(text)
	if strings.Contains(l, "remoto") { return "Remote - Brazil eligibility to be evaluated", "Remote" }
	if strings.Contains(l, "maceió") || strings.Contains(l, "maceio") { return "Maceió, Alagoas, Brazil", "Hybrid" }
	if strings.Contains(l, "presencial") { return text, "Onsite" }
	if strings.Contains(l, "híbrido") || strings.Contains(l, "hibrido") { return text, "Hybrid" }
	return "", "Unknown"
}

func inferSeniority(text string) string {
	l := strings.ToLower(text)
	switch {
	case strings.Contains(l, "sênior"), strings.Contains(l, "senior"): return "Senior"
	case strings.Contains(l, "pleno"): return "Mid-level"
	case strings.Contains(l, "júnior"), strings.Contains(l, "junior"): return "Junior"
	default: return ""
	}
}

func inferEmployment(text string) string {
	l := strings.ToLower(text)
	switch {
	case strings.Contains(l, "clt / pj"), strings.Contains(l, "clt/pj"): return "CLT/PJ"
	case strings.Contains(l, "clt"): return "CLT"
	case strings.Contains(l, "pj"): return "PJ"
	case strings.Contains(l, "estágio"), strings.Contains(l, "estagio"): return "Internship"
	default: return ""
	}
}

func inferSalary(text string) string {
	re := regexp.MustCompile(`(?i)(até|acima de)\s*r\$\s*[0-9.]+(?:,[0-9]+)?`)
	if m := re.FindString(text); m != "" { return m }
	return ""
}

func inferCompany(text, title string) string {
	normalized := strings.TrimSpace(strings.TrimPrefix(text, title))
	if normalized == "" { return "Programathor" }
	for _, marker := range []string{"Remoto", "Presencial", "Híbrido", "Hibrido"} {
		if i := strings.Index(strings.ToLower(normalized), strings.ToLower(marker)); i > 0 {
			normalized = strings.TrimSpace(normalized[:i])
			break
		}
	}
	normalized = strings.TrimSpace(staleRE.ReplaceAllString(normalized, ""))
	if normalized == "" { return "Programathor" }
	return normalized
}

var _ sources.JobSource = (*Client)(nil)
