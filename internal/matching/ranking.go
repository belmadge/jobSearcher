package matching

import "jobsearcher/internal/domain"

// BetterMatch returns true when a should appear before b in ranked results.
// FitScore remains the primary signal; the remaining fields only resolve ties.
func BetterMatch(a, b domain.Job) bool {
	if a.FitScore != b.FitScore {
		return a.FitScore > b.FitScore
	}
	tieBreakers := [][2]int{
		{a.RoleMatch, b.RoleMatch},
		{a.SkillMatch, b.SkillMatch},
		{a.TechnicalMatch, b.TechnicalMatch},
		{a.ResponsibilityMatch, b.ResponsibilityMatch},
		{a.ExperienceMatch, b.ExperienceMatch},
		{a.SeniorityMatch, b.SeniorityMatch},
		{a.RelatedSkillMatch, b.RelatedSkillMatch},
		{a.CloudMatch, b.CloudMatch},
	}
	for _, pair := range tieBreakers {
		if pair[0] != pair[1] {
			return pair[0] > pair[1]
		}
	}
	// Keep ordering deterministic when all ranking signals are equal.
	if a.Title != b.Title {
		return a.Title < b.Title
	}
	if a.Company != b.Company {
		return a.Company < b.Company
	}
	return a.URL < b.URL
}
