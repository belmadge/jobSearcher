# Job sources

## Greenhouse Job Board API

- **Documentation:** [Greenhouse Job Board API](https://developers.greenhouse.io/job-board.html)
- **Adapter:** `internal/sources/greenhouse`
- **Endpoints:** `GET https://boards-api.greenhouse.io/v1/boards/{board_token}/jobs?content=true`
- **Authentication:** Greenhouse documents its public Job Board GET endpoints as unauthenticated. Application submission endpoints require authentication; this project only issues GET requests.

The API is scoped to one employer's public job board token per request; it is not a cross-company job search service. The CLI accepts multiple board tokens from `config/search.json` (`greenhouse_boards`) and/or the comma-separated `GREENHOUSE_BOARD_TOKENS` variable in `.env`. Tokens identify public boards and are not credentials. The adapter queries them sequentially so one board failure does not discard successful results from other boards.

`content=true` returns published job posts with descriptions, office and department information. The adapter converts API DTOs privately into `domain.Job`; IDs are stable as `greenhouse:{board_token}:{job_id}`. Greenhouse's list response does not provide a company name, so the board token is used as the company label. HTML is stripped and entities are decoded before the existing location and matching stages run.

Greenhouse does not expose a cross-board keyword query in this endpoint. The configured title terms are passed through the source query abstraction; the adapter fetches each configured board's public postings and leaves relevance filtering to the existing location and matching pipeline. A board token must be supplied by the user for every employer to search.

### Limits and failure handling

The official Job Board API documentation does not publish a numeric rate limit. The adapter therefore makes one sequential request per configured board, uses a 20-second HTTP timeout, does not retry automatically, and reports HTTP 429 including `Retry-After` when provided. HTTP failures, invalid JSON, timeouts, and canceled contexts are returned as source errors. Successful results from other boards are retained when a board fails. The response body is limited to 16 MiB per board response.

The API returns only jobs published on the configured boards. It does not establish that a role is eligible to be worked from Brazil: location scope is still decided by the shared deterministic filter, which examines both normalized location and description. Ambiguous roles remain `uncertain_location`.

### Local setup

Copy `.env.example` to `.env` and enter one or more public board tokens, separated by commas. A board token is normally the employer slug used in its Greenhouse-hosted jobs URL. Alternatively, list tokens in `config/search.json` under `greenhouse_boards`.

```powershell
go test ./...
go run ./cmd/jobsearch run
```

The real source is the default. Use `go run ./cmd/jobsearch run --source mock` for an offline demonstration. The command does not submit applications or send messages.
