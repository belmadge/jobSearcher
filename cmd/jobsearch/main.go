package main

import (
 "context";"encoding/json";"errors";"fmt";"os";"path/filepath";"sort";"strings";"time"
 "jobsearcher/internal/config";"jobsearcher/internal/dedupe";"jobsearcher/internal/domain";"jobsearcher/internal/filter";"jobsearcher/internal/matching";"jobsearcher/internal/report";"jobsearcher/internal/sources";"jobsearcher/internal/sources/greenhouse";"jobsearcher/internal/sources/lever";"jobsearcher/internal/sources/remoteok";"jobsearcher/internal/sources/remotive";"jobsearcher/internal/storage/sqlite"
)
func main(){if err:=run(context.Background(),os.Args[1:]);err!=nil{fmt.Fprintln(os.Stderr,"error:",err);os.Exit(1)}}
func run(ctx context.Context,args []string)error{
 command,sourceName,dryRun:="run","all",false
 if len(args)>0{command=args[0];for _,arg:=range args[1:]{if arg=="--dry-run"{dryRun=true};if strings.HasPrefix(arg,"--source="){sourceName=strings.TrimPrefix(arg,"--source=")}}}
 if err:=config.LoadDotEnv(".env");err!=nil{return err};profile,err:=config.Load[config.Profile]("config/profile.json");if err!=nil{return err};search,err:=config.Load[config.Search]("config/search.json");if err!=nil{return err};boards,err:=config.Load[config.Boards]("config/boards.json");if err!=nil{return err}
 switch command{case "run":return runSearch(ctx,profile,search,boards,sourceName,dryRun);case "report":return latestReport();case "stats":return stats(ctx);case "--help","help":fmt.Println("jobsearch run [--dry-run] [--source=all|greenhouse|lever|remoteok|remotive|mock]");fmt.Println("jobsearch report");fmt.Println("jobsearch stats");return nil;default:return fmt.Errorf("unknown command %q; use --help",command)}
}
func runSearch(ctx context.Context,profile config.Profile,search config.Search,boards config.Boards,sourceName string,dryRun bool)error{
 tokens:=[]string{};sites:=[]string{}
 for _,b:=range boards.Greenhouse{if b.Enabled&&strings.TrimSpace(b.Token)!=""{tokens=append(tokens,b.Token)}}
 for _,b:=range boards.Lever{if b.Enabled&&strings.TrimSpace(b.Site)!=""{sites=append(sites,b.Site)}}
 tokens=append(tokens,splitEnv("GREENHOUSE_BOARD_TOKENS")...);sites=append(sites,splitEnv("LEVER_SITES")...)
 tokens,sites=unique(tokens),unique(sites)
 query:=sources.Query{Terms:search.PreferredTitles,Location:search.Location,BoardTokens:tokens,LeverSites:sites}
 jobs,err:=fetchSources(ctx,sourceName,query)
 if err!=nil{fmt.Fprintln(os.Stderr,"source warning:",err)}
 if len(jobs)==0&&err!=nil{return err}
 accepted:=[]domain.Job{};uncertain:=[]domain.Job{};rejected:=0;irrelevant:=0
 for i:=range jobs{
  if !matching.IsRelevant(jobs[i]) { irrelevant++; continue }
  job:=filter.Evaluate(jobs[i]);switch job.LocationEligible{case domain.LocationEligible:matching.Score(&job,profile);job.RecommendationStatus=recommendation(job.FitScore,search.MinimumFitScore);accepted=append(accepted,job);case domain.LocationUnknown:job.RecommendationStatus=domain.UncertainLocation;uncertain=append(uncertain,job);default:job.RecommendationStatus=domain.RejectedLocation;rejected++}}
 accepted=dedupe.Jobs(accepted);sort.SliceStable(accepted,func(i,j int)bool{return accepted[i].FitScore>accepted[j].FitScore})
 if !dryRun{if err:=os.MkdirAll("data",0755);err!=nil{return err};repo,err:=sqlite.Open(filepath.Join("data","jobsearch.db"));if err!=nil{return err};defer repo.Close();for i:=range accepted{if _,err:=repo.SaveJob(ctx,&accepted[i]);err!=nil{return err}};for i:=range uncertain{if _,err:=repo.SaveJob(ctx,&uncertain[i]);err!=nil{return err}}}
 now:=time.Now().Format("2006-01-02");if err:=os.MkdirAll("reports",0755);err!=nil{return err};if err:=os.WriteFile(filepath.Join("reports",now+".md"),[]byte(report.Markdown(accepted,uncertain,rejected,len(jobs),now)),0644);err!=nil{return err}
 data,err:=json.MarshalIndent(map[string]any{"generated_at":time.Now().UTC().Format(time.RFC3339),"jobs":accepted,"uncertain":uncertain},"","  ");if err!=nil{return err};if err:=os.WriteFile(filepath.Join("reports",now+".json"),data,0644);err!=nil{return err}
 fmt.Printf("Vagas encontradas: %d\nElegíveis: %d\nIncertas: %d\nRejeitadas por localização: %d\nFora do perfil: %d\n",len(jobs),len(accepted),len(uncertain),rejected,irrelevant);for _,j:=range accepted{fmt.Printf("%3d  %-45s  %s  %s\n",j.FitScore,j.Title,j.Company,j.URL)};return nil
}
func fetchSources(ctx context.Context,name string,q sources.Query)([]domain.Job,error){
 if name=="mock"{return (sources.Mock{}).FetchJobs(ctx,q)}
 if name=="greenhouse"{if len(q.BoardTokens)==0{return nil,errors.New("no Greenhouse boards configured; add enabled boards to config/boards.json or set GREENHOUSE_BOARD_TOKENS")};return greenhouse.NewClient().FetchJobs(ctx,q)}
 if name=="lever"{if len(q.LeverSites)==0{return nil,errors.New("no Lever sites configured; add enabled sites to config/boards.json or set LEVER_SITES")};return lever.NewClient().FetchJobs(ctx,q)}
 if name=="remoteok"{return remoteok.NewClient().FetchJobs(ctx,q)}
 if name=="remotive"{return remotive.NewClient().FetchJobs(ctx,q)}
 if name!="all"{return nil,fmt.Errorf("unknown source %q",name)}
 if len(q.BoardTokens)==0&&len(q.LeverSites)==0{j1,e1:=remoteok.NewClient().FetchJobs(ctx,q);j2,e2:=remotive.NewClient().FetchJobs(ctx,q);if e1!=nil&&e2!=nil{return nil,fmt.Errorf("remote sources failed: %v; %v",e1,e2)};return append(j1,j2...),nil}
 all:=[]domain.Job{};errs:=[]string{}
 if len(q.BoardTokens)>0{j,e:=greenhouse.NewClient().FetchJobs(ctx,q);all=append(all,j...);if e!=nil{errs=append(errs,"greenhouse: "+e.Error())}}
 if len(q.LeverSites)>0{j,e:=lever.NewClient().FetchJobs(ctx,q);all=append(all,j...);if e!=nil{errs=append(errs,"lever: "+e.Error())}}
 j,e:=remoteok.NewClient().FetchJobs(ctx,q);all=append(all,j...);if e!=nil{errs=append(errs,"remoteok: "+e.Error())}
 j,e=remotive.NewClient().FetchJobs(ctx,q);all=append(all,j...);if e!=nil{errs=append(errs,"remotive: "+e.Error())}
 if len(errs)>0{return all,errors.New(strings.Join(errs,"; "))}
 return all,nil
}
func splitEnv(key string)[]string{v:=strings.TrimSpace(os.Getenv(key));if v==""{return nil};return strings.Split(v,",")}
func recommendation(score,minimum int)domain.RecommendationStatus{if score>=80{return domain.Recommended};if score>=minimum{return domain.PossibleMatch};return domain.LowMatch}
func latestReport()error{entries,err:=os.ReadDir("reports");if err!=nil{return err};names:=[]string{};for _,e:=range entries{if !e.IsDir()&&strings.HasSuffix(e.Name(),".md"){names=append(names,e.Name())}};if len(names)==0{return errors.New("no reports found")};sort.Strings(names);b,err:=os.ReadFile(filepath.Join("reports",names[len(names)-1]));if err!=nil{return err};fmt.Print(string(b));return nil}
func stats(ctx context.Context)error{repo,err:=sqlite.Open(filepath.Join("data","jobsearch.db"));if err!=nil{return err};defer repo.Close();s,err:=repo.GetStats(ctx);if err!=nil{return err};fmt.Printf("Total: %d\nNovas: %d\nVistas: %d\nAtualizadas: %d\nExpiradas: %d\nRecomendadas: %d\nPossíveis: %d\n",s.Total,s.New,s.Seen,s.Updated,s.Expired,s.Recommended,s.Possible);return nil}
func unique(values []string)[]string{seen:=map[string]struct{}{};out:=[]string{};for _,v:=range values{k:=strings.TrimSpace(v);if k==""{continue};if _,ok:=seen[k];ok{continue};seen[k]=struct{}{};out=append(out,k)};return out}
