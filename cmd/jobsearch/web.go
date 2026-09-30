package main

import (
 "context"
 "fmt"
 "html/template"
 "net/http"
 "sort"
 "strconv"
 "strings"

 "jobsearcher/internal/config"
 "jobsearcher/internal/dedupe"
 "jobsearcher/internal/domain"
 "jobsearcher/internal/filter"
 "jobsearcher/internal/matching"
 "jobsearcher/internal/sources"
)

const webPage = `<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>JobSearcher</title>
<style>
body{font-family:system-ui,sans-serif;max-width:900px;margin:40px auto;padding:0 20px;background:#f7f7f8;color:#202124}.card{background:white;border:1px solid #ddd;border-radius:14px;padding:24px;margin-bottom:18px}label{display:block;font-weight:600;margin:14px 0 6px}input,select{width:100%;box-sizing:border-box;padding:11px;border:1px solid #bbb;border-radius:8px}.checks{display:flex;gap:18px;flex-wrap:wrap}.checks label{font-weight:400;margin:8px 0}.checks input{width:auto}button{margin-top:20px;padding:12px 18px;border:0;border-radius:9px;background:#111;color:#fff;font-weight:700;cursor:pointer}.job{border-top:1px solid #eee;padding:16px 0}.score{font-size:22px;font-weight:800}.muted{color:#666}
@media(max-width:600px){body{margin:15px auto}.card{padding:18px}.score{font-size:18px}}.pill{display:inline-block;padding:4px 8px;border-radius:999px;background:#eee;font-size:12px;margin:2px}</style></head><body>
<div class="card"><h1>JobSearcher</h1><p>Encontre vagas compatíveis com seu perfil sem enviar ou armazenar seu currículo.</p>
<form method="post">
<label>Cargo / área</label><input name="roles" placeholder="Backend Developer, Software Engineer" value="{{.Roles}}">
<label>Skills</label><input name="skills" placeholder="Go, AWS, PostgreSQL, Docker" value="{{.Skills}}">
<label>Senioridade</label><select name="seniority"><option value="junior" {{if eq .Seniority "junior"}}selected{{end}}>Júnior</option><option value="mid" {{if eq .Seniority "mid"}}selected{{end}}>Pleno</option><option value="senior" {{if eq .Seniority "senior"}}selected{{end}}>Sênior</option><option value="staff" {{if eq .Seniority "staff"}}selected{{end}}>Staff / Lead</option></select>
<label>Anos de experiência</label><input type="number" min="0" name="experience" value="{{.Experience}}">
<label>Localização</label><input name="location" placeholder="Brasil, São Paulo, Maceió..." value="{{.Location}}">
<label>Modelo de trabalho</label><div class="checks"><label><input type="checkbox" name="remote" checked> Remoto</label><label><input type="checkbox" name="hybrid" checked> Híbrido</label><label><input type="checkbox" name="onsite"> Presencial</label></div>
<button type="submit">🔎 Buscar vagas</button></form></div>
{{if .Searched}}<div class="card"><h2>Resultados</h2><p class="muted">{{.Count}} vagas elegíveis encontradas.</p>{{range .Jobs}}<div class="job"><div class="score">{{.FitScore}}% — {{.Title}}</div><strong>{{.Company}}</strong><div class="muted">{{.Location}} · {{.WorkplaceType}} · {{.Seniority}}</div><div class="skills">✓ {{join .MustHaveMatch}} {{if .MustHaveMissing}} · △ {{join .MustHaveMissing}}{{end}}</div><div class="skills"><span class="pill">{{.Source}}</span><span class="pill">{{.RecommendationStatus}}</span></div><p><a href="{{.URL}}" target="_blank" rel="noopener">Ver vaga →</a></p></div>{{else}}<p>Nenhuma vaga encontrada com esses critérios.</p>{{end}}</div>{{end}}
</body></html>`

type webView struct { Roles, Skills, Experience, Location, Seniority string; Searched bool; Count int; Jobs []domain.Job }

