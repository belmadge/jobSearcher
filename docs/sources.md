# Job sources

## Greenhouse

The project supports Greenhouse's public Job Board API.

- Public GET endpoints do not require authentication.
- Jobs are fetched from configured board tokens.
- Full descriptions are requested with `content=true`.
- Configure boards in `config/boards.json` under `greenhouse`.
- Environment alternative: `GREENHOUSE_BOARD_TOKENS=company_a,company_b`.

Official API documentation: https://docs.greenhouse.io/job-board.html

## Lever

The project also supports Lever's public Postings API.

- Published postings are publicly accessible.
- The API is company/site scoped rather than a global cross-company search.
- Configure Lever site handles in `config/boards.json` under `lever`.
- Environment alternative: `LEVER_SITES=company_a,company_b`.
- The source requests JSON with `mode=json` and preserves direct hosted/apply URLs.

Official Lever Postings API documentation: https://github.com/lever/postings-api

## Source design

All providers implement `sources.JobSource` and return normalized `domain.Job` values. Source-specific DTOs stay inside the adapter package.

A source may return successful jobs together with an error when only some configured boards/sites fail. The CLI keeps successful results and reports the source warning.

No source is allowed to submit applications or contact recruiters.

## Remote OK

The project also supports Remote OK's public JSON feed for broad remote-job discovery.

- Endpoint: https://remoteok.com/api
- No authentication is required for the public feed.
- Jobs are normalized as remote candidates and still pass the project's location and profile matching rules.
- Reports preserve the original job URL and identify Remote OK as the source.
- The feed's terms require attribution and a link back to Remote OK; JobSearcher does not republish job descriptions publicly beyond its local report.

Public API: https://remoteok.com/api

## Remotive

Remotive provides a public API containing active remote job listings. The API supports a software-development category and returns candidate geographic restrictions, salary, publication date and the original Remotive job URL. The public API data may be delayed by 24 hours and has request-rate guidance, so JobSearcher uses it as a discovery source rather than a real-time feed. The local report identifies Remotive as the source and links to the original listing, as required by its public API terms.
