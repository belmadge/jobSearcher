# JobSearcher architecture

## Current scope (Phases 1–2)

The first phase established a small Go CLI with mock discovery, deterministic location filtering, explainable profile matching, configuration, and unit tests. Phase 2 adds the public Greenhouse Job Board API adapter. No applications are submitted and no messages are sent. Durable storage and dated daily reports are later phases.

## Data flow

```text
JobSource -> normalized Job -> location filter -> profile matching -> deduplication -> report
```

Each boundary has a small Go interface or package. `internal/sources` owns discovery (`FetchJobs(ctx, query)`); `internal/sources/greenhouse` keeps API DTOs and HTTP behavior isolated; `internal/filter` makes location decisions without relying on a score; `internal/matching` computes weighted, explainable dimensions and gaps from the JSON profile; `internal/report` renders results. The `mock` source proves the flow without network access. Greenhouse board tokens are configured as a list in `config/search.json` or `GREENHOUSE_BOARD_TOKENS` in `.env`. Persistence is intentionally deferred; `--dry-run` is supported and no persistent changes are made.

## Packages

- `cmd/jobsearch`: command parsing and orchestration for `run` and `report`.
- `internal/domain`: normalized job and decision models.
- `internal/config`: profile and search JSON loading.
- `internal/filter`: Brazil-eligible remote, Maceió onsite/hybrid, and uncertain classification.
- `internal/matching`: weighted score, seniority assessment, requirement coverage, gaps, reasons.
- `internal/sources`: source contract and deterministic mock listings.
- `internal/sources/greenhouse`: public Greenhouse Job Board API adapter and normalization.
- `internal/report`: readable Markdown output.

## Location policy

Workplace type and geographic eligibility are separate decisions. Remote is accepted only with explicit Brazil, LATAM, South America, Worldwide, or Americas scope, and rejected for explicit incompatible regions/countries. Onsite and hybrid listings are accepted only when the physical location identifies Maceió/Alagoas, Brazil. Incomplete or conflicting information is `uncertain_location`, never silently accepted.

## Configuration and privacy

`config/profile.json` contains the user profile and matching weights; `config/search.json` contains search preferences. Secrets belong in environment variables and `.env` is ignored by Git; `.env.example` documents optional settings without credentials. Profile details are not compiled into matching code.

## Next phases

Phase 3 adds another source. Phase 4 adds local persistence and duplicate history. Phase 5 adds dated reports. Later work can improve discovery and optionally add a scheduler that invokes the same CLI. Each source must respect its public access terms and rate limits. See [sources.md](sources.md) for the current adapter details.
