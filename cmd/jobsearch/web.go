package main

import (
 "context"
 "fmt"
 "html/template"
 "net/http"
 "strconv"
 "strings"

 "jobsearcher/internal/config"
 "jobsearcher/internal/domain"
 "jobsearcher/internal/filter"
 "jobsearcher/internal/matching"
 "jobsearcher/internal/sources"
)

const webPage = `<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>JobSearcher — encontre sua próxima vaga</title>
<style>
:root{--bg:#f7f5fb;--surface:#fff;--ink:#181622;--muted:#706b7c;--line:#e8e3ef;--primary:#6d4aff;--primary-dark:#5233d7;--shadow:0 12px 35px rgba(42,28,75,.08);--radius:22px}
*{box-sizing:border-box}html{scroll-behavior:smooth}
body{margin:0;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;background:radial-gradient(circle at 5% 0%,#eee9ff 0,transparent 28rem),radial-gradient(circle at 100% 10%,#f3e9ff 0,transparent 25rem),var(--bg);color:var(--ink)}
.container{width:min(1180px,calc(100% - 32px));margin:0 auto;padding:38px 0 70px}
.hero{position:relative;overflow:hidden;background:linear-gradient(135deg,#fff 0%,#faf8ff 100%);border:1px solid var(--line);border-radius:30px;padding:42px;box-shadow:var(--shadow);margin-bottom:24px}
.hero:after{content:"";position:absolute;width:230px;height:230px;border-radius:50%;right:-80px;top:-100px;background:#e8e1ff;opacity:.8}
.brand{display:flex;align-items:center;gap:12px;margin-bottom:14px;position:relative;z-index:1}.logo{width:44px;height:44px;border-radius:14px;display:grid;place-items:center;background:#181622;color:#fff;font-weight:900;box-shadow:0 8px 18px rgba(24,22,34,.18)}
h1{font-size:clamp(2rem,5vw,3.35rem);letter-spacing:-.055em;line-height:1.02;margin:0;position:relative;z-index:1}.hero-copy{max-width:650px;color:var(--muted);font-size:1.05rem;line-height:1.65;margin:16px 0 28px;position:relative;z-index:1}
.form-grid{display:grid;grid-template-columns:1.25fr 1fr .65fr;gap:16px;position:relative;z-index:1}.field{min-width:0}.field label{display:block;font-weight:750;font-size:.9rem;margin:0 0 8px}
input,select{width:100%;height:48px;border:1px solid #d9d4e2;border-radius:13px;background:#fff;padding:0 14px;color:var(--ink);font:inherit;outline:none;transition:.2s}input:focus,select:focus{border-color:#9a84ff;box-shadow:0 0 0 4px #eeeaff}.full{grid-column:1/-1}
.search-row{display:flex;align-items:center;justify-content:space-between;gap:18px;margin-top:18px}.remote-note{color:var(--muted);font-size:.92rem;display:flex;align-items:center;gap:7px}
button{border:0;border-radius:13px;background:var(--primary);color:#fff;font:inherit;font-weight:800;padding:13px 20px;cursor:pointer;box-shadow:0 8px 18px rgba(109,74,255,.24);transition:.2s}button:hover{background:var(--primary-dark);transform:translateY(-1px)}button:disabled{opacity:.7;cursor:wait;transform:none}
.error{background:#fff1f1;border:1px solid #efc3c3;color:#8e2525;border-radius:13px;padding:12px 14px;margin-top:16px}
.section{margin-top:30px}.section-head{display:flex;align-items:end;justify-content:space-between;gap:20px;margin-bottom:18px}.section-head h2{font-size:1.7rem;letter-spacing:-.035em;margin:0}.section-head p{margin:5px 0 0;color:var(--muted)}
.summary{display:flex;gap:8px;flex-wrap:wrap}.summary .stat{background:#fff;border:1px solid var(--line);border-radius:999px;padding:8px 12px;font-size:.82rem;color:var(--muted)}.summary strong{color:var(--ink)}
.toolbar{background:rgba(255,255,255,.78);backdrop-filter:blur(12px);border:1px solid var(--line);border-radius:18px;padding:14px;margin-bottom:20px;display:grid;grid-template-columns:1.6fr .8fr .8fr;gap:12px}.toolbar label{font-size:.78rem;font-weight:750;color:var(--muted)}.toolbar input,.toolbar select{height:42px;margin-top:6px}
.jobs-grid{columns:2 420px;column-gap:18px}.job-card{display:inline-block;width:100%;break-inside:avoid;background:var(--surface);border:1px solid var(--line);border-radius:20px;padding:22px;margin:0 0 18px;box-shadow:0 7px 22px rgba(42,28,75,.055);transition:transform .2s,box-shadow .2s,border-color .2s}.job-card:hover{transform:translateY(-3px);box-shadow:0 16px 36px rgba(42,28,75,.11);border-color:#d7cff7}
.job-top{display:flex;align-items:flex-start;justify-content:space-between;gap:12px}.score{font-size:1.65rem;font-weight:900;letter-spacing:-.045em;color:var(--ink)}.match{padding:6px 10px;border-radius:999px;font-size:.72rem;font-weight:800;white-space:nowrap}.match.strong{background:#e4f7ed;color:#18794e}.match.compatible{background:#eeeaff;color:#5940ce}.match.possible{background:#fff3d8;color:#956700}.match.low{background:#fae3e3;color:#a22b2b}
.job-title{font-size:1.17rem;line-height:1.25;margin:8px 0 5px;letter-spacing:-.025em}.company{font-weight:750;margin-bottom:6px}.meta{color:var(--muted);font-size:.83rem;line-height:1.5}
.chips{display:flex;flex-wrap:wrap;gap:6px;margin:15px 0}.chip{background:#f5f3f8;border:1px solid #ebe7f0;border-radius:999px;padding:5px 8px;font-size:.72rem;color:#5d5868}
.reason{background:#faf9fc;border-radius:14px;padding:13px 15px;margin-top:14px}.reason-title{font-weight:800;font-size:.82rem;margin-bottom:7px}.reason ul{margin:0;padding-left:18px;color:#5d5868;font-size:.8rem;line-height:1.55}
.requirements{font-size:.76rem;color:#5d5868;line-height:1.5;margin-top:12px}.source-row{display:flex;align-items:center;justify-content:space-between;gap:10px;margin-top:16px}.source{background:#f1eef7;border-radius:999px;padding:5px 9px;font-size:.7rem;font-weight:750;color:#5f586d}.recommend{font-size:.72rem;color:#18794e;font-weight:750}
.job-link{display:inline-flex;margin-top:16px;text-decoration:none;color:#fff;background:#181622;border-radius:11px;padding:10px 13px;font-size:.8rem;font-weight:800}.job-link:hover{background:#302b3c}
.empty{background:#fff;border:1px dashed #d7d0e2;border-radius:18px;padding:35px;text-align:center;color:var(--muted)}.hidden{display:none!important}
@media(max-width:800px){.container{width:min(100% - 20px,680px);padding-top:18px}.hero{padding:26px 20px;border-radius:24px}.form-grid{grid-template-columns:1fr 1fr}.full{grid-column:1/-1}.toolbar{grid-template-columns:1fr 1fr}.toolbar label:first-child{grid-column:1/-1}}
@media(max-width:560px){.container{width:calc(100% - 16px);padding-bottom:35px}.hero{padding:22px 16px}.brand{margin-bottom:10px}.logo{width:38px;height:38px;border-radius:12px}.hero-copy{font-size:.93rem;margin-bottom:20px}.form-grid{grid-template-columns:1fr}.full{grid-column:auto}.search-row{align-items:stretch;flex-direction:column}.search-row button{width:100%}.toolbar{grid-template-columns:1fr;padding:11px}.toolbar label:first-child{grid-column:auto}.section-head{align-items:flex-start;flex-direction:column}.jobs-grid{columns:1}.job-card{padding:18px;border-radius:17px}.job-top{display:block}.match{display:inline-block;margin-top:9px}.score{font-size:1.45rem}}
</style></head>
<body><main class="container">
<section class="hero">
<div class="brand"><div class="logo">JS</div><span class="muted">JobSearcher</span></div>
<h1>Encontre vagas que<br><span style="color:var(--primary)">combinam com você.</span></h1>
<p class="hero-copy">Pesquise oportunidades remotas em várias fontes e veja de forma transparente por que cada vaga combina com seu perfil — sem enviar ou armazenar seu currículo.</p>
<form method="post"><div class="form-grid">
<div class="field"><label for="roles">Cargo / área</label><input id="roles" name="roles" placeholder="Backend Developer, Software Engineer" value="{{.Roles}}"></div>
<div class="field"><label for="skills">Skills</label><input id="skills" name="skills" placeholder="Go, AWS, PostgreSQL, Docker" value="{{.Skills}}"></div>
<div class="field"><label for="seniority">Senioridade</label><select id="seniority" name="seniority"><option value="junior" {{if eq .Seniority "junior"}}selected{{end}}>Júnior</option><option value="mid" {{if eq .Seniority "mid"}}selected{{end}}>Pleno</option><option value="senior" {{if eq .Seniority "senior"}}selected{{end}}>Sênior</option><option value="staff" {{if eq .Seniority "staff"}}selected{{end}}>Staff / Lead</option></select></div>
<div class="field full"><label for="experience">Anos de experiência</label><input id="experience" type="number" min="0" name="experience" value="{{.Experience}}"></div></div>
{{if .Error}}<p class="error" role="alert">{{.Error}}</p>{{end}}
<div class="search-row"><div class="remote-note">🌎 Busca global · somente vagas 100% remotas</div><button id="searchButton" type="submit">🔎 Buscar vagas</button></div></form></section>
{{if .Searched}}<section class="section"><div class="section-head"><div><h2>Vagas para você</h2><p>Resultados elegíveis ordenados por compatibilidade.</p></div><div class="summary"><span class="stat"><strong id="visibleCount">{{.Count}}</strong> visíveis</span><span class="stat"><strong>{{.Count}}</strong> elegíveis</span></div></div>
<div class="toolbar"><label>Filtrar resultados<input id="jobFilter" type="search" placeholder="Cargo ou empresa"></label><label>Compatibilidade<select id="bucketFilter"><option value="all">Todas</option><option value="strong">Forte</option><option value="compatible">Compatível</option><option value="possible">Possível</option></select></label><label>Fonte<select id="sourceFilter"><option value="all">Todas</option>{{range .Jobs}}<option value="{{.Source}}">{{.Source}}</option>{{end}}</select></label></div>
<div id="jobsList" class="jobs-grid">{{range .Jobs}}<article class="job-card" data-bucket="{{.MatchBucket}}" data-source="{{.Source}}" data-search="{{.Title}} {{.Company}}">
<div class="job-top"><div class="score">{{.FitScore}}%</div><span class="match {{.MatchBucket}}">{{.MatchBucket}}</span></div>
<h3 class="job-title">{{.Title}}</h3><div class="company">{{.Company}}</div><div class="meta">{{.Location}} · {{.WorkplaceType}} · {{.Seniority}}</div>
<div class="chips"><span class="chip">Cargo {{.RoleMatch}}%</span><span class="chip">Skills {{.SkillMatch}}%</span><span class="chip">Relacionadas {{.RelatedSkillMatch}}%</span><span class="chip">Técnico {{.TechnicalMatch}}%</span><span class="chip">Responsabilidades {{.ResponsibilityMatch}}%</span><span class="chip">Senioridade {{.SeniorityMatch}}%</span><span class="chip">Experiência {{.ExperienceMatch}}%</span></div>
<div class="reason"><div class="reason-title">Por que combina</div><ul>{{range .MatchHighlights}}<li>{{.}}</li>{{end}}</ul></div>
<div class="requirements">✓ {{join .MustHaveMatch}} {{if .MustHaveMissing}} · △ {{join .MustHaveMissing}}{{end}}</div>
<div class="source-row"><span class="source">{{.Source}}</span><span class="recommend">{{.RecommendationStatus}}</span></div>
<a class="job-link" href="{{.URL}}" target="_blank" rel="noopener">Ver vaga ↗</a></article>{{else}}<div class="empty">Nenhuma vaga encontrada com esses critérios.</div>{{end}}</div>
<p id="emptyFiltered" class="empty hidden">Nenhuma vaga corresponde aos filtros selecionados.</p></section>
<script>
(function(){const text=document.getElementById('jobFilter'),bucket=document.getElementById('bucketFilter'),source=document.getElementById('sourceFilter');const cards=[...document.querySelectorAll('.job-card')],count=document.getElementById('visibleCount'),empty=document.getElementById('emptyFiltered');const values=[...source.options].slice(1).map(o=>o.value).filter(Boolean).sort((a,b)=>a.localeCompare(b));source.innerHTML='<option value="all">Todas</option>'+[...new Set(values)].map(v=>'<option value="'+v.replace(/"/g,'&quot;')+'">'+v+'</option>').join('');function apply(){const q=text.value.trim().toLowerCase(),b=bucket.value,s=source.value;let visible=0;cards.forEach(c=>{const ok=(!q||c.dataset.search.toLowerCase().includes(q))&&(b==='all'||c.dataset.bucket===b)&&(s==='all'||c.dataset.source===s);c.classList.toggle('hidden',!ok);if(ok)visible++});count.textContent=visible;empty.classList.toggle('hidden',visible!==0)}[text,bucket,source].forEach(el=>el.addEventListener('input',apply));apply()})();
const form=document.querySelector('form'),button=document.getElementById('searchButton');if(form&&button){form.addEventListener('submit',function(){button.disabled=true;button.textContent='⏳ Buscando vagas...'})}
</script>{{end}}</main></body></html>`

