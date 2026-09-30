# JobSearcher — Codex Instructions

## Goal

Build a job-search automation that runs on demand and can also be scheduled daily.

The automation must:
1. Search current job postings from reliable/public sources.
2. Keep ONLY:
   - Remote jobs, regardless of employer location, OR
   - On-site/hybrid jobs located in Maceió, Alagoas, Brazil.
3. Evaluate fit against the candidate profile defined in \`config/profile.json\`.
4. Deduplicate jobs across sources and executions.
5. Produce a ranked, explainable daily report with:
   - Job title
   - Company
   - Location / workplace type
   - Match score
   - Why it matches the profile
   - Missing/weak requirements
   - Seniority fit
   - Salary, when available
   - Publication/update date, when available
   - Direct job URL
6. Never apply automatically. The system only discovers, filters, analyzes, and reports jobs.
7. Avoid scraping sources that explicitly prohibit automated access. Prefer public APIs, feeds, structured job-board endpoints, and search providers where permitted.

## Candidate profile

Use \`config/profile.json\` as the source of truth. Do not hard-code candidate details across the codebase.

Core profile signals from the current resume:
- Backend Developer / Backend Software Engineer
- Go
- PostgreSQL / SQL
- AWS
- RabbitMQ
- GraphQL
- Docker
- Kubernetes
- Terraform
- CI/CD
- Datadog / observability / tracing
- Automated testing
- API/backend service development
- Data migrations
- Production troubleshooting
- Fintech / production systems experience
- AI Agents and AI Skills as emerging skills

Experience level:
- Junior Backend Developer at Dock since 2024
- Junior Software Engineer at Levee from Oct 2021 to Apr 2024
- Fullstack Intern at Simge from Dec 2020 to Oct 2021
- Backend Developer experience at Facedoor

Education:
- Systems Analysis, Centro Universitário Cesmac

English:
- Advanced / IELTS

Certification:
- Go Expert, Full Cycle, May 2024

## Matching policy

### Hard filters

Reject if:
- Workplace is explicitly on-site and location is not Maceió/Alagoas/Brazil.
- Workplace is explicitly hybrid and the job requires regular physical presence outside Maceió/Alagoas/Brazil.
- The posting is closed, expired, or clearly no longer accepting applications.
- The role is clearly unrelated to backend/software engineering.

Accept when:
- Explicitly remote.
- Remote from Brazil / Latin America / worldwide.
- Hybrid in Maceió/Alagoas/Brazil.
- On-site in Maceió/Alagoas/Brazil.

When workplace/location cannot be confidently determined, keep the job in an \`uncertain_location\` bucket instead of silently treating it as remote.

### Fit scoring

Use an explainable weighted score, not a black box.

Suggested weights:
- 35% technical stack overlap
- 20% backend/API responsibility overlap
- 15% seniority/years-of-experience fit
- 10% cloud/infrastructure overlap
- 10% domain/production experience overlap
- 5% language/communication fit
- 5% AI/modern engineering alignment

Do not penalize a job merely because it contains additional nice-to-have technologies.

Separate:
- \`must_have_match\`
- \`nice_to_have_match\`
- \`gaps\`
- \`fit_score\`

A job can be high-fit even when not every skill is present.

### Seniority handling

Prioritize roles compatible with Junior, Mid-level, Associate, and Software Engineer positions where requirements are reasonably aligned.

Do not discard a role solely because the title says \`Software Engineer\`, \`Backend Engineer\`, or similar. Use requirements and years of experience.

Do not present Senior/Staff/Principal roles as high-fit unless the posting's actual requirements support the candidate's experience.

## Search strategy

Design the system around pluggable \`Source\` adapters.

Initial source types should support:
- Public ATS endpoints such as Lever/Greenhouse where accessible.
- Other legally accessible job APIs/search sources added later.

Do not tie the architecture to a single website.

Normalize all postings into a common model:
- id
- source
- title
- company
- url
- apply_url
- location
- workplace_type
- employment_type
- seniority
- description
- requirements
- salary
- posted_at
- updated_at
- fetched_at

## Daily execution

The CLI should support:
- \`jobsearch run\`
- \`jobsearch run --source lever\`
- \`jobsearch run --dry-run\`
- \`jobsearch report\`

The scheduled execution should be idempotent.

Persist:
- Seen jobs
- First-seen timestamp
- Last-seen timestamp
- Last match score
- Status such as new/seen/expired/rejected

The daily report should emphasize newly discovered/high-fit jobs first.

## Configuration

Use environment variables for secrets/API keys.

Never commit secrets.

Provide:
- \`.env.example\`
- \`config/profile.json\`
- \`config/search.json\`

Search configuration should include:
- preferred role keywords
- excluded keywords
- allowed countries/regions for remote
- allowed physical locations
- minimum match score for report
- maximum jobs per source
- search freshness window

## Quality requirements

Use Go unless there is a strong reason otherwise.

Target:
- small packages
- interfaces for sources
- deterministic filtering
- unit tests for location and fit rules
- integration-testable source adapters
- structured logging
- graceful rate-limit/error handling

Every filtering decision should be explainable and testable.

Before adding a source, verify its current public access method and document it in \`docs/sources.md\`.

## Output

Human-readable Markdown report in \`reports/\` and machine-readable JSON in \`data/\`.

Never include private credentials in reports or logs.

## Git

Keep commits small and conventional.

Example:
- \`feat: add job source abstraction\`
- \`feat: add location filtering\`
- \`test: cover remote and Maceio rules\`
- \`docs: document source adapters\`
