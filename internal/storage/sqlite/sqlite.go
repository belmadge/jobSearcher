package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"jobsearcher/internal/dedupe"
	"jobsearcher/internal/domain"
	"jobsearcher/internal/storage"
	_ "modernc.org/sqlite"
)

type Repository struct { db *sql.DB }

func Open(path string) (*Repository, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil { return nil, fmt.Errorf("open SQLite: %w", err) }
	db.SetMaxOpenConns(1)
	for _, stmt := range []string{
		`PRAGMA journal_mode=WAL`, `PRAGMA busy_timeout=5000`,
		`CREATE TABLE IF NOT EXISTS jobs (
			record_id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '',
			canonical_url TEXT NOT NULL DEFAULT '', fallback_key TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL, recommendation_status TEXT NOT NULL,
			fit_score INTEGER NOT NULL, first_seen_at TEXT NOT NULL, last_seen_at TEXT NOT NULL,
			seen_count INTEGER NOT NULL, payload TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS job_source_ids (source TEXT NOT NULL, source_id TEXT NOT NULL, job_id INTEGER NOT NULL REFERENCES jobs(record_id), UNIQUE(source, source_id))`,
		`CREATE INDEX IF NOT EXISTS jobs_canonical_url_idx ON jobs(canonical_url) WHERE canonical_url <> ''`,
		`CREATE INDEX IF NOT EXISTS jobs_fallback_key_idx ON jobs(fallback_key) WHERE fallback_key <> ''`,
		`CREATE INDEX IF NOT EXISTS jobs_last_seen_idx ON jobs(last_seen_at)`,
	} {
		if _, err := db.Exec(stmt); err != nil { _ = db.Close(); return nil, fmt.Errorf("initialize SQLite: %w", err) }
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error { return r.db.Close() }

type stored struct { id int64; job domain.Job }

func decode(row *sql.Row) (stored, error) {
	var result stored; var payload string
	err := row.Scan(&result.id, &payload)
	if err != nil { return stored{}, err }
	if err := json.Unmarshal([]byte(payload), &result.job); err != nil { return stored{}, err }
	return result, nil
}

func (r *Repository) lookup(ctx context.Context, tx *sql.Tx, source, sourceID, canonical, fallback string) (stored, error) {
	queries := []struct{ sql string; arg any }{}
	if source != "" && sourceID != "" { queries = append(queries, struct{sql string; arg any}{`SELECT j.record_id,j.payload FROM job_source_ids s JOIN jobs j ON j.record_id=s.job_id WHERE s.source=? AND s.source_id=? LIMIT 1`, source}, sourceID) }
	_ = queries
	if source != "" && sourceID != "" {
		x, err := decode(tx.QueryRowContext(ctx, `SELECT j.record_id,j.payload FROM job_source_ids s JOIN jobs j ON j.record_id=s.job_id WHERE s.source=? AND s.source_id=? LIMIT 1`, source, sourceID))
		if err == nil { return x, nil }; if !errors.Is(err, sql.ErrNoRows) { return stored{}, err }
	}
	if canonical != "" { x, err := decode(tx.QueryRowContext(ctx, `SELECT record_id,payload FROM jobs WHERE canonical_url=? ORDER BY record_id LIMIT 1`, canonical)); if err == nil { return x,nil }; if !errors.Is(err,sql.ErrNoRows) { return stored{},err } }
	if fallback != "" { x, err := decode(tx.QueryRowContext(ctx, `SELECT record_id,payload FROM jobs WHERE fallback_key=? ORDER BY record_id LIMIT 1`, fallback)); if err == nil { return x,nil }; if !errors.Is(err,sql.ErrNoRows) { return stored{},err } }
	return stored{}, sql.ErrNoRows
}

func material(j domain.Job) string {
	return strings.Join([]string{j.Title,j.Company,j.CanonicalURL,j.Location,j.WorkplaceType,j.Description,j.Requirements,j.Salary,j.PostedAt,j.UpdatedAt,fmt.Sprint(j.FitScore),string(j.RecommendationStatus),string(j.LocationEligible)}, "\x00")
}

func (r *Repository) SaveJob(ctx context.Context, job *domain.Job) (storage.SaveResult, error) {
	job.CanonicalURL = dedupe.CanonicalURL(job.URL)
	fallback := dedupe.FallbackKey(*job)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	tx, err := r.db.BeginTx(ctx, nil); if err != nil { return storage.SaveResult{}, err }; defer tx.Rollback()
	old, err := r.lookup(ctx, tx, job.Source, job.ID, job.CanonicalURL, fallback)
	result := storage.SaveResult{}
	var recordID int64
	if errors.Is(err, sql.ErrNoRows) {
		result.New = true
		job.Status, job.IsNew, job.SeenCount = "new", true, 1
		job.FirstSeenAt, job.LastSeenAt = now, now
	} else if err != nil { return result, fmt.Errorf("find duplicate job: %w", err) } else {
		recordID, result.Duplicate = old.id, old.job.Source != job.Source || old.job.ID != job.ID
		job.FirstSeenAt, job.LastSeenAt, job.SeenCount = old.job.FirstSeenAt, now, old.job.SeenCount+1
		job.IsNew = false
		if material(old.job) != material(*job) { job.Status, result.Updated = "updated", true } else { job.Status = "seen" }
	}
	if recordID == 0 {
		payload, err := json.Marshal(job); if err != nil { return result, err }
		res, err := tx.ExecContext(ctx, `INSERT INTO jobs(source,source_id,canonical_url,fallback_key,status,recommendation_status,fit_score,first_seen_at,last_seen_at,seen_count,payload) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, job.Source, job.ID, job.CanonicalURL, fallback, job.Status, job.RecommendationStatus, job.FitScore, job.FirstSeenAt, job.LastSeenAt, job.SeenCount, string(payload))
		if err != nil { return result, fmt.Errorf("insert job: %w", err) }; recordID, err = res.LastInsertId(); if err != nil { return result, err }
	} else {
		payload, err := json.Marshal(job); if err != nil { return result, err }
		if _, err := tx.ExecContext(ctx, `UPDATE jobs SET canonical_url=?,fallback_key=?,status=?,recommendation_status=?,fit_score=?,last_seen_at=?,seen_count=?,payload=? WHERE record_id=?`, job.CanonicalURL, fallback, job.Status, job.RecommendationStatus, job.FitScore, job.LastSeenAt, job.SeenCount, string(payload), recordID); err != nil { return result, fmt.Errorf("update job: %w", err) }
	}
	if job.Source != "" && job.ID != "" { if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO job_source_ids(source,source_id,job_id) VALUES(?,?,?)`, job.Source, job.ID, recordID); err != nil { return result, fmt.Errorf("save source identifier: %w", err) } }
	if err := tx.Commit(); err != nil { return result, err }
	return result, nil
}

func (r *Repository) GetJob(ctx context.Context, sourceID string) (domain.Job, error) {
	var payload string
	err := r.db.QueryRowContext(ctx, `SELECT j.payload FROM jobs j LEFT JOIN job_source_ids s ON s.job_id=j.record_id WHERE j.source_id=? OR s.source_id=? ORDER BY j.record_id LIMIT 1`, sourceID, sourceID).Scan(&payload)
	if err != nil { return domain.Job{}, err }; var j domain.Job; err = json.Unmarshal([]byte(payload), &j); return j, err
}

func (r *Repository) GetJobs(ctx context.Context) ([]domain.Job, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT payload FROM jobs ORDER BY fit_score DESC, first_seen_at DESC`); if err != nil { return nil, err }; defer rows.Close()
	jobs := []domain.Job{}
	for rows.Next() { var payload string; if err := rows.Scan(&payload); err != nil { return nil, err }; var j domain.Job; if err := json.Unmarshal([]byte(payload), &j); err != nil { return nil, err }; jobs = append(jobs,j) }
	return jobs, rows.Err()
}

func (r *Repository) MarkSeen(ctx context.Context, source, sourceID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET status='seen',seen_count=seen_count+1,last_seen_at=? WHERE record_id=(SELECT job_id FROM job_source_ids WHERE source=? AND source_id=?)`, time.Now().UTC().Format(time.RFC3339Nano), source, sourceID); return err
}

func (r *Repository) lookupPayload(ctx context.Context, query string, args ...any) (domain.Job,error) {
	var payload string; if err:=r.db.QueryRowContext(ctx,query,args...).Scan(&payload); err!=nil{return domain.Job{},err}; var j domain.Job; err:=json.Unmarshal([]byte(payload),&j); return j,err
}
func (r *Repository) FindByCanonicalURL(ctx context.Context, rawURL string) (domain.Job,error) { return r.lookupPayload(ctx, `SELECT payload FROM jobs WHERE canonical_url=? LIMIT 1`, dedupe.CanonicalURL(rawURL)) }
func (r *Repository) FindBySourceID(ctx context.Context, source, sourceID string) (domain.Job,error) { return r.lookupPayload(ctx, `SELECT j.payload FROM job_source_ids s JOIN jobs j ON j.record_id=s.job_id WHERE s.source=? AND s.source_id=? LIMIT 1`, source, sourceID) }

func (r *Repository) MarkUnseenExpired(ctx context.Context, source string, before time.Time) (int64,error) {
	res,err:=r.db.ExecContext(ctx, `UPDATE jobs SET status='expired' WHERE source=? AND last_seen_at<? AND status NOT IN ('expired','rejected')`, source, before.UTC().Format(time.RFC3339Nano)); if err!=nil{return 0,err}; return res.RowsAffected()
}

func (r *Repository) GetStats(ctx context.Context) (storage.Stats,error) {
	var s storage.Stats; s.Sources=map[string]int{}
	err:=r.db.QueryRowContext(ctx, `SELECT COUNT(*),SUM(status='new'),SUM(status='seen'),SUM(status='updated'),SUM(status='expired'),SUM(recommendation_status='rejected_location'),SUM(recommendation_status='recommended'),SUM(recommendation_status='possible_match') FROM jobs`).Scan(&s.Total,&s.New,&s.Seen,&s.Updated,&s.Expired,&s.Rejected,&s.Recommended,&s.Possible)
	if err!=nil && err!=sql.ErrNoRows { return s,err }
	rows,err:=r.db.QueryContext(ctx, `SELECT source,COUNT(*) FROM jobs GROUP BY source`); if err!=nil{return s,err}; defer rows.Close()
	for rows.Next(){var source string;var count int;if err:=rows.Scan(&source,&count);err!=nil{return s,err};s.Sources[source]=count}; return s,rows.Err()
}

var _ storage.JobRepository = (*Repository)(nil)