type webView struct { Roles, Skills, Experience, Seniority, Error string; Searched bool; Count int; Jobs []domain.Job }

func startWebServer(ctx context.Context, boards config.Boards) error {
 mux:=http.NewServeMux()
 mux.HandleFunc("/",func(w http.ResponseWriter,r *http.Request){
  if r.Method != http.MethodGet && r.Method != http.MethodPost {
   http.Error(w,"método não permitido",http.StatusMethodNotAllowed)
   return
  }
  v:=webView{Experience:"3",Seniority:"junior"}
  if r.Method==http.MethodPost {
   if err:=r.ParseForm(); err!=nil {v.Error="não foi possível ler os dados enviados"; renderWebPage(w,v); return}
   v.Roles=strings.TrimSpace(r.FormValue("roles")); v.Skills=strings.TrimSpace(r.FormValue("skills")); v.Experience=r.FormValue("experience"); v.Seniority=r.FormValue("seniority")
   roles:=csvValues(v.Roles); skills:=csvValues(v.Skills)
   switch {
   case len(roles)==0:
    v.Error="Informe pelo menos um cargo ou área."
   case len(skills)==0:
    v.Error="Informe pelo menos uma skill."
   case !validWebSeniority(v.Seniority):
    v.Error="Selecione uma senioridade válida."
   default:
    years,err:=strconv.Atoi(strings.TrimSpace(v.Experience))
    if err!=nil || years<0 {
     v.Error="Informe um número válido de anos de experiência."
    } else {
     profile:=config.Profile{TargetSeniority:v.Seniority,Titles:roles,Technologies:skills,YearsExperience:years,Weights:config.Weights{Technical:35,Responsibility:15,Seniority:25,Cloud:10,Domain:7,Language:4,AI:4}}
     search:=config.Search{Location:"",RemoteAllowed:[]string{"Worldwide"},PreferredTitles:roles,MinimumFitScore:55,FreshnessDays:7,ArchiveDays:30,MaxJobsPerSource:100}
     jobs,err:=searchForWeb(ctx,profile,search,boards,roles)
     if err!=nil {
      v.Error="Não foi possível concluir a busca: "+err.Error()
     } else {
      v.Searched=true;v.Jobs=jobs;v.Count=len(jobs)
     }
    }
   }
  }
  renderWebPage(w,v)
 })
 fmt.Println("JobSearcher web: http://localhost:8080")
 return http.ListenAndServe(":8080",mux)
}

