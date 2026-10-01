package matching

import (
	"strings"

	"jobsearcher/internal/config"
)

// SearchProfile is the normalized search intent derived from the user's form.
type SearchProfile struct {
	Roles          []string
	ExpandedRoles  []string
	Skills         []string
	ExpandedSkills []string
}

func BuildSearchProfile(p config.Profile) SearchProfile {
	roles := uniqueNormalized(p.Titles)
	skills := uniqueNormalized(p.Technologies)
	return SearchProfile{
		Roles: roles, ExpandedRoles: expandRoles(roles),
		Skills: skills, ExpandedSkills: expandSkills(skills),
	}
}

var roleAliases = map[string][]string{
	"backend developer": {"backend developer","backend engineer","back end developer","back end engineer","back-end developer","back-end engineer","backend software engineer","api developer","api engineer","server-side developer"},
	"backend engineer": {"backend developer","backend engineer","back end developer","back end engineer","back-end developer","back-end engineer","backend software engineer","api developer","api engineer","server-side developer"},
	"software engineer": {"software engineer","software developer","backend engineer","backend developer","api engineer","platform engineer"},
	"software developer": {"software engineer","software developer","backend engineer","backend developer","api engineer","platform engineer"},
	"qa": {"qa","qa engineer","quality assurance","quality assurance engineer","quality engineer","software qa","sdet","software development engineer in test","test automation engineer","automation qa","qa automation","test engineer"},
	"quality assurance": {"qa","qa engineer","quality assurance","quality assurance engineer","quality engineer","software qa","sdet","software development engineer in test","test automation engineer","automation qa","qa automation","test engineer"},
	"devops": {"devops engineer","devops developer","site reliability engineer","sre","platform engineer","infrastructure engineer"},
	"sre": {"site reliability engineer","sre","reliability engineer","platform engineer","infrastructure engineer"},
	"data engineer": {"data engineer","analytics engineer","data platform engineer"},
}

var skillAliases = map[string][]string{
	"go": {"go","golang"}, "golang": {"go","golang"},
	"postgresql": {"postgresql","postgres","postgre sql"}, "postgres": {"postgresql","postgres","postgre sql"},
	"sql": {"sql","structured query language"}, "aws": {"aws","amazon web services"},
	"gcp": {"gcp","google cloud","google cloud platform"}, "azure": {"azure","microsoft azure"},
	"graphql": {"graphql","graph ql"}, "docker": {"docker","containerization","containers"},
	"kubernetes": {"kubernetes","k8s"}, "terraform": {"terraform","infrastructure as code","iac"},
	"rabbitmq": {"rabbitmq","rabbit mq","message broker"},
	"ci/cd": {"ci/cd","ci cd","continuous integration","continuous delivery"},
	"datadog": {"datadog","data dog"}, "observability": {"observability","monitoring","distributed tracing","tracing"},
	"api": {"api","apis","rest api","rest apis","web api","web apis"}, "rest": {"rest","restful","rest api","rest apis"},
	"grpc": {"grpc","g rpc"}, "redis": {"redis"}, "python": {"python"}, "java": {"java"},
	"ruby": {"ruby","ruby on rails","rails"},
}

func expandRoles(roles []string) []string {
	out := []string{}
	for _, role := range roles {
		out = appendUniqueText(out, role)
		for _, alias := range roleAliases[normalizeRoleText(role)] { out = appendUniqueText(out, alias) }
	}
	return out
}

func expandSkills(skills []string) []string {
	out := []string{}
	for _, skill := range skills {
		out = appendUniqueText(out, skill)
		for _, alias := range skillAliases[normalizeRoleText(skill)] { out = appendUniqueText(out, alias) }
	}
	return out
}

func uniqueNormalized(values []string) []string {
	out := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" { out = appendUniqueText(out, value) }
	}
	return out
}

func appendUniqueText(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" { return values }
	key := normalizeRoleText(value)
	for _, existing := range values {
		if normalizeRoleText(existing) == key { return values }
	}
	return append(values, value)
}
