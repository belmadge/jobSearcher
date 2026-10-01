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
<title>JobSearcher — vagas remotas para sua carreira</title>
<style>
:root{--bg:#fbfaff;--surface:#fff;--ink:#182038;--muted:#697087;--line:#e7e8ef;--primary:#5636c9;--primary2:#7651df;--lav:#f1edff;--shadow:0 12px 34px rgba(47,35,88,.07)}
*{box-sizing:border-box}html{scroll-behavior:smooth}
body{margin:0;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:var(--ink);background:linear-gradient(135deg,#fff 0%,#fbf9ff 62%,#f2ecff 100%);min-height:100vh}
a{text-decoration:none;color:inherit}.container{width:min(1400px,calc(100% - 56px));margin:auto;padding:16px 0 70px}
.nav{height:48px;display:flex;align-items:center;justify-content:space-between;margin-bottom:24px}.brand{display:flex;align-items:center;gap:12px;font-weight:850;font-size:1.05rem}.logo{width:42px;height:42px;border-radius:12px;background:#171927;color:#fff;display:grid;place-items:center;font-weight:900;box-shadow:0 8px 18px rgba(20,20,35,.16)}
.nav-meta{display:flex;align-items:center;gap:22px;color:#37394a;font-size:.8rem;font-weight:650}.nav-meta span{display:flex;align-items:center;gap:7px}.divider{width:1px;height:17px;background:#d9d8e2}
.hero{position:relative;overflow:hidden;padding:10px 4px 30px;min-height:365px}.hero-copy{position:relative;z-index:2;max-width:770px;padding-top:14px}.eyebrow{font-size:.78rem;letter-spacing:.13em;text-transform:uppercase;font-weight:850;color:#7150a4;margin-bottom:14px}
h1{font-size:clamp(2.8rem,5vw,4.4rem);line-height:.96;letter-spacing:-.065em;margin:0 0 18px;font-weight:900;color:#172039}h1 span{background:linear-gradient(90deg,#5133bd,#8b43b9);-webkit-background-clip:text;background-clip:text;color:transparent}.hero-desc{font-size:1rem;line-height:1.55;color:#59627a;max-width:700px;margin:0}
.scene{position:absolute;right:-10px;top:-15px;width:530px;height:320px;opacity:.9;pointer-events:none}.glow{position:absolute;width:320px;height:320px;border-radius:50%;right:40px;top:5px;background:radial-gradient(circle,#eadfff 0,#f7f2ff 62%,transparent 70%)}.plant{position:absolute;right:15px;top:18px;width:100px;height:235px}.leaf{position:absolute;width:70px;height:30px;border-radius:100% 0 100% 0;background:#3b3a56;transform-origin:100% 100%;opacity:.9}.leaf:nth-child(1){right:18px;top:15px;transform:rotate(-35deg)}.leaf:nth-child(2){right:2px;top:45px;transform:rotate(8deg)}.leaf:nth-child(3){right:24px;top:70px;transform:rotate(-48deg)}.leaf:nth-child(4){right:0;top:102px;transform:rotate(18deg)}.leaf:nth-child(5){right:22px;top:135px;transform:rotate(-55deg)}.stem{position:absolute;width:5px;height:150px;background:#56536e;right:32px;top:66px}.pot{position:absolute;right:7px;bottom:5px;width:78px;height:48px;background:#e9d9c9;border-radius:8px 8px 20px 20px}.laptop{position:absolute;right:140px;bottom:35px;width:250px;height:150px;border:8px solid #dedde5;border-radius:13px;background:linear-gradient(135deg,#fafafa,#dfe1e8);transform:skew(-5deg);box-shadow:0 12px 22px rgba(50,45,65,.13)}.laptop:after{content:"";position:absolute;left:-32px;right:-32px;bottom:-28px;height:16px;border-radius:8px 8px 4px 4px;background:#c7c5d0}.screen{position:absolute;inset:10px;background:linear-gradient(135deg,#e9e6f8,#f9f8ff)}.mug{position:absolute;right:355px;bottom:27px;width:52px;height:42px;background:#2a2840;border-radius:5px 5px 13px 13px;box-shadow:0 8px 12px rgba(40,35,60,.12)}.mug:after{content:"";position:absolute;right:-15px;top:9px;width:20px;height:20px;border:5px solid #2a2840;border-left:0;border-radius:0 14px 14px 0}.mug:before{content:"JS";position:absolute;color:#fff;font-size:11px;font-weight:900;left:17px;top:14px}.books{position:absolute;right:55px;bottom:5px;width:180px;height:35px;background:#c6b5aa;box-shadow:0 -14px 0 #ddd1c8,0 -28px 0 #8c8498;border-radius:2px}
.search-panel{position:relative;z-index:3;background:rgba(255,255,255,.9);border:1px solid #e0e1e9;border-radius:12px;box-shadow:0 8px 24px rgba(47,35,88,.06);padding:13px;margin-top:-5px}
.form-grid{display:grid;grid-template-columns:1.05fr 1.1fr .72fr .7fr auto;gap:18px;align-items:end}.field label{display:flex;gap:7px;align-items:center;font-size:.76rem;font-weight:800;margin:0 0 8px}.field-icon{font-size:.9rem}.field input,.field select{width:100%;height:45px;border:1px solid #dcdde6;border-radius:9px;background:#fff;padding:0 13px;color:var(--ink);font:inherit;outline:none}.field input:focus,.field select:focus{border-color:#8465d8;box-shadow:0 0 0 3px #eee9fb}.field input:-webkit-autofill,.field input:-webkit-autofill:focus{ -webkit-text-fill-color:var(--ink);-webkit-box-shadow:0 0 0 1000px #fff inset;box-shadow:0 0 0 1000px #fff inset}
.search-button{height:45px;border:0;border-radius:9px;background:linear-gradient(135deg,#5131b7,#7146dc);color:#fff;font:inherit;font-weight:800;padding:0 25px;cursor:pointer;white-space:nowrap;box-shadow:0 8px 18px rgba(83,49,183,.2)}.search-button:disabled{opacity:.7;cursor:wait}.error{background:#fff0f0;border:1px solid #efc3c3;color:#8e2525;border-radius:9px;padding:10px;margin:12px 0 0}
.remote-row{display:flex;align-items:center;gap:10px;margin-top:12px;font-size:.78rem;color:#555e74}.toggle{width:39px;height:22px;border-radius:99px;background:#7b4bdb;padding:3px;display:inline-flex;justify-content:flex-end}.toggle i{width:16px;height:16px;background:#fff;border-radius:50%;display:block}
.results{display:grid;grid-template-columns:260px 1fr;gap:18px;margin-top:22px}.filters{background:rgba(255,255,255,.92);border:1px solid #e8e8ef;border-radius:16px;padding:18px;height:max-content;box-shadow:0 7px 22px rgba(47,35,88,.045);position:sticky;top:15px}.filter-title{display:flex;justify-content:space-between;align-items:center;font-weight:850;font-size:1.05rem;margin-bottom:18px}.clear{color:#5b3abf;font-size:.75rem}.filter-label{font-size:.72rem;font-weight:800;margin:15px 0 7px;display:block}.filters input,.filters select{width:100%;height:40px;border:1px solid #e0e1e8;border-radius:8px;background:#fff;padding:0 10px;font:inherit;color:var(--ink)}.result-count{margin-top:26px;font-weight:850}.result-count small{display:block;color:#7a8191;font-weight:500;margin-top:3px}
.results-main{min-width:0}.results-head{display:flex;justify-content:space-between;align-items:center;margin-bottom:12px}.results-head h2{margin:0;font-size:1.25rem;letter-spacing:-.025em}.sort{font-size:.75rem;color:#687084;display:flex;align-items:center;gap:7px}.sort select{height:35px;border:1px solid #dedfe7;border-radius:8px;background:#fff;padding:0 9px;color:var(--ink)}
.jobs-grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:15px}.job-card{background:#fff;border:1px solid #e5e5ec;border-radius:15px;padding:16px;box-shadow:0 7px 20px rgba(47,35,88,.045);transition:.2s;min-width:0}.job-card:hover{transform:translateY(-2px);box-shadow:0 14px 28px rgba(47,35,88,.09);border-color:#d8cef4}
.card-top{display:flex;justify-content:space-between;gap:10px;align-items:flex-start}.company-mark{width:40px;height:40px;border-radius:10px;background:#191b2a;color:#fff;display:grid;place-items:center;font-weight:900;font-size:.95rem}.company-info{display:flex;gap:9px;align-items:center;min-width:0}.company-info strong{font-size:.8rem;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.source{font-size:.62rem;background:#f0f0f3;border-radius:99px;padding:5px 8px;color:#697083}
.score-ring{width:58px;height:58px;border-radius:50%;background:conic-gradient(#7139cf calc(var(--score)*1%),#eeeafa 0);display:grid;place-items:center;flex:0 0 auto}.score-ring:before{content:"";width:45px;height:45px;background:#fff;border-radius:50%;grid-area:1/1}.score-ring span{grid-area:1/1;position:relative;font-size:.78rem;font-weight:900}
.job-card h3{font-size:1rem;line-height:1.25;margin:14px 0 6px;letter-spacing:-.025em}.meta{font-size:.72rem;color:#697183;line-height:1.5}.meta span{margin-right:7px}.match-label{text-align:right;font-size:.65rem;font-weight:800;margin-top:3px;color:#5940ad}
.chips{display:flex;flex-wrap:wrap;gap:5px;margin:12px 0}.chip{background:#f0f0f3;border-radius:99px;padding:4px 7px;font-size:.63rem;color:#5d6474}
.reason{background:#faf9fc;border-radius:11px;padding:10px 11px;margin-top:9px}.reason-title{font-size:.72rem;font-weight:850;margin-bottom:4px}.reason ul{margin:0;padding-left:15px;color:#677083;font-size:.66rem;line-height:1.5}
.card-bottom{display:flex;gap:7px;margin-top:12px}.job-link{flex:1;text-align:center;background:#f3efff;border:1px solid #ddd2fa;color:#5034ae;border-radius:8px;padding:8px;font-size:.72rem;font-weight:850}.bookmark{width:36px;border:1px solid #dedfe7;background:#fff;border-radius:8px;font-size:1rem}.recommend{font-size:.62rem;color:#18794e;font-weight:750;margin-top:7px}
.empty{background:#fff;border:1px dashed #d7d0e2;border-radius:15px;padding:35px;text-align:center;color:var(--muted)}.hidden{display:none!important}
@media(max-width:1050px){.scene{opacity:.35;right:-100px}.form-grid{grid-template-columns:1fr 1fr}.field:nth-child(3){grid-column:1}.field:nth-child(4){grid-column:2}.search-button{width:100%}.jobs-grid{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media(max-width:760px){.container{width:min(100% - 24px,680px)}.nav-meta{display:none}.hero{min-height:0}.scene{display:none}.hero-copy{padding-top:8px}h1{font-size:clamp(2.5rem,11vw,3.7rem)}.results{grid-template-columns:1fr}.filters{position:static}.jobs-grid{grid-template-columns:1fr}.form-grid{grid-template-columns:1fr}.field:nth-child(3),.field:nth-child(4){grid-column:auto}.results-head{align-items:flex-start;gap:10px;flex-direction:column}}
@media(max-width:480px){.container{width:calc(100% - 16px);padding-top:10px}.nav{margin-bottom:12px}.logo{width:38px;height:38px}.hero{padding:0 3px 20px}.hero-copy{padding:5px 0}.eyebrow{font-size:.68rem}h1{font-size:2.55rem}.search-panel{padding:11px}.results{margin-top:15px}.job-card{padding:14px}}
</style></head>
<body><main class="container">
<header class="nav"><div class="brand"><div class="logo">JS</div><span>JobSearcher</span></div><div class="nav-meta"><span>◉ &nbsp;100% Remoto</span><i class="divider"></i><span>✣ &nbsp;Várias fontes</span><i class="divider"></i><span>♢ &nbsp;Sem currículo</span></div></header>
<section class="hero"><div class="hero-copy"><div class="eyebrow">VAGAS REMOTAS PARA A SUA CARREIRA</div><h1>Encontre vagas que<br><span>combinam com você.</span></h1><p class="hero-desc">Pesquise oportunidades remotas em várias fontes e veja de forma transparente por que cada vaga combina com o seu perfil — sem enviar ou armazenar seu currículo.</p></div>
<div class="scene" aria-hidden="true"><div class="glow"></div><div class="laptop"><div class="screen"></div></div><div class="mug"></div><div class="books"></div><div class="plant"><div class="stem"></div><div class="leaf"></div><div class="leaf"></div><div class="leaf"></div><div class="leaf"></div><div class="leaf"></div><div class="pot"></div></div></div>
<div class="search-panel"><form method="post"><div class="form-grid">
<div class="field"><label for="roles"><span class="field-icon">▣</span> Cargo / área</label><input id="roles" name="roles" placeholder="Backend Developer, Software Engineer" value="{{.Roles}}"></div>
<div class="field"><label for="skills"><span class="field-icon">&lt;/&gt;</span> Skills</label><input id="skills" name="skills" placeholder="Go, AWS, PostgreSQL, Docker" value="{{.Skills}}"></div>
<div class="field"><label for="seniority"><span class="field-icon">▥</span> Senioridade</label><select id="seniority" name="seniority"><option value="junior" {{if eq .Seniority "junior"}}selected{{end}}>Júnior</option><option value="mid" {{if eq .Seniority "mid"}}selected{{end}}>Pleno</option><option value="senior" {{if eq .Seniority "senior"}}selected{{end}}>Sênior</option><option value="staff" {{if eq .Seniority "staff"}}selected{{end}}>Staff / Lead</option></select></div>
<div class="field"><label for="experience"><span class="field-icon">♙</span> Anos de experiência</label><input id="experience" type="number" min="0" name="experience" value="{{.Experience}}"></div>
<div class="field"><button class="search-button" id="searchButton" type="submit">⌕ &nbsp;Buscar vagas&nbsp; →</button></div>
</div>{{if .Error}}<p class="error" role="alert">{{.Error}}</p>{{end}}<div class="remote-row"><span class="toggle"><i></i></span><span>Busca global — somente vagas 100% remotas</span></div></form></div>
</section>
{{if .Searched}}<section class="results">
<aside class="filters"><div class="filter-title"><span>☷ &nbsp;Filtros</span><a class="clear" href="#" id="clearFilters">Limpar</a></div>
<label class="filter-label">Buscar por cargo ou empresa</label><input id="jobFilter" type="search" placeholder="Cargo ou empresa...">
<label class="filter-label">Compatibilidade mínima</label><select id="bucketFilter"><option value="all">Todas</option><option value="strong">Forte (≥ 75%)</option><option value="compatible">Compatível (≥ 60%)</option><option value="possible">Possível (≥ 40%)</option></select>
<label class="filter-label">Fonte</label><select id="sourceFilter"><option value="all">Todas</option>{{range .Jobs}}<option value="{{.Source}}">{{.Source}}</option>{{end}}</select>
<div class="result-count"><span id="visibleCount">{{.Count}} vagas</span><small>de {{.Count}} elegíveis</small></div></aside>
<div class="results-main"><div class="results-head"><div><h2>Resultados</h2></div><label class="sort">Ordenar por <select id="sortFilter"><option value="score">Melhor compatibilidade</option><option value="title">Cargo</option><option value="source">Fonte</option></select></label></div>
<div id="jobsList" class="jobs-grid">{{range .Jobs}}<article class="job-card" data-bucket="{{.MatchBucket}}" data-source="{{.Source}}" data-search="{{.Title}} {{.Company}}" data-score="{{.FitScore}}">
<div class="card-top"><div class="company-info"><div class="company-mark">{{if .Company}}{{printf "%.1s" .Company}}{{else}}J{{end}}</div><strong>{{.Company}}</strong></div><div><span class="source">{{.Source}}</span><div class="score-ring" style="--score:{{.FitScore}}"><span>{{.FitScore}}%</span></div><div class="match-label">{{.MatchBucket}}</div></div></div>
<h3>{{.Title}}</h3><div class="meta"><span>⌖ {{.Location}}</span><span>⌂ {{.WorkplaceType}}</span><span>▣ {{.Seniority}}</span></div>
<div class="chips"><span class="chip">Cargo {{.RoleMatch}}%</span><span class="chip">Skills {{.SkillMatch}}%</span><span class="chip">Técnico {{.TechnicalMatch}}%</span><span class="chip">Responsabilidades {{.ResponsibilityMatch}}%</span><span class="chip">Senioridade {{.SeniorityMatch}}%</span><span class="chip">Experiência {{.ExperienceMatch}}%</span></div>
<div class="reason"><div class="reason-title">✦ &nbsp;Por que combina</div><ul>{{range .MatchHighlights}}<li>{{.}}</li>{{end}}</ul></div>
<div class="recommend">{{.RecommendationStatus}}</div><div class="card-bottom"><a class="job-link" href="{{.URL}}" target="_blank" rel="noopener">Ver vaga &nbsp;↗</a><button class="bookmark" type="button" aria-label="Salvar vaga">♡</button></div>
</article>{{else}}<div class="empty">Nenhuma vaga encontrada com esses critérios.</div>{{end}}</div><p id="emptyFiltered" class="empty hidden">Nenhuma vaga corresponde aos filtros selecionados.</p></div>
</section>
<script>
(function(){const text=document.getElementById('jobFilter'),bucket=document.getElementById('bucketFilter'),source=document.getElementById('sourceFilter'),sort=document.getElementById('sortFilter'),clear=document.getElementById('clearFilters'),list=document.getElementById('jobsList'),cards=[...document.querySelectorAll('.job-card')],count=document.getElementById('visibleCount'),empty=document.getElementById('emptyFiltered');const vals=[...source.options].slice(1).map(o=>o.value).filter(Boolean).sort();source.innerHTML='<option value="all">Todas</option>'+[...new Set(vals)].map(v=>'<option value="'+v.replace(/"/g,'&quot;')+'">'+v+'</option>').join('');
function apply(){const q=text.value.trim().toLowerCase(),b=bucket.value,s=source.value;let visible=0;cards.forEach(c=>{const ok=(!q||c.dataset.search.toLowerCase().includes(q))&&(b==='all'||c.dataset.bucket===b)&&(s==='all'||c.dataset.source===s);c.classList.toggle('hidden',!ok);if(ok)visible++});count.textContent=visible+' vagas';empty.classList.toggle('hidden',visible!==0)}
function order(){const arr=[...cards];arr.sort((a,b)=>sort.value==='title'?a.querySelector('h3').textContent.localeCompare(b.querySelector('h3').textContent):sort.value==='source'?a.dataset.source.localeCompare(b.dataset.source):Number(b.dataset.score)-Number(a.dataset.score));arr.forEach(x=>list.appendChild(x));apply()}
[text,bucket,source].forEach(x=>x.addEventListener('input',apply));sort.addEventListener('change',order);clear.addEventListener('click',e=>{e.preventDefault();text.value='';bucket.value='all';source.value='all';sort.value='score';order()});apply()})();
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
