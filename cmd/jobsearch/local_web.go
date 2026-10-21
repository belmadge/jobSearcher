package main

import (
	"context"
	"html/template"
	"net/http"
	"sort"
	"strings"

	"jobsearcher/internal/config"
	"jobsearcher/internal/dedupe"
	"jobsearcher/internal/domain"
	"jobsearcher/internal/localjobs"
	"jobsearcher/internal/sources"
)

const localJobsPage = `<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>JobSearcher — Empregos locais</title>
<style>
:root{--ink:#1c2538;--muted:#657087;--line:#e5e9ef;--green:#16794c;--pale:#e5f7ed;--purple:#5636c9}
*{box-sizing:border-box}body{margin:0;background:linear-gradient(135deg,#fbfffc,#f0faf4);color:var(--ink);font-family:Inter,system-ui,-apple-system,"Segoe UI",sans-serif}.wrap{width:min(1120px,calc(100% - 32px));margin:auto;padding:22px 0 60px}.nav{display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:46px}.brand{font-weight:850;display:flex;align-items:center;gap:10px}.logo{background:#192234;color:white;border-radius:12px;width:42px;height:42px;display:grid;place-items:center}.pill{background:var(--pale);color:var(--green);border-radius:999px;padding:9px 13px;font-size:.82rem;font-weight:750}.hero{display:grid;grid-template-columns:1.1fr .9fr;gap:28px;align-items:center;margin-bottom:26px}.eyebrow{text-transform:uppercase;letter-spacing:.13em;font-size:.75rem;font-weight:800;color:var(--green)}h1{font-size:clamp(2.4rem,5vw,4.1rem);line-height:1;letter-spacing:-.06em;margin:14px 0}h1 span{color:var(--green)}p{line-height:1.6;color:var(--muted)}.art{height:210px;border-radius:28px;background:radial-gradient(circle at 60% 40%,#bfead0,#e5f7ed 60%,transparent 70%);display:grid;place-items:center;font-size:5rem}.panel,.card{background:white;border:1px solid var(--line);border-radius:20px;padding:22px;box-shadow:0 12px 32px #193b2410}.panel h2{margin:0 0 6px}.form{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:15px;margin-top:20px}.field{display:flex;flex-direction:column;gap:7px}.field label{font-size:.82rem;font-weight:750}.field input,.field select{width:100%;border:1px solid #dce2e8;border-radius:11px;padding:13px;background:#fff;color:var(--ink);font:inherit;outline-color:#7ac99c}.fixed{background:#f5faf7;border:1px solid #e2f1e7;border-radius:11px;padding:13px;font-size:.94rem}.submit{grid-column:1/-1;border:0;border-radius:12px;background:var(--green);color:#fff;padding:14px;font-weight:800;font-size:1rem;cursor:pointer}.submit:disabled{opacity:.65}.note{font-size:.82rem;color:var(--muted)}.results{margin-top:34px}.results h2{margin-bottom:5px}.grid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:16px;margin-top:18px}.card{padding:19px;display:flex;flex-direction:column;gap:10px}.score{display:inline-flex;background:var(--pale);color:var(--green);border-radius:999px;padding:6px 9px;font-weight:800;font-size:.78rem}.source{font-size:.75rem;color:var(--muted)}.card h3{margin:0;font-size:1.1rem}.meta{font-size:.85rem;color:var(--muted)}.link{margin-top:auto;padding-top:12px;color:var(--green);font-weight:800}.error{background:#fff1f1;color:#9c2525;border-radius:12px;padding:13px;margin-top:14px}.empty{padding:24px;border:1px dashed #cbd5ce;border-radius:14px;color:var(--muted);margin-top:16px}
@media(max-width:760px){.hero{grid-template-columns:1fr}.art{height:120px;font-size:3.5rem}.grid{grid-template-columns:repeat(2,minmax(0,1fr))}.nav{margin-bottom:30px}}@media(max-width:520px){.wrap{width:calc(100% - 22px);padding-top:12px}.pill{font-size:.7rem;padding:8px}.form,.grid{grid-template-columns:1fr}.panel,.card{padding:16px}.hero{gap:8px}}
</style></head><body><main class="wrap">
<header class="nav"><a class="brand" href="/"><span class="logo">JS</span><span>JobSearcher</span></a><span class="pill">⌖ Maceió, AL · Presencial</span></header>
<section class="hero"><div><div class="eyebrow">Empregos locais</div><h1>Seu próximo trabalho pode estar <span>perto.</span></h1><p>Encontre oportunidades de vendas, atendimento, saúde e outras áreas em Maceió. Sem currículo, sem cadastro e sem salvar seus dados pessoais.</p></div><div class="art" aria-hidden="true">⌂ ✦</div></section>
<section class="panel"><h2>Buscar vagas presenciais</h2><p>Local e modalidade são fixos para manter os resultados dentro dos critérios deste módulo.</p>
<form method="post" autocomplete="off" class="form">
<div class="field"><label for="title">Cargo desejado</label><input id="title" name="title" value="{{.Title}}" required autocomplete="off" placeholder="Ex.: Vendedora"></div>
<div class="field"><label for="category">Área profissional</label><select id="category" name="category"><option {{if eq .Category "Vendas e comércio"}}selected{{end}}>Vendas e comércio</option><option {{if eq .Category "Saúde e clínicas"}}selected{{end}}>Saúde e clínicas</option><option {{if eq .Category "Atendimento ao público"}}selected{{end}}>Atendimento ao público</option><option {{if eq .Category "Administrativo"}}selected{{end}}>Administrativo</option><option {{if eq .Category "Hotelaria e serviços"}}selected{{end}}>Hotelaria e serviços</option><option {{if eq .Category "Todas as áreas"}}selected{{end}}>Todas as áreas</option></select></div>
<div class="field"><label for="experience">Experiência</label><select id="experience" name="experience"><option value="any" {{if eq .Experience "any"}}selected{{end}}>Sem preferência</option><option value="none" {{if eq .Experience "none"}}selected{{end}}>Sem experiência / primeiro emprego</option><option value="required" {{if eq .Experience "required"}}selected{{end}}>Priorizar vagas com experiência</option></select></div>
<div class="field"><label for="education">Escolaridade</label><select id="education" name="education"><option value="any" {{if eq .Education "any"}}selected{{end}}>Qualquer</option><option value="elementary" {{if eq .Education "elementary"}}selected{{end}}>Ensino fundamental</option><option value="high_school" {{if eq .Education "high_school"}}selected{{end}}>Ensino médio</option><option value="higher" {{if eq .Education "higher"}}selected{{end}}>Ensino superior</option></select></div>
<div class="field"><label>Localização fixa</label><div class="fixed">Maceió, Alagoas, Brasil</div></div><div class="field"><label>Modalidade fixa</label><div class="fixed">Presencial — sem remoto ou híbrido</div></div>
<button class="submit" id="localSearchButton" type="submit">Encontrar vagas em Maceió</button>
</form>{{if .Error}}<div class="error">{{.Error}}</div>{{end}}<p class="note">Os campos são usados apenas para esta busca. "Sem preferência" e "Qualquer" não excluem vagas por experiência ou escolaridade.</p>
</section>
{{if .Searched}}<section class="results"><h2>{{.Count}} vagas presenciais elegíveis</h2><p>Resultados encontrados nas fontes disponíveis. Confirme os detalhes na publicação original.</p><div class="grid">{{range .Jobs}}<article class="card"><div><span class="score">{{.LocalScore}}% de alinhamento</span> <span class="source">{{.Source}}</span></div><h3>{{.Title}}</h3><strong>{{.Company}}</strong><div class="meta">⌖ {{.Location}} · Presencial</div><p>{{.Description}}</p><a class="link" href="{{.URL}}" target="_blank" rel="noopener">Ver publicação original ↗</a></article>{{else}}<div class="empty">Nenhuma vaga encontrada com todos os critérios. Isso não significa que não existam oportunidades; as fontes conectadas podem não cobrir todas as vagas locais.</div>{{end}}</div></section>{{end}}
</main><script>const f=document.querySelector('form'),b=document.getElementById('localSearchButton');if(f&&b)f.addEventListener('submit',()=>{b.disabled=true;b.textContent='Buscando vagas...'})</script></body></html>`

