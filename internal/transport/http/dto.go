package http

type ShortenRequest struct {
	URL     string `json:"url"`
	Alias   string `json:"alias,omitempty"`
	TTLDays int    `json:"ttl_days,omitempty"`
}

type ShortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

type StatsResponse struct {
	URL       string  `json:"url"`
	CreatedAt string  `json:"created_at"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	HitCount  int     `json:"hit_count"`
}
