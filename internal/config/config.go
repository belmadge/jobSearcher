package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Weights struct {
	Technical      int `json:"technical"`
	Responsibility int `json:"responsibility"`
	Seniority      int `json:"seniority"`
	Cloud          int `json:"cloud"`
	Domain         int `json:"domain"`
	Language       int `json:"language"`
	AI             int `json:"ai"`
}
type Experience struct {
	Company string `json:"company"`
	Title   string `json:"title"`
	Period  string `json:"period"`
}
type Profile struct {
	Titles         []string     `json:"titles"`
	Technologies   []string     `json:"technologies"`
	Experience     []Experience `json:"experience"`
	Domains        []string     `json:"domains"`
	EmergingSkills []string     `json:"emerging_skills"`
	Languages      []string     `json:"languages"`
	Certifications []string     `json:"certifications"`
	Weights        Weights      `json:"weights"`
}
type Search struct {
	Location         string   `json:"location"`
	RemoteAllowed    []string `json:"remote_allowed"`
	PreferredTitles  []string `json:"preferred_titles"`
	MinimumFitScore  int      `json:"minimum_fit_score"`
	GreenhouseBoards []string `json:"greenhouse_boards,omitempty"`
}

type Board struct {
	Token string `json:"token"`
	Company string `json:"company"`
	Enabled bool `json:"enabled"`
}
type Boards struct { Greenhouse []Board `json:"greenhouse"` }

// LoadDotEnv loads KEY=VALUE entries without overriding variables already set by the process.
func LoadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key, value = strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), "\"'")
		if key != "" {
			if _, exists := os.LookupEnv(key); !exists {
				if err := os.Setenv(key, value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func Load[T any](path string) (T, error) {
	var value T
	b, err := os.ReadFile(path)
	if err != nil {
		return value, fmt.Errorf("read %s: %w", path, err)
	}
	if err := json.Unmarshal(b, &value); err != nil {
		return value, fmt.Errorf("parse %s: %w", path, err)
	}
	return value, nil
}
