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
body{font-family:system-ui,sans-serif;max-width:900px;margin:40px auto;padding:0 20px;background:#f7f7f8;color:#202124}.card{background:white;border:1px solid #ddd;border-radius:14px;padding:24px;margin-bottom:18px}label{display:block;font-weight:600;margin:14px 0 6px}input,select{width:100%;box-sizing:border-box;padding:11px;border:1px solid #bbb;border-radius:8px}.filters{display:grid;grid-template-columns:1.5fr 1fr 1fr;gap:12px}.filters label{margin:0}.summary{display:flex;gap:10px;flex-wrap:wrap;margin:14px 0}.summary .pill{font-size:13px}.job{border-top:1px solid #eee;padding:16px 0}.score{font-size:22px;font-weight:800}.muted{color:#666}.hidden{display:none}.pill{display:inline-block;padding:4px 8px;border-radius:999px;background:#eee;font-size:12px;margin:2px}.pill.strong{background:#dff5e5}.pill.compatible{background:#e7f0ff}.pill.possible{background:#fff3d6}.pill.low{background:#f4dddd}
@media(max-width:600px){body{margin:15px auto}.card{padding:18px}.score{font-size:18px}.filters{grid-template-columns:1fr}}
</style></head><body>
<div class="card"><h1>JobSearcher</h1><p>Encontre vagas compatíveis com seu perfil sem enviar ou armazenar seu currículo.</p>
<form method="post">
<label>Cargo / área</label><input name="roles" placeholder="Backend Developer, Software Engineer" value="{{.Roles}}">
<label>Skills</label><input name="skills" placeholder="Go, AWS, PostgreSQL, Docker" value="{{.Skills}}">
<label>Senioridade</label><select name="seniority"><option value="junior" {{if eq .Seniority "junior"}}selected{{end}}>Júnior</option><option value="mid" {{if eq .Seniority "mid"}}selected{{end}}>Pleno</option><option value="senior" {{if eq .Seniority "senior"}}selected{{end}}>Sênior</option><option value="staff" {{if eq .Seniority "staff"}}selected{{end}}>Staff / Lead</option></select>
<label>Anos de experiência</label><input type="number" min="0" name="experience" value="{{.Experience}}">
<p class="muted">🌎 Busca global — somente vagas 100% remotas.</p>
<button type="submit">🔎 Buscar vagas</button></form></div>
{{if .Searched}}<div class="card"><h2>Resultados</h2><div class="summary"><span class="pill"><strong id="visibleCount">{{.Count}}</strong> visíveis</span><span class="pill">{{.Count}} elegíveis</span></div>
<div class="filters">
<label>Filtrar resultados<input id="jobFilter" type="search" placeholder="Cargo ou empresa"></label>
<label>Compatibilidade<select id="bucketFilter"><option value="all">Todas</option><option value="strong">Forte</option><option value="compatible">Compatível</option><option value="possible">Possível</option></select></label>
<label>Fonte<select id="sourceFilter"><option value="all">Todas</option>{{range .Jobs}}<option value="{{.Source}}">{{.Source}}</option>{{end}}</select></label>
</div>
<div id="jobsList">{{range .Jobs}}<div class="job job-card" data-bucket="{{.MatchBucket}}" data-source="{{.Source}}" data-search="{{.Title}} {{.Company}}">
<div class="score">{{.FitScore}}% — {{.Title}}</div><div class="skills"><span class="pill {{.MatchBucket}}">Compatibilidade {{.MatchBucket}}</span></div>
<strong>{{.Company}}</strong><div class="muted">{{.Location}} · {{.WorkplaceType}} · {{.Seniority}}</div>
<div class="skills"><span class="pill">Cargo {{.RoleMatch}}%</span><span class="pill">Skills {{.SkillMatch}}%</span><span class="pill">Relacionadas {{.RelatedSkillMatch}}%</span><span class="pill">Técnico {{.TechnicalMatch}}%</span><span class="pill">Responsabilidades {{.ResponsibilityMatch}}%</span><span class="pill">Senioridade {{.SeniorityMatch}}%</span><span class="pill">Experiência {{.ExperienceMatch}}%</span></div>
<div class="skills"><strong>Por que combina</strong><ul>{{range .MatchHighlights}}<li>{{.}}</li>{{end}}</ul></div>
<div class="skills">✓ {{join .MustHaveMatch}} {{if .MustHaveMissing}} · △ {{join .MustHaveMissing}}{{end}}</div>
<div class="skills"><span class="pill">{{.Source}}</span><span class="pill">{{.RecommendationStatus}}</span></div>
<p><a href="{{.URL}}" target="_blank" rel="noopener">Ver vaga →</a></p></div>{{else}}<p>Nenhuma vaga encontrada com esses critérios.</p>{{end}}</div>
<p id="emptyFiltered" class="muted hidden">Nenhuma vaga corresponde aos filtros selecionados.</p>
</div>
<script>
(function(){
 const text=document.getElementById('jobFilter'), bucket=document.getElementById('bucketFilter'), source=document.getElementById('sourceFilter');
 const cards=[...document.querySelectorAll('.job-card')], count=document.getElementById('visibleCount'), empty=document.getElementById('emptyFiltered');
 const sources=[...source.options].map(o=>o.value).filter(v=>v!=='all').sort((a,b)=>a.localeCompare(b));
 const seen=new Set(); source.innerHTML='<option value="all">Todas</option>'+sources.filter(v=>!seen.has(v)&&seen.add(v)).map(v=>'<option></option>').join('');
 [...source.options].forEach((o,i)=>{if(i>0)o.value=sources[i-1];o.textContent=i===0?'Todas':sources[i-1]});
 function apply(){const q=text.value.trim().toLowerCase(), b=bucket.value, s=source.value;let visible=0;
  cards.forEach(c=>{const ok=(!q||c.dataset.search.toLowerCase().includes(q))&&(b==='all'||c.dataset.bucket===b)&&(s==='all'||c.dataset.source===s);c.classList.toggle('hidden',!ok);if(ok)visible++});
  count.textContent=visible;empty.classList.toggle('hidden',visible!==0);
 }
 [text,bucket,source].forEach(el=>el.addEventListener('input',apply));apply();
})();
</script>{{end}}
</body></html>`

type webView struct { Roles, Skills, Experience, Seniority string; Searched bool; Count int; Jobs []domain.Job }

func startWebServer(ctx context.Context, boards config.Boards) error {
 mux:=http.NewServeMux()
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
  v:=webView{Experience:"3",Seniority:"junior"}
  if r.Method==http.MethodPost {
   _=r.ParseForm(); v.Roles=strings.TrimSpace(r.FormValue("roles")); v.Skills=strings.TrimSpace(r.FormValue("skills")); v.Experience=r.FormValue("experience"); v.Seniority=r.FormValue("seniority")
   years,_:=strconv.Atoi(v.Experience); if years<0 {years=0}
   roles:=csvValues(v.Roles); skills:=csvValues(v.Skills); if len(roles)==0 {http.Error(w,"informe pelo menos um cargo ou área",400);return}; if len(skills)==0 {http.Error(w,"informe pelo menos uma skill",400);return}
   profile:=config.Profile{TargetSeniority:v.Seniority,Titles:roles,Technologies:skills,YearsExperience:years,Weights:config.Weights{Technical:35,Responsibility:15,Seniority:25,Cloud:10,Domain:7,Language:4,AI:4}}
   search:=config.Search{Location:"",RemoteAllowed:[]string{"Worldwide"},PreferredTitles:roles,MinimumFitScore:55,FreshnessDays:7,ArchiveDays:30,MaxJobsPerSource:100}
   jobs,err:=searchForWeb(ctx,profile,search,boards,roles); if err!=nil {http.Error(w,"erro ao buscar vagas: "+err.Error(),http.StatusBadGateway);return}; v.Searched=true;v.Jobs=jobs;v.Count=len(jobs)
  }
  t,err:=template.New("page").Funcs(template.FuncMap{"join":func(v []string)string{return strings.Join(v,", ")}}).Parse(webPage);if err!=nil{http.Error(w,err.Error(),500);return};_=t.Execute(w,v)
 })
 fmt.Println("JobSearcher web: http://localhost:8080")
 return http.ListenAndServe(":8080",mux)
}

func searchForWeb(ctx context.Context,profile config.Profile,search config.Search,boards config.Boards,requestedRoles []string)([]domain.Job,error){
 tokens:=[]string{};sites:=[]string{}
 for _,b:=range boards.Greenhouse {if b.Enabled&&strings.TrimSpace(b.Token)!=""{tokens=append(tokens,b.Token)}}
 for _,b:=range boards.Lever {if b.Enabled&&strings.TrimSpace(b.Site)!=""{sites=append(sites,b.Site)}}
 searchProfile:=matching.BuildSearchProfile(profile); jobs,err:=fetchSources(ctx,"all",sources.Query{Terms:searchProfile.DiscoveryTerms,Location:"",BoardTokens:tokens,LeverSites:sites});if err!=nil&&len(jobs)==0{return nil,err}
 out:=[]domain.Job{}
 for _,j:=range jobs {if !isWithinDays(j,search.ArchiveDays)||!matching.IsRelevant(j,searchProfile.ExpandedRoles...)||!isFresh(j,search.FreshnessDays){continue}; x:=filter.Evaluate(j);matching.ScoreForRoles(&x,profile,searchProfile.Roles);if x.WorkplaceType=="remote"&&x.FitScore>=search.MinimumFitScore&&x.SeniorityMatch>20{out=append(out,x)}}
 out=dedupe.Jobs(out)
 sort.SliceStable(out,func(i,j int)bool{return matching.BetterMatch(out[i],out[j])})
 return limitJobsPerSource(out,search.MaxJobsPerSource),nil
}
func csvValues(s string)[]string{out:=[]string{};for _,v:=range strings.Split(s,","){if v=strings.TrimSpace(v);v!=""{out=append(out,v)}};return out}

func workModelAllowed(j domain.Job,remote,hybrid,onsite bool) bool { if !remote&&!hybrid&&!onsite{return true}; text:=strings.ToLower(j.WorkplaceType+" "+j.Location+" "+j.Description[:minLen(len(j.Description),500)]); ok:=false; if remote&&(strings.Contains(text,"remote")||strings.Contains(text,"remoto")){ok=true}; if hybrid&&(strings.Contains(text,"hybrid")||strings.Contains(text,"híbrido")||strings.Contains(text,"hibrido")){ok=true}; if onsite&&(strings.Contains(text,"onsite")||strings.Contains(text,"on-site")||strings.Contains(text,"presencial")){ok=true}; return ok }
func minLen(a,b int) int {if a<b{return a};return b}
