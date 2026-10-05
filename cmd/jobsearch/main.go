package main

import (
 "context";"encoding/json";"errors";"fmt";"os";"path/filepath";"sort";"strings";"sync";"time"
 "jobsearcher/internal/config";"jobsearcher/internal/dedupe";"jobsearcher/internal/domain";"jobsearcher/internal/filter";"jobsearcher/internal/matching";"jobsearcher/internal/report";"jobsearcher/internal/sources";"jobsearcher/internal/sources/greenhouse";"jobsearcher/internal/sources/lever";"jobsearcher/internal/sources/remoteok";"jobsearcher/internal/sources/remotive";"jobsearcher/internal/sources/programathor";"jobsearcher/internal/sources/himalayas";"jobsearcher/internal/sources/gupy";"jobsearcher/internal/sources/adzuna";"jobsearcher/internal/sources/themuse";"jobsearcher/internal/sources/arbeitnow";"jobsearcher/internal/storage/sqlite"
)
func main(){if err:=run(context.Background(),os.Args[1:]);err!=nil{fmt.Fprintln(os.Stderr,"error:",err);os.Exit(1)}}
func run(ctx context.Context,args []string)error{
 command,sourceName,dryRun:="run","all",false
 if len(args)>0{command=args[0];for _,arg:=range args[1:]{if arg=="--dry-run"{dryRun=true};if strings.HasPrefix(arg,"--source="){sourceName=strings.TrimPrefix(arg,"--source=")}}}
 if err:=config.LoadDotEnv(".env");err!=nil{return err};boards,err:=config.Load[config.Boards]("config/boards.json");if err!=nil{return err}
 switch command{case "web":return startWebServer(ctx,boards);case "run":profile,e:=config.Load[config.Profile]("config/profile.json");if e!=nil{return e};search,e:=config.Load[config.Search]("config/search.json");if e!=nil{return e};return runSearch(ctx,profile,search,boards,sourceName,dryRun);case "report":return latestReport();case "stats":return stats(ctx);case "--help","help":fmt.Println("jobsearch web");fmt.Println("jobsearch run [--dry-run] [--source=all|greenhouse|lever|remoteok|remotive|programathor|himalayas|gupy|adzuna|themuse|arbeitnow|mock]");fmt.Println("jobsearch report");fmt.Println("jobsearch stats");return nil;default:return fmt.Errorf("unknown command %q; use --help",command)}
}
func runSearch(ctx context.Context,profile config.Profile,search config.Search,boards config.Boards,sourceName string,dryRun bool)error{
 tokens:=[]string{};sites:=[]string{}
 for _,b:=range boards.Greenhouse{if b.Enabled&&strings.TrimSpace(b.Token)!=""{tokens=append(tokens,b.Token)}}
 for _,b:=range boards.Lever{if b.Enabled&&strings.TrimSpace(b.Site)!=""{sites=append(sites,b.Site)}}
 tokens=append(tokens,splitEnv("GREENHOUSE_BOARD_TOKENS")...);sites=append(sites,splitEnv("LEVER_SITES")...)
 tokens,sites=unique(tokens),unique(sites)
 profileSearch:=matching.BuildSearchProfile(profile); searchRoles:=profileSearch.ExpandedRoles; query:=sources.Query{Terms:profileSearch.DiscoveryTerms,Location:"",BoardTokens:tokens,LeverSites:sites}
 jobs,err:=fetchSources(ctx,sourceName,query)
 found:=len(jobs)
 if err!=nil{fmt.Fprintln(os.Stderr,"source warning:",err)}
 if len(jobs)==0&&err!=nil{return err}
 accepted:=[]domain.Job{};uncertain:=[]domain.Job{};senior:=[]domain.Job{};archived:=[]domain.Job{};rejected:=0;irrelevant:=0;stale:=0;outsideSeniority:=0;lowFit:=0
 freshCandidates:=make([]domain.Job,0,len(jobs))
 archivedCandidates:=make([]domain.Job,0,len(jobs))
 for i:=range jobs{
  if !isWithinDays(jobs[i],search.ArchiveDays){stale++;continue}
  if !matching.IsRelevant(jobs[i], searchRoles...) { irrelevant++; continue }
  if isFresh(jobs[i],search.FreshnessDays) { freshCandidates=append(freshCandidates,jobs[i]) } else { archivedCandidates=append(archivedCandidates,jobs[i]) }
 }
 process:=func(candidates []domain.Job, old bool){
  for i:=range candidates{
   job:=filter.Evaluate(candidates[i])
   matching.ScoreForRoles(&job,profile,profileSearch.Roles)
   if job.SeniorityMatch <= 55 {
    if !old && job.SeniorityMatch == 55 && job.FitScore >= search.MinimumFitScore && job.LocationEligible == domain.LocationEligible {
     job.RecommendationStatus=recommendation(job.FitScore,search.MinimumFitScore)
     senior=append(senior,job)
    } else { outsideSeniority++ }
    continue
   }
   job.RecommendationStatus=recommendation(job.FitScore,search.MinimumFitScore)
   if job.LocationEligible==domain.LocationEligible {
    if job.FitScore < search.MinimumFitScore { lowFit++; continue }
    if old { archived=append(archived,job) } else { accepted=append(accepted,job) }
   } else if job.LocationEligible==domain.LocationUnknown && !old {
    uncertain=append(uncertain,job)
   } else if job.LocationEligible==domain.LocationRejected { rejected++ }
  }
 }
 process(freshCandidates,false)
 process(archivedCandidates,true)
 accepted=rankAndDedupeJobs(accepted)
 senior=rankAndDedupeJobs(senior)
 archived=rankAndDedupeJobs(archived)
 accepted=limitJobsPerSource(accepted,search.MaxJobsPerSource)
 senior=limitJobsPerSource(senior,search.MaxJobsPerSource)
 archived=limitJobsPerSource(archived,search.MaxJobsPerSource)
 if !dryRun{if err:=os.MkdirAll("data",0755);err!=nil{return err};repo,err:=sqlite.Open(filepath.Join("data","jobsearch.db"));if err!=nil{return err};defer repo.Close();for i:=range accepted{if _,err:=repo.SaveJob(ctx,&accepted[i]);err!=nil{return err}};for i:=range senior{if _,err:=repo.SaveJob(ctx,&senior[i]);err!=nil{return err}};for i:=range archived{if _,err:=repo.SaveJob(ctx,&archived[i]);err!=nil{return err}};for i:=range uncertain{if _,err:=repo.SaveJob(ctx,&uncertain[i]);err!=nil{return err}}}
 now:=time.Now().Format("2006-01-02");if err:=os.MkdirAll("reports",0755);err!=nil{return err};if err:=os.WriteFile(filepath.Join("reports",now+".md"),[]byte(report.Markdown(accepted,senior,archived,uncertain,rejected,found,irrelevant,outsideSeniority,lowFit,stale,now)),0644);err!=nil{return err}
 data,err:=json.MarshalIndent(map[string]any{"generated_at":time.Now().UTC().Format(time.RFC3339),"jobs":accepted,"senior":senior,"archived":archived,"uncertain":uncertain},"","  ");if err!=nil{return err};if err:=os.WriteFile(filepath.Join("reports",now+".json"),data,0644);err!=nil{return err}
 fmt.Printf("Vagas encontradas: %d\nElegíveis: %d\nSenior ≥60: %d\nAntigas elegíveis: %d\nIncertas: %d\nRejeitadas por localização: %d\nFora do perfil: %d\nSenioridade fora do alvo: %d\nAbaixo do score mínimo: %d\nAntigas fora da janela: %d\n",found,len(accepted),len(senior),len(archived),len(uncertain),rejected,irrelevant,outsideSeniority,lowFit,stale);for _,j:=range accepted{fmt.Printf("%3d  %-45s  %s  %s\n",j.FitScore,j.Title,j.Company,j.URL)};return nil
}
type sourceResult struct {
	jobs []domain.Job
	err  error
}

