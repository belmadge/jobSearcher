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
