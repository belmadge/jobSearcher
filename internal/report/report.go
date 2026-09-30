package report

import (
	"fmt"
	"jobsearcher/internal/domain"
	"strings"
)

func Markdown(jobs []domain.Job, uncertain, rejected, found int) string {
	var b strings.Builder
	b.WriteString("# JobSearcher — search results\n\n")
	for _, section := range []struct {
		title string
		min   int
	}{{"🔥 Alta compatibilidade", 75}, {"🟡 Compatibilidade moderada", 35}} {
		fmt.Fprintf(&b, "## %s\n\n", section.title)
		for _, j := range jobs {
			if j.FitScore < section.min {
				continue
			}
			fmt.Fprintf(&b, "### %s — %s\n\n**Score:** %d/100  \n**Local:** %s  \n**Senioridade:** %s\n\n**Por que combina:** %s\n\n**Gaps:** %s\n\n**Salário:** %s  \n**Publicado:** %s  \n**Link:** %s\n\n", j.Title, j.Company, j.FitScore, j.Location, j.Seniority, strings.Join(j.Reasons, "; "), strings.Join(j.Gaps, ", "), j.Salary, j.PostedAt, j.URL)
		}
	}
	fmt.Fprintf(&b, "## ⚪ Para analisar\n\n%d vagas com localização incerta.\n\n## Estatísticas\n\n- Vagas encontradas: %d\n- Vagas aprovadas: %d\n- Vagas incertas: %d\n- Rejeitadas por localização: %d\n", uncertain, found, len(jobs), uncertain, rejected)
	return b.String()
}