type localJobResult struct {
	domain.Job
	LocalScore int
}

type localJobsView struct {
	Title, Category, Experience, Education, Error string
	Searched bool
	Count int
	Jobs []localJobResult
}

func handleLocalJobs(ctx context.Context, boards config.Boards, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
		return
	}
	p := localjobs.DefaultProfile()
	v := localJobsView{Title:p.Title, Category:p.Category, Experience:p.Experience, Education:p.Education}
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			v.Error = "Não foi possível ler os dados enviados."
			renderLocalJobs(w, v)
			return
		}
		p.Title = strings.TrimSpace(r.FormValue("title"))
		p.Category = strings.TrimSpace(r.FormValue("category"))
		p.Experience = r.FormValue("experience")
		p.Education = r.FormValue("education")
		v.Title, v.Category, v.Experience, v.Education = p.Title, p.Category, p.Experience, p.Education
		if p.Title == "" {
			v.Error = "Informe o cargo desejado."
		} else if !validLocalChoice(p.Experience, "any", "none", "required") {
			v.Error = "Selecione uma opção válida para experiência."
		} else if !validLocalChoice(p.Education, "any", "elementary", "high_school", "higher") {
			v.Error = "Selecione uma opção válida para escolaridade."
		} else {
			p.City, p.State, p.Country = "Maceió", "AL", "Brasil"
			terms := localjobs.SearchTerms(p)
			tokens, sites := []string{}, []string{}
			for _, board := range boards.Greenhouse { if board.Enabled && strings.TrimSpace(board.Token) != "" { tokens = append(tokens, board.Token) } }
			for _, site := range boards.Lever { if site.Enabled && strings.TrimSpace(site.Site) != "" { sites = append(sites, site.Site) } }
			jobs, err := fetchSources(ctx, "all", sources.Query{Terms: terms, Location: "Maceió, AL, Brazil", BoardTokens: tokens, LeverSites: sites})
			if err != nil && len(jobs) == 0 {
				v.Error = "Não foi possível consultar as fontes: " + err.Error()
			} else {
				if err != nil {
					v.Error = "Algumas fontes falharam; os resultados disponíveis foram mantidos."
				}
				seen := map[string]bool{}
				for _, job := range jobs {
					if !localjobs.Eligible(job, p) {
						continue
					}
					score := localjobs.MatchScore(job, p)
					key := dedupe.CanonicalURL(job.URL)
					if key == "" || seen[key] {
						continue
					}
					seen[key] = true
					v.Jobs = append(v.Jobs, localJobResult{Job: job, LocalScore: score})
				}
				sort.SliceStable(v.Jobs, func(i, j int) bool {
					if v.Jobs[i].LocalScore != v.Jobs[j].LocalScore {
						return v.Jobs[i].LocalScore > v.Jobs[j].LocalScore
					}
					if v.Jobs[i].Title != v.Jobs[j].Title {
						return v.Jobs[i].Title < v.Jobs[j].Title
					}
					return v.Jobs[i].Company < v.Jobs[j].Company
				})
				v.Searched = true
				v.Count = len(v.Jobs)
				// Do not repopulate the form after a successful search.
				v.Title, v.Category, v.Experience, v.Education = "", "", "", ""
			}
		}
	}
	renderLocalJobs(w, v)
}

func renderLocalJobs(w http.ResponseWriter, v localJobsView) {
	t, err := template.New("local-jobs").Parse(localJobsPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := t.Execute(w, v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func validLocalChoice(value string, allowed ...string) bool {
	for _, choice := range allowed {
		if value == choice {
			return true
		}
	}
	return false
}

