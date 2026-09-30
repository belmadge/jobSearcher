package dedupe

import (
	"net/url"
	"net"
	"regexp"
	"strings"

	"jobsearcher/internal/domain"
)

var tokens = regexp.MustCompile(`[a-z0-9]+`)
var accent = strings.NewReplacer("\u00e1", "a", "\u00e0", "a", "\u00e2", "a", "\u00e3", "a", "\u00e9", "e", "\u00ea", "e", "\u00ed", "i", "\u00f3", "o", "\u00f4", "o", "\u00f5", "o", "\u00fa", "u", "\u00e7", "c")

func CanonicalURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" { return "" }
	u, err := url.Parse(raw)
	if err != nil { return raw }
	u.Fragment = ""
	u.Scheme = strings.ToLower(u.Scheme)
	host, port := strings.ToLower(u.Hostname()), u.Port()
	host = strings.TrimPrefix(host, "www.")
	if (u.Scheme == "https" && port == "443") || (u.Scheme == "http" && port == "80") { port = "" }
	if port != "" { u.Host = net.JoinHostPort(host, port) } else if strings.Contains(host, ":") { u.Host = "[" + host + "]" } else { u.Host = host }
	u.User = nil
	u.Path = strings.TrimRight(u.Path, "/")
	query := u.Query()
	for key := range query {
		lower := strings.ToLower(key)
		if strings.HasPrefix(lower, "utm_") || lower == "ref" || lower == "gclid" || lower == "fbclid" { query.Del(key) }
	}
	u.RawQuery = query.Encode()
	return u.String()
}

func NormalizeText(s string) string {
	s = accent.Replace(strings.ToLower(s))
	return strings.Join(tokens.FindAllString(s, -1), " ")
}

func FallbackKey(j domain.Job) string {
	parts := []string{NormalizeText(j.Company), NormalizeText(j.Title), NormalizeText(j.Location)}
	if parts[0]+parts[1]+parts[2] == "" { return "" }
	return strings.Join(parts, "|")
}

func SourceKey(source, sourceID string) string { if source == "" || sourceID == "" { return "" }; return strings.ToLower(source) + "|" + sourceID }
