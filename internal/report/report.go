package report

import (
 "fmt"
 "jobsearcher/internal/domain"
 "strings"
)

func Markdown(jobs []domain.Job, senior []domain.Job, archived []domain.Job, uncertain []domain.Job, rejected, found, irrelevant, outsideSeniority, lowFit, stale int, generatedAt string) string {
 var b strings.Builder
 b.WriteString("# JobSearcher — resultados da busca\n\n")
 b.WriteString("**Executado em:** "+generatedAt+"\n\n")
 sections:=[]struct{title string;min int}{{"🔥 Alta compatibilidade",80},{"🟡 Compatibilidade possível",60}}
 for _,section:=range sections{
  fmt.Fprintf(&b,"## %s\n\n",section.title)
  count:=0
  for _,j:=range jobs{
   if (section.min==0 && j.FitScore>=60) || (section.min>0 && j.FitScore<section.min) || (section.min==80 && j.FitScore<80){continue}
   if section.min==60 && j.FitScore>=80{continue}
   count++
   fmt.Fprintf(&b,"### %s — %s\n\n**Score:** %d/100  \n**Fonte:** %s  \n**Local:** %s  \n**Modelo:** %s  \n**Senioridade:** %s  \n**Leitura de senioridade:** %s\n\n**Por que combina:** %s\n\n**Destaques do matching:** %s\n\n**Gaps:** %s\n\n**Salário:** %s  \n**Publicado:** %s  \n**Link:** %s\n\n",j.Title,j.Company,j.FitScore,j.Source,j.Location,j.WorkplaceType,j.Seniority,j.SeniorityReason,strings.Join(j.Reasons,"; "),strings.Join(j.MatchHighlights,"; "),strings.Join(j.Gaps,", "),j.Salary,j.PostedAt,j.URL)
  }
  if count==0{b.WriteString("Nenhuma vaga nesta faixa.\n\n")}
 }
 fmt.Fprintf(&b,"## 🟠 Acima da senioridade alvo\n\n%d vagas Senior com score ≥ 60 para análise.\n\n",len(senior))
 for _,j:=range senior {
  fmt.Fprintf(&b,"### %s — %s\n\n**Score:** %d/100  \n**Fonte:** %s  \n**Local:** %s  \n**Modelo:** %s  \n**Senioridade:** %s  \n**Leitura de senioridade:** %s\n\n**Por que combina:** %s\n\n**Gaps:** %s\n\n**Salário:** %s  \n**Publicado:** %s  \n**Link:** %s\n\n",j.Title,j.Company,j.FitScore,j.Source,j.Location,j.WorkplaceType,j.Seniority,j.SeniorityReason,strings.Join(j.Reasons,"; "),strings.Join(j.MatchHighlights,"; "),strings.Join(j.Gaps,", "),j.Salary,j.PostedAt,j.URL)
 }
 fmt.Fprintf(&b,"## 🕰️ Oportunidades antigas\n\n%d vagas compatíveis publicadas/atualizadas há mais de %d e até 30 dias.\n\n",len(archived),7)
 for _,j:=range archived {
  fmt.Fprintf(&b,"- **%s — %s** — Score %d/100 — %s — %s — %s\n",j.Title,j.Company,j.FitScore,j.Location,j.WorkplaceType,j.URL)
 }
 b.WriteString("\n")
 fmt.Fprintf(&b,"## ⚪ Para analisar\n\n%d vagas com localização incerta.\n\n",len(uncertain))
 for _,j:=range uncertain {
  fmt.Fprintf(&b,"- **%s — %s** — %s. Local: %s. Link: %s\n",j.Title,j.Company,j.LocationReason,j.Location,j.URL)
 }
 b.WriteString("\n")
 fmt.Fprintf(&b,"## Estatísticas\n\n- Vagas encontradas: %d\n- Vagas elegíveis (score mínimo): %d\n- Senior ≥60 para análise: %d\n- Oportunidades antigas: %d\n- Vagas incertas: %d\n- Rejeitadas por localização: %d\n- Fora do perfil por título: %d\n- Senioridade fora do alvo: %d\n- Abaixo do score mínimo: %d\n- Antigas: %d\n",found,len(jobs),len(senior),len(archived),len(uncertain),rejected,irrelevant,outsideSeniority,lowFit,stale)
 return b.String()
}