func startWebServer(ctx context.Context, boards config.Boards) error {
 mux:=http.NewServeMux()
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
  v:=webView{Experience:"3",Location:"Brazil",Seniority:"junior"}
  if r.Method==http.MethodPost {
   _=r.ParseForm(); v.Roles=strings.TrimSpace(r.FormValue("roles")); v.Skills=strings.TrimSpace(r.FormValue("skills")); v.Experience=r.FormValue("experience"); v.Location=strings.TrimSpace(r.FormValue("location")); v.Seniority=r.FormValue("seniority")
   years,_:=strconv.Atoi(v.Experience); if years<0 {years=0}
   roles:=csvValues(v.Roles); skills:=csvValues(v.Skills); if len(roles)==0 {http.Error(w,"informe pelo menos um cargo ou área",400);return}; if len(skills)==0 {http.Error(w,"informe pelo menos uma skill",400);return}
   profile:=config.Profile{Titles:roles,Technologies:skills,YearsExperience:years,Weights:config.Weights{Technical:35,Responsibility:15,Seniority:25,Cloud:10,Domain:7,Language:4,AI:4}}
   search:=config.Search{Location:v.Location,RemoteAllowed:[]string{"Brazil","LATAM","South America","Worldwide","Americas"},PreferredTitles:roles,MinimumFitScore:60,FreshnessDays:7,ArchiveDays:30,MaxJobsPerSource:100}
   jobs,err:=searchForWeb(ctx,profile,search,boards,roles,r.FormValue("remote")!="" ,r.FormValue("hybrid")!="" ,r.FormValue("onsite")!=""); if err!=nil {http.Error(w,"erro ao buscar vagas: "+err.Error(),http.StatusBadGateway);return}; v.Searched=true;v.Jobs=jobs;v.Count=len(jobs)
  }
  t,err:=template.New("page").Funcs(template.FuncMap{"join":func(v []string)string{return strings.Join(v,", ")}}).Parse(webPage);if err!=nil{http.Error(w,err.Error(),500);return};_=t.Execute(w,v)
 })
 fmt.Println("JobSearcher web: http://localhost:8080")
 return http.ListenAndServe(":8080",mux)
}

func searchForWeb(ctx context.Context,profile config.Profile,search config.Search,boards config.Boards,requestedRoles []string,remote,hybrid,onsite bool)([]domain.Job,error){
 tokens:=[]string{};sites:=[]string{}
 for _,b:=range boards.Greenhouse {if b.Enabled&&strings.TrimSpace(b.Token)!=""{tokens=append(tokens,b.Token)}}
 for _,b:=range boards.Lever {if b.Enabled&&strings.TrimSpace(b.Site)!=""{sites=append(sites,b.Site)}}
 jobs,err:=fetchSources(ctx,"all",sources.Query{Terms:search.PreferredTitles,Location:search.Location,BoardTokens:tokens,LeverSites:sites});if err!=nil&&len(jobs)==0{return nil,err}
 out:=[]domain.Job{}
 for _,j:=range jobs {if !isWithinDays(j,search.ArchiveDays)||!matching.IsRelevant(j,requestedRoles)||!isFresh(j,search.FreshnessDays){continue}; x:=filter.Evaluate(j);matching.Score(&x,profile);if x.LocationEligible==domain.LocationEligible&&x.FitScore>=search.MinimumFitScore&&x.SeniorityMatch>55&&workModelAllowed(x,remote,hybrid,onsite){out=append(out,x)}}
 sort.SliceStable(out,func(i,j int)bool{return out[i].FitScore>out[j].FitScore})
 return dedupe.Jobs(out),nil
}
func csvValues(s string)[]string{out:=[]string{};for _,v:=range strings.Split(s,","){if v=strings.TrimSpace(v);v!=""{out=append(out,v)}};return out}

func workModelAllowed(j domain.Job,remote,hybrid,onsite bool) bool { if !remote&&!hybrid&&!onsite{return true}; text:=strings.ToLower(j.WorkplaceType+" "+j.Location+" "+j.Description[:minLen(len(j.Description),500)]); ok:=false; if remote&&(strings.Contains(text,"remote")||strings.Contains(text,"remoto")){ok=true}; if hybrid&&(strings.Contains(text,"hybrid")||strings.Contains(text,"híbrido")||strings.Contains(text,"hibrido")){ok=true}; if onsite&&(strings.Contains(text,"onsite")||strings.Contains(text,"on-site")||strings.Contains(text,"presencial")){ok=true}; return ok }
func minLen(a,b int) int {if a<b{return a};return b}