func mergeSourceResults(results ...sourceResult) ([]domain.Job, error) {
	all := []domain.Job{}
	errs := []string{}
	failed := 0
	for _, result := range results {
		all = append(all, sanitizeJobs(result.jobs)...)
		if result.err != nil {
			failed++
			errs = append(errs, result.err.Error())
		}
	}
	if failed == len(results) && len(results) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}
	if len(errs) > 0 {
		return all, errors.New(strings.Join(errs, "; "))
	}
	return all, nil
}

func fetchSources(ctx context.Context,name string,q sources.Query)([]domain.Job,error){
 if name=="mock"{jobs,err:=(sources.Mock{}).FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="greenhouse"{if len(q.BoardTokens)==0{return nil,errors.New("no Greenhouse boards configured; add enabled boards to config/boards.json or set GREENHOUSE_BOARD_TOKENS")};jobs,err:=greenhouse.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="lever"{if len(q.LeverSites)==0{return nil,errors.New("no Lever sites configured; add enabled sites to config/boards.json or set LEVER_SITES")};jobs,err:=lever.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="remoteok"{jobs,err:=remoteok.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="remotive"{jobs,err:=remotive.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="programathor"{jobs,err:=programathor.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="himalayas"{jobs,err:=himalayas.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="gupy"{jobs,err:=gupy.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="adzuna"{jobs,err:=adzuna.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="themuse"{jobs,err:=themuse.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name=="arbeitnow"{jobs,err:=arbeitnow.NewClient().FetchJobs(ctx,q);return sanitizeJobs(jobs),err}
 if name!="all"{return nil,fmt.Errorf("unknown source %q",name)}

 sourcesToFetch:=[]sources.JobSource{
  remoteok.NewClient(),
  remotive.NewClient(),
  programathor.NewClient(),
  himalayas.NewClient(),
  gupy.NewClient(),
  themuse.NewClient(),
  arbeitnow.NewClient(),
 }
 results:=make([]sourceResult,0,len(sourcesToFetch)+2)
 var mu sync.Mutex
 var wg sync.WaitGroup
 if strings.TrimSpace(os.Getenv("ADZUNA_APP_ID"))!="" && strings.TrimSpace(os.Getenv("ADZUNA_APP_KEY"))!="" {
  sourcesToFetch=append(sourcesToFetch,adzuna.NewClient())
 }
 for _,source:=range sourcesToFetch {
  source:=source
  wg.Add(1)
  go func(){
   defer wg.Done()
   jobs,err:=source.FetchJobs(ctx,q)
   mu.Lock()
   results=append(results,sourceResult{jobs:sanitizeJobs(jobs),err:err})
   mu.Unlock()
  }()
 }
 if len(q.BoardTokens)>0 {
  wg.Add(1)
  go func(){
   defer wg.Done()
   jobs,err:=greenhouse.NewClient().FetchJobs(ctx,q)
   mu.Lock()
   results=append(results,sourceResult{jobs:sanitizeJobs(jobs),err:err})
   mu.Unlock()
  }()
 }
 if len(q.LeverSites)>0 {
  wg.Add(1)
  go func(){
   defer wg.Done()
   jobs,err:=lever.NewClient().FetchJobs(ctx,q)
   mu.Lock()
   results=append(results,sourceResult{jobs:sanitizeJobs(jobs),err:err})
   mu.Unlock()
  }()
 }
 wg.Wait()
 jobs,err:=mergeSourceResults(results...)
 if err!=nil && len(jobs)>0{return jobs,err}
 return jobs,err
}

func sanitizeJobs(jobs []domain.Job) []domain.Job {
	out := make([]domain.Job, 0, len(jobs))
	seen := map[string]struct{}{}
	for _, job := range jobs {
		job.Source = strings.TrimSpace(job.Source)
		job.Title = strings.TrimSpace(job.Title)
		job.Company = strings.TrimSpace(job.Company)
		job.URL = strings.TrimSpace(job.URL)
		job.ApplyURL = strings.TrimSpace(job.ApplyURL)
		if job.URL == "" {
			job.URL = job.ApplyURL
		}
		if job.ApplyURL == "" {
			job.ApplyURL = job.URL
		}
		if job.Source == "" || job.Title == "" || job.URL == "" {
			continue
		}
		if job.CanonicalURL == "" {
			job.CanonicalURL = dedupe.CanonicalURL(job.URL)
		}
		if job.ID == "" {
			job.ID = job.Source + ":" + job.CanonicalURL
		}
		key := job.Source + "|" + job.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, job)
	}
	return out
}

func splitEnv(key string)[]string{v:=strings.TrimSpace(os.Getenv(key));if v==""{return nil};return strings.Split(v,",")}
func isWithinDays(j domain.Job, days int) bool { if days <= 0 { return true }; var newest time.Time; for _, stamp := range []string{j.UpdatedAt, j.PostedAt} { if parsed := parseJobTime(stamp); !parsed.IsZero() && parsed.After(newest) { newest = parsed } }; if newest.IsZero() { return true }; cutoff := time.Now().Add(-time.Duration(days)*24*time.Hour); return !newest.Before(cutoff) }

func isFresh(j domain.Job, days int) bool {
	if days <= 0 { return true }
	var newest time.Time
	for _, stamp := range []string{j.UpdatedAt, j.PostedAt} {
		if parsed := parseJobTime(stamp); !parsed.IsZero() && parsed.After(newest) {
			newest = parsed
		}
	}
	if newest.IsZero() {
		return true
	}
	cutoff := time.Now().Add(-time.Duration(days)*24*time.Hour)
	return !newest.Before(cutoff)
}

func parseJobTime(value string) time.Time {
	value = strings.TrimSpace(value)
	layouts := []string{
		time.RFC3339, time.RFC3339Nano,
		"2006-01-02", "2006-01-02 15:04:05", "2006-01-02 15:04:05 -0700 MST",
		"Mon, 02 Jan 2006 15:04:05 MST", "Mon, 02 Jan 2006 15:04:05 -0700",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil { return t }
	}
	return time.Time{}
}

func recommendation(score,minimum int)domain.RecommendationStatus{if score>=80{return domain.Recommended};if score>=minimum{return domain.PossibleMatch};return domain.LowMatch}
func latestReport()error{entries,err:=os.ReadDir("reports");if err!=nil{return err};names:=[]string{};for _,e:=range entries{if !e.IsDir()&&strings.HasSuffix(e.Name(),".md"){names=append(names,e.Name())}};if len(names)==0{return errors.New("no reports found")};sort.Strings(names);b,err:=os.ReadFile(filepath.Join("reports",names[len(names)-1]));if err!=nil{return err};fmt.Print(string(b));return nil}
func stats(ctx context.Context)error{repo,err:=sqlite.Open(filepath.Join("data","jobsearch.db"));if err!=nil{return err};defer repo.Close();s,err:=repo.GetStats(ctx);if err!=nil{return err};fmt.Printf("Total: %d\nNovas: %d\nVistas: %d\nAtualizadas: %d\nExpiradas: %d\nRecomendadas: %d\nPossíveis: %d\n",s.Total,s.New,s.Seen,s.Updated,s.Expired,s.Recommended,s.Possible);return nil}
func unique(values []string)[]string{seen:=map[string]struct{}{};out:=[]string{};for _,v:=range values{k:=strings.TrimSpace(v);if k==""{continue};if _,ok:=seen[k];ok{continue};seen[k]=struct{}{};out=append(out,k)};return out}


func rankAndDedupeJobs(jobs []domain.Job) []domain.Job {
	sort.SliceStable(jobs, func(i, j int) bool { return matching.BetterMatch(jobs[i], jobs[j]) })
	return dedupe.Jobs(jobs)
}

func limitJobsPerSource(jobs []domain.Job, limit int) []domain.Job {
	if limit <= 0 {
		return jobs
	}
	counts := map[string]int{}
	out := make([]domain.Job, 0, len(jobs))
	for _, job := range jobs {
		source := strings.ToLower(strings.TrimSpace(job.Source))
		if counts[source] >= limit {
			continue
		}
		counts[source]++
		out = append(out, job)
	}
	return out
}
