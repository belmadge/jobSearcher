package report

import (
 "fmt"
 "jobsearcher/internal/domain"
 "strings"
)

func Markdown(jobs []domain.Job, uncertain []domain.Job, rejected, found int, generatedAt string) string {
 var b strings.Builder
 b.WriteString("# JobSearcher — resultados da busca\n\n")
 b.WriteString("**Executado em:** "+generatedAt+"\n\n")
 sections:=[]struct{title string;min int}{{"🔥 Alta compatibilidade",80},{"🟡 Compatibilidade possível",60},{"⚪ Compatibilidade abaixo do limite",0}}
 for _,section:=range sections{
  fmt.Fprintf(&b,"## %s\n\n",section.title)
  count:=0
  for _,j:=range jobs{
   if (section.min==0 && j.FitScore>=60) || (section.min>0 && j.FitScore<section.min) || (section.min==80 && j.FitScore<80){continue}
   if section.min==60 && j.FitScore>=80{continue}
   count++
   fmt.Fprintf(&b,"### %s — %s\n\n**Score:** %d/100  \n**Local:** %s  \n**Modelo:** %s  \n**Senioridade:** %s\n\n**Por que combina:** %s\n\n**Gaps:** %s\n\n**Salário:** %s  \n**Publicado:** %s  \n**Link:** %s\n\n",j.Title,j.Company,j.FitScore,j.Location,j.WorkplaceType,j.Seniority,strings.Join(j.Reasons,"; "),strings.Join(j.Gaps,", "),j.Salary,j.PostedAt,j.URL)
  }
  if count==0{b.WriteString("Nenhuma vaga nesta faixa.\n\n")}
 }
 fmt.Fprintf(&b,"## ⚪ Para analisar\n\n%d vagas com localização incerta.\n\n",len(uncertain))
 fmt.Fprintf(&b,"## Estatísticas\n\n- Vagas encontradas: %d\n- Vagas elegíveis: %d\n- Vagas incertas: %d\n- Rejeitadas por localização: %d\n",found,len(jobs),len(uncertain),rejected)
 return b.String()
}
