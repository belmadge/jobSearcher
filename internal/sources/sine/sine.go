package sine

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"jobsearcher/internal/domain"
	"jobsearcher/internal/sources"
)

const PortalURL = "https://www.online.maceio.al.gov.br/n/talentos/"

type Client struct {
	HTTPClient *http.Client
	URL        string
}

func NewClient() *Client {
	return &Client{HTTPClient: &http.Client{Timeout: 15 * time.Second}, URL: PortalURL}
}

func (c *Client) Name() string { return "sine-maceio" }

var (
	tagRE = regexp.MustCompile(`(?s)<[^>]*>`)
	blockRE = regexp.MustCompile(`(?i)</(div|p|li|h[1-6]|article|section|br|tr|header|footer|a)>`)
	spaceRE = regexp.MustCompile(`[\t\x0b\f\r ]+`)
	idRE = regexp.MustCompile(`#\s*(\d{3,})`)
	vacancyTitleRE = regexp.MustCompile(`(?i)^(.{4,160}?)\s*:\s*\d+\s*vaga(?:\(s\))?`)
	dateRE = regexp.MustCompile(`(?i)data limite para se candidatar:\s*(\d{2}/\d{2}/\d{4})`)
)

func (c *Client) FetchJobs(ctx context.Context, q sources.Query) ([]domain.Job, error) {
	url := strings.TrimSpace(c.URL)
	if url == "" { url = PortalURL }
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil { return nil, err }
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "pt-BR,pt;q=0.9,en;q=0.8")
	// Use a normal browser UA: some public portals serve a reduced page to unknown bots.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.online.maceio.al.gov.br/")
	resp, err := c.HTTPClient.Do(req)
	if err != nil { return nil, fmt.Errorf("SINE Maceió: %w", err) }
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil { return nil, err }
	if resp.StatusCode < 200 || resp.StatusCode >= 300 { return nil, fmt.Errorf("SINE Maceió: HTTP %d", resp.StatusCode) }
	jobs := ParseListings(string(body))
	if len(jobs) == 0 {
		return nil, fmt.Errorf("SINE Maceió: a página respondeu, mas nenhum anúncio foi reconhecido (HTML recebido: %d bytes); o portal pode ter alterado a estrutura ou exigido conteúdo dinâmico", len(body))
	}
	terms := normalizeTerms(q.Terms)
	out := make([]domain.Job, 0, len(jobs))
	for _, job := range jobs {
		if len(terms) > 0 {
			haystack := normalize(job.Title + " " + job.Description)
			matched := false
			for _, term := range terms { if strings.Contains(haystack, term) { matched = true; break } }
			if !matched { continue }
		}
		out = append(out, job)
	}
	return out, nil
}

// ParseListings reads the public listing page without relying on private APIs.
// Each result links to the official listing page; users can locate the posting
// by its visible SINE vacancy number.
func ParseListings(raw string) []domain.Job {
	text := blockRE.ReplaceAllString(raw, "$0\n")
	text = tagRE.ReplaceAllString(text, "\n")
	text = html.UnescapeString(text)
	lines := strings.Split(text, "\n")
	clean := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(spaceRE.ReplaceAllString(line, " "))
		if line != "" { clean = append(clean, line) }
	}
	full := strings.Join(clean, "\n")
	ids := idRE.FindAllStringSubmatchIndex(full, -1)
	out := []domain.Job{}
	seen := map[string]bool{}
	for _, match := range ids {
		id := full[match[2]:match[3]]
		if seen[id] { continue }
		seen[id] = true
		prefix := strings.TrimSpace(full[:match[0]])
		prev := strings.Split(prefix, "\n")
		title := ""
		for i := len(prev)-1; i >= 0 && i >= len(prev)-5; i-- {
			candidate := strings.TrimSpace(prev[i])
			lower := strings.ToLower(candidate)
			if candidate == "" || strings.Contains(lower, "ver mais") || strings.Contains(lower, "vagas disponíveis") || strings.Contains(lower, "filtros") { continue }
			if len(candidate) >= 4 && !strings.Contains(lower, "tutorial") { title = candidate; break }
		}
		title = strings.TrimSpace(strings.Trim(title, "·• "))
		if title == "" { continue }
		// Take only the text following this vacancy number until the next ID.
		start := match[1]
		end := len(full)
		for _, next := range ids { if next[0] > match[0] { end = next[0]; break } }
		detail := strings.TrimSpace(full[start:end])
		if len(detail) > 1800 { detail = detail[:1800] }
		if strings.Contains(normalize(detail), "inscricoes encerradas") { continue }
		deadline := ""
		if m := dateRE.FindStringSubmatch(detail); len(m) == 2 {
			deadline = m[1]
			if due, err := time.Parse("02/01/2006", deadline); err == nil && due.Before(time.Now().Truncate(24*time.Hour)) {
				continue
			}
		}
		out = append(out, domain.Job{
			ID: "sine-maceio:" + id, Source: "sine-maceio", Title: title,
			URL: PortalURL, ApplyURL: PortalURL, Location: "Maceió, AL, Brasil",
			WorkplaceType: "Presencial", Description: "Código SINE: " + id + ". " + detail,
			PostedAt: "", UpdatedAt: deadline,
		})
	}
	if len(out) == 0 {
		return parseListingsWithoutCodes(clean)
	}
	return out
}

