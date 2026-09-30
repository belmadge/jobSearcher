package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"jobsearcher/internal/config"
	"jobsearcher/internal/dedupe"
	"jobsearcher/internal/domain"
	"jobsearcher/internal/filter"
	"jobsearcher/internal/matching"
	"jobsearcher/internal/report"
	"jobsearcher/internal/sources"
	"jobsearcher/internal/sources/greenhouse"
	"jobsearcher/internal/storage/sqlite"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	command := "run"
	sourceName := ""
	dryRun := false
	if len(args) > 0 {
		command = args[0]
		for _, arg := range args[1:] {
			switch {
			case arg == "--dry-run":
				dryRun = true
			case strings.HasPrefix(arg, "--source="):
				sourceName = strings.TrimPrefix(arg, "--source=")
			}
		}
	}

	if err := config.LoadDotEnv(".env"); err != nil {
		return err
	}
	profile, err := config.Load[config.Profile]("config/profile.json")
	if err != nil {
		return err
	}
	search, err := config.Load[config.Search]("config/search.json")
	if err != nil {
		return err
	}
	boards, err := config.Load[config.Boards]("config/boards.json")
	if err != nil {
		return err
	}

	switch command {
	case "run":
		return runSearch(ctx, profile, search, boards, sourceName, dryRun)
	case "report":
		return latestReport()
	case "stats":
		return stats(ctx)
	case "--help", "help":
		fmt.Println("jobsearch run [--dry-run] [--source=greenhouse]")
		fmt.Println("jobsearch report")
		fmt.Println("jobsearch stats")
		return nil
	default:
		return fmt.Errorf("unknown command %q; use --help", command)
	}
}

func runSearch(ctx context.Context, profile config.Profile, search config.Search, boards config.Boards, sourceName string, dryRun bool) error {
	var source sources.JobSource
	switch sourceName {
	case "", "greenhouse":
		source = greenhouse.NewClient()
	case "mock":
		source = sources.Mock{}
	default:
		return fmt.Errorf("unknown source %q", sourceName)
	}

	tokens := make([]string, 0, len(boards.Greenhouse))
	for _, board := range boards.Greenhouse {
		if board.Enabled && strings.TrimSpace(board.Token) != "" {
			tokens = append(tokens, board.Token)
		}
	}
	if env := strings.TrimSpace(os.Getenv("GREENHOUSE_BOARD_TOKENS")); env != "" {
		for _, token := range strings.Split(env, ",") {
			if strings.TrimSpace(token) != "" {
				tokens = append(tokens, strings.TrimSpace(token))
			}
		}
	}
	tokens = unique(tokens)

	if source.Name() == "greenhouse" && len(tokens) == 0 {
		return errors.New("no Greenhouse boards configured; add enabled boards to config/boards.json or set GREENHOUSE_BOARD_TOKENS")
	}

	jobs, err := source.FetchJobs(ctx, sources.Query{
		Terms: search.PreferredTitles,
		Location: search.Location,
		BoardTokens: tokens,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "source warning:", err)
	}
	if len(jobs) == 0 && err != nil {
		return err
	}

	accepted := make([]domain.Job, 0, len(jobs))
	uncertain := make([]domain.Job, 0)
	rejected := 0

	for i := range jobs {
		job := filter.Evaluate(jobs[i])
		switch job.LocationEligible {
		case domain.LocationEligible:
			matching.Score(&job, profile)
			job.RecommendationStatus = recommendation(job.FitScore, search.MinimumFitScore)
			accepted = append(accepted, job)
		case domain.LocationUnknown:
			job.RecommendationStatus = domain.UncertainLocation
			uncertain = append(uncertain, job)
		default:
			job.RecommendationStatus = domain.RejectedLocation
			rejected++
		}
	}

	accepted = dedupe.Jobs(accepted)
	sort.SliceStable(accepted, func(i, j int) bool { return accepted[i].FitScore > accepted[j].FitScore })

	if !dryRun {
		if err := os.MkdirAll("data", 0o755); err != nil {
			return err
		}
		repo, err := sqlite.Open(filepath.Join("data", "jobsearch.db"))
		if err != nil {
			return err
		}
		defer repo.Close()
		for i := range accepted {
			if _, err := repo.SaveJob(ctx, &accepted[i]); err != nil {
				return err
			}
		}
		for i := range uncertain {
			if _, err := repo.SaveJob(ctx, &uncertain[i]); err != nil {
				return err
			}
		}
	}

	now := time.Now().Format("2006-01-02")
	markdown := report.Markdown(accepted, uncertain, rejected, len(jobs), now)
	if err := os.MkdirAll("reports", 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("reports", now+".md"), []byte(markdown), 0o644); err != nil {
		return err
	}
	jsonData, err := json.MarshalIndent(map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"jobs": accepted,
		"uncertain": uncertain,
	}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("reports", now+".json"), jsonData, 0o644); err != nil {
		return err
	}

	fmt.Printf("Vagas encontradas: %d\nElegíveis: %d\nIncertas: %d\nRejeitadas por localização: %d\n", len(jobs), len(accepted), len(uncertain), rejected)
	for _, job := range accepted {
		fmt.Printf("%3d  %-45s  %s  %s\n", job.FitScore, job.Title, job.Company, job.URL)
	}
	return nil
}

func recommendation(score, minimum int) domain.RecommendationStatus {
	if score >= 80 {
		return domain.Recommended
	}
	if score >= minimum {
		return domain.PossibleMatch
	}
	return domain.LowMatch
}

func latestReport() error {
	entries, err := os.ReadDir("reports")
	if err != nil {
		return err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 0 {
		return errors.New("no reports found")
	}
	sort.Strings(names)
	data, err := os.ReadFile(filepath.Join("reports", names[len(names)-1]))
	if err != nil {
		return err
	}
	fmt.Print(string(data))
	return nil
}

func stats(ctx context.Context) error {
	repo, err := sqlite.Open(filepath.Join("data", "jobsearch.db"))
	if err != nil {
		return err
	}
	defer repo.Close()
	stats, err := repo.GetStats(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Total: %d\nNovas: %d\nVistas: %d\nAtualizadas: %d\nExpiradas: %d\nRecomendadas: %d\nPossíveis: %d\n", stats.Total, stats.New, stats.Seen, stats.Updated, stats.Expired, stats.Recommended, stats.Possible)
	return nil
}

func unique(values []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		key := strings.TrimSpace(value)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
