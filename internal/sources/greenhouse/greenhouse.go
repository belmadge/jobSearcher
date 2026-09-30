// Package greenhouse adapts Greenhouse's public Job Board API to domain.Job.
package greenhouse

import (
	"context"
	"encoding/json"
	"errors"
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

const defaultBaseURL = "https://boards-api.greenhouse.io/v1/boards"

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient() *Client {
	return &Client{HTTPClient: &http.Client{Timeout: 45 * time.Second}, BaseURL: defaultBaseURL}
}
func NewClientWithHTTP(client *http.Client, baseURL string) *Client {
	return &Client{HTTPClient: client, BaseURL: strings.TrimRight(baseURL, "/")}
}
func (c *Client) Name() string { return "greenhouse" }

// API DTOs remain private to this adapter; the rest of the application sees domain.Job only.
type boardResponse struct {
	Jobs []jobDTO `json:"jobs"`
}
type jobDTO struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	UpdatedAt string `json:"updated_at"`
	Location  struct {
		Name string `json:"name"`
	} `json:"location"`
	AbsoluteURL string `json:"absolute_url"`
	Content     string `json:"content"`
	Offices     []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	} `json:"offices"`
	Departments []struct {
		Name string `json:"name"`
	} `json:"departments"`
}

func (c *Client) FetchJobs(ctx context.Context, q sources.Query) ([]domain.Job, error) {
	if len(q.BoardTokens) == 0 {
		return nil, errors.New("greenhouse: no board tokens configured; set GREENHOUSE_BOARD_TOKENS")
	}
	jobs := make([]domain.Job, 0)
	queriedBoards := 0
	var errs []error
	for _, token := range q.BoardTokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		if !validBoardToken.MatchString(token) {
			errs = append(errs, fmt.Errorf("board %q: invalid board token", token))
			continue
		}
		queriedBoards++
		boardJobs, err := c.fetchBoard(ctx, token)
		if err != nil {
			errs = append(errs, fmt.Errorf("board %q: %w", token, err))
			continue
		}
		jobs = append(jobs, boardJobs...)
	}
	if queriedBoards == 0 && len(errs) == 0 {
		return nil, errors.New("greenhouse: no valid board tokens configured")
	}
	return jobs, errors.Join(errs...)
}

func (c *Client) fetchBoard(ctx context.Context, token string) ([]domain.Job, error) {
	url := c.BaseURL + "/" + token + "/jobs?content=true"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request jobs: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		if resp.StatusCode == http.StatusTooManyRequests {
			return nil, fmt.Errorf("HTTP 429 rate limited (Retry-After: %s)", resp.Header.Get("Retry-After"))
		}
		return nil, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	var payload boardResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	jobs := make([]domain.Job, 0, len(payload.Jobs))
	for _, dto := range payload.Jobs {
		jobs = append(jobs, normalize(token, dto))
	}
	return jobs, nil
}

var tags = regexp.MustCompile(`<[^>]*>`)
var whitespace = regexp.MustCompile(`\s+`)
var validBoardToken = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func plainText(s string) string {
	return whitespace.ReplaceAllString(strings.TrimSpace(tags.ReplaceAllString(html.UnescapeString(s), " ")), " ")
}

func normalize(token string, dto jobDTO) domain.Job {
	location := strings.TrimSpace(dto.Location.Name)
	for _, office := range dto.Offices {
		value := strings.TrimSpace(strings.Join([]string{office.Name, office.Location}, ", "))
		if value != "" && !strings.Contains(location, value) {
			location = strings.Trim(strings.Join([]string{location, value}, "; "), "; ")
		}
	}
	description := plainText(dto.Content)
	lower := strings.ToLower(location + " " + description)
	workplace := ""
	switch {
	case strings.Contains(lower, "hybrid"):
		workplace = "Hybrid"
	case strings.Contains(lower, "remote"):
		workplace = "Remote"
	case strings.Contains(lower, "onsite"), strings.Contains(lower, "on-site"), strings.Contains(lower, "on site"), strings.Contains(lower, "in person"):
		workplace = "Onsite"
	}
	departments := make([]string, 0, len(dto.Departments))
	for _, d := range dto.Departments {
		if d.Name != "" {
			departments = append(departments, d.Name)
		}
	}
	return domain.Job{
		ID: fmt.Sprintf("greenhouse:%s:%d", token, dto.ID), Source: "greenhouse", Title: dto.Title,
		Company: token, URL: dto.AbsoluteURL, ApplyURL: dto.AbsoluteURL,
		Location: location, WorkplaceType: workplace, Description: description,
		Requirements: strings.Join(departments, ", "), UpdatedAt: dto.UpdatedAt,
	}
}

var _ sources.JobSource = (*Client)(nil)
