package domain

type LocationEligibility string

const (
	LocationEligible LocationEligibility = "true"
	LocationRejected LocationEligibility = "false"
	LocationUnknown  LocationEligibility = "unknown"
)

type RecommendationStatus string

const (
	Recommended      RecommendationStatus = "recommended"
	PossibleMatch    RecommendationStatus = "possible_match"
	LowMatch         RecommendationStatus = "low_match"
	RejectedLocation RecommendationStatus = "rejected_location"
	UncertainLocation RecommendationStatus = "uncertain_location"
)

type Job struct {
	ID string `json:"id"`
	Source string `json:"source"`
	Title string `json:"title"`
	Company string `json:"company"`
	URL string `json:"url"`
	ApplyURL string `json:"apply_url,omitempty"`
	CanonicalURL string `json:"canonical_url"`
	Location string `json:"location"`
	WorkplaceType string `json:"workplace_type"`
	EmploymentType string `json:"employment_type,omitempty"`
	Seniority string `json:"seniority,omitempty"`
	Description string `json:"description"`
	Requirements string `json:"requirements,omitempty"`
	Salary string `json:"salary,omitempty"`
	PostedAt string `json:"posted_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	FetchedAt string `json:"fetched_at,omitempty"`
	LocationEligible LocationEligibility `json:"location_eligible"`
	LocationReason string `json:"location_reason,omitempty"`
	RecommendationStatus RecommendationStatus `json:"recommendation_status"`
	Status string `json:"status"`
	FitScore int `json:"fit_score"`
	RoleMatch int `json:"role_match"`
	SkillMatch int `json:"skill_match"`
	TechnicalMatch int `json:"technical_match"`
	ResponsibilityMatch int `json:"responsibility_match"`
	ExperienceMatch int `json:"experience_match"`
	SeniorityMatch int `json:"seniority_match"`
	SeniorityReason string `json:"seniority_reason,omitempty"`
	CloudMatch int `json:"cloud_match"`
	DomainMatch int `json:"domain_match"`
	LanguageMatch int `json:"language_match"`
	AIMatch int `json:"ai_match"`
	MustHaveMatch []string `json:"must_have_match"`
	MustHaveMissing []string `json:"must_have_missing"`
	NiceToHaveMatch []string `json:"nice_to_have_match"`
	Gaps []string `json:"gaps"`
	Reasons []string `json:"reasons"`
	FirstSeenAt string `json:"first_seen_at,omitempty"`
	LastSeenAt string `json:"last_seen_at,omitempty"`
	SeenCount int `json:"seen_count,omitempty"`
	IsNew bool `json:"is_new"`
}