// Some versions of the portal omit the visible #code from the rendered card.
// In that case, recover cards from their visible "TITLE : N vaga(s)" heading.
func parseListingsWithoutCodes(lines []string) []domain.Job {
	// Inline HTML elements can split a visible heading across several lines.
	// Flatten the cleaned text, then find the portal's "TITLE : N vaga(s)" pattern.
	flat := strings.Join(lines, " ")
	flat = strings.Join(strings.Fields(flat), " ")
	headingRE := regexp.MustCompile(`(?i)([A-ZÀ-Ý0-9][A-ZÀ-Ý0-9 /().,'-]{3,150}?)\s*:\s*\d+\s*vaga(?:\(s\))?`)
	matches := headingRE.FindAllStringSubmatchIndex(flat, -1)
	out := []domain.Job{}
	seen := map[string]bool{}
	for i, match := range matches {
		title := strings.TrimSpace(flat[match[2]:match[3]])
		title = strings.Trim(title, " ·•")
		if title == "" || strings.Contains(normalize(title), "total de vagas") {
			continue
		}
		if len(title) > 160 {
			title = title[len(title)-160:]
		}
		start := match[1]
		end := len(flat)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		detail := strings.TrimSpace(flat[start:end])
		if len(detail) > 1800 {
			detail = detail[:1800]
		}
		normalizedDetail := normalize(detail)
		if strings.Contains(normalizedDetail, "inscricoes encerradas") {
			continue
		}
		deadline := ""
		if m := dateRE.FindStringSubmatch(detail); len(m) == 2 {
			deadline = m[1]
			if due, err := time.Parse("02/01/2006", deadline); err == nil && due.Before(time.Now().Truncate(24*time.Hour)) {
				continue
			}
		}
		key := normalize(title)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, domain.Job{
			ID: "sine-maceio:title:" + key, Source: "sine-maceio", Title: title,
			URL: PortalURL, ApplyURL: PortalURL, Location: "Maceió, AL, Brasil",
			WorkplaceType: "Presencial", Description: detail, UpdatedAt: deadline,
		})
	}
	return out
}

func normalizeTerms(values []string) []string {
	out := []string{}
	for _, value := range values {
		v := normalize(value)
		if v == "vendedora" { out = append(out, "vendedora", "vendedor"); continue }
		if v == "vendedor" { out = append(out, "vendedor", "vendedora"); continue }
		if v != "" { out = append(out, v) }
	}
	return out
}

func normalize(value string) string {
	value = strings.ToLower(value)
	return strings.NewReplacer("á","a","à","a","â","a","ã","a","é","e","ê","e","í","i","ó","o","ô","o","õ","o","ú","u","ç","c").Replace(value)
}

var _ sources.JobSource = (*Client)(nil)