func renderWebPage(w http.ResponseWriter,v webView){
 t,err:=template.New("page").Funcs(template.FuncMap{"join":func(v []string)string{return strings.Join(v,", ")}}).Parse(webPage)
 if err!=nil {http.Error(w,err.Error(),http.StatusInternalServerError);return}
 if err:=t.Execute(w,v);err!=nil {http.Error(w,err.Error(),http.StatusInternalServerError)}
}

func validWebSeniority(value string) bool {
 switch value {
 case "junior","mid","senior","staff":
  return true
 default:
  return false
 }
}

func searchForWeb(ctx context.Context,profile config.Profile,search config.Search,boards config.Boards,requestedRoles []string)([]domain.Job,error){
 tokens:=[]string{};sites:=[]string{}
 for _,b:=range boards.Greenhouse {if b.Enabled&&strings.TrimSpace(b.Token)!=""{tokens=append(tokens,b.Token)}}
 for _,b:=range boards.Lever {if b.Enabled&&strings.TrimSpace(b.Site)!=""{sites=append(sites,b.Site)}}
 searchProfile:=matching.BuildSearchProfile(profile); jobs,err:=fetchSources(ctx,"all",sources.Query{Terms:searchProfile.DiscoveryTerms,Location:"",BoardTokens:tokens,LeverSites:sites});if err!=nil&&len(jobs)==0{return nil,err}
 out:=[]domain.Job{}
 for _,j:=range jobs {if !isWithinDays(j,search.ArchiveDays)||!matching.IsRelevant(j,searchProfile.ExpandedRoles...)||!isFresh(j,search.FreshnessDays){continue}; x:=filter.Evaluate(j);matching.ScoreForRoles(&x,profile,searchProfile.Roles);if webJobEligible(x,search.MinimumFitScore){out=append(out,x)}}
 out=rankAndDedupeJobs(out)
 return limitJobsPerSource(out,search.MaxJobsPerSource),nil
}
func webJobEligible(j domain.Job, minimumFit int) bool {
	return j.WorkplaceType == "remote" && j.LocationEligible == domain.LocationEligible && j.FitScore >= minimumFit && j.SeniorityMatch > 20
}

func csvValues(s string)[]string{out:=[]string{};for _,v:=range strings.Split(s,","){if v=strings.TrimSpace(v);v!=""{out=append(out,v)}};return out}

func workModelAllowed(j domain.Job,remote,hybrid,onsite bool) bool { if !remote&&!hybrid&&!onsite{return true}; text:=strings.ToLower(j.WorkplaceType+" "+j.Location+" "+j.Description[:minLen(len(j.Description),500)]); ok:=false; if remote&&(strings.Contains(text,"remote")||strings.Contains(text,"remoto")){ok=true}; if hybrid&&(strings.Contains(text,"hybrid")||strings.Contains(text,"híbrido")||strings.Contains(text,"hibrido")){ok=true}; if onsite&&(strings.Contains(text,"onsite")||strings.Contains(text,"on-site")||strings.Contains(text,"presencial")){ok=true}; return ok }
func minLen(a,b int) int {if a<b{return a};return b}
