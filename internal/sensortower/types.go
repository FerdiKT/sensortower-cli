package sensortower

import (
	"encoding/json"
	"time"
)

type PublisherAppsResponse struct {
	Meta PublisherAppsMeta `json:"meta"`
	Data []PublisherApp    `json:"data"`
}

type PublisherAppsMeta struct {
	Count int `json:"count"`
}

type HumanizedMetric struct {
	Downloads        int64   `json:"downloads,omitempty"`
	Revenue          int64   `json:"revenue,omitempty"`
	DownloadsRounded float64 `json:"downloads_rounded,omitempty"`
	RevenueRounded   float64 `json:"revenue_rounded,omitempty"`
	Prefix           string  `json:"prefix,omitempty"`
	String           string  `json:"string,omitempty"`
	Units            string  `json:"units,omitempty"`
}

type PublisherApp struct {
	AppID                        int64           `json:"app_id"`
	ID                           int64           `json:"id"`
	Name                         string          `json:"name"`
	HumanizedName                string          `json:"humanized_name"`
	PublisherName                string          `json:"publisher_name"`
	PublisherID                  int64           `json:"publisher_id"`
	CanonicalCountry             string          `json:"canonical_country"`
	IconURL                      string          `json:"icon_url"`
	OS                           string          `json:"os"`
	Active                       bool            `json:"active"`
	Price                        float64         `json:"price"`
	ReleaseDate                  string          `json:"release_date"`
	UpdatedDate                  string          `json:"updated_date"`
	ValidCountries               []string        `json:"valid_countries"`
	WorldwideLast30DaysDownloads HumanizedMetric `json:"humanized_worldwide_last_30_days_downloads"`
	WorldwideLast30DaysRevenue   HumanizedMetric `json:"humanized_worldwide_last_30_days_revenue"`
}

func (a PublisherApp) HumanizedNameOrName() string {
	if a.HumanizedName != "" {
		return a.HumanizedName
	}
	return a.Name
}

type AppDetails struct {
	AppID                       int64            `json:"app_id"`
	Name                        string           `json:"name"`
	PublisherID                 int64            `json:"publisher_id"`
	PublisherName               string           `json:"publisher_name"`
	Country                     string           `json:"country"`
	OS                          string           `json:"os"`
	CurrentVersion              string           `json:"current_version"`
	Rating                      float64          `json:"rating"`
	RatingCount                 int64            `json:"rating_count"`
	FileSize                    int64            `json:"file_size"`
	Language                    string           `json:"language"`
	TopCountries                []string         `json:"top_countries"`
	SupportURL                  string           `json:"support_url"`
	WebsiteURL                  string           `json:"website_url"`
	ReleaseDate                 int64            `json:"release_date"`
	RecentReleaseDate           int64            `json:"recent_release_date"`
	MinimumOSVersion            string           `json:"minimum_os_version"`
	Subtitle                    string           `json:"subtitle"`
	PromoText                   string           `json:"promo_text"`
	Description                 Description      `json:"description"`
	SupportedLanguages          []string         `json:"supported_languages"`
	HasInAppPurchases           bool             `json:"has_in_app_purchases"`
	WorldwideLastMonthRevenue   MonetaryMetric   `json:"worldwide_last_month_revenue"`
	WorldwideLastMonthDownloads CountMetric      `json:"worldwide_last_month_downloads"`
	CategoryRankings            CategoryRankings `json:"category_rankings"`
	Raw                         map[string]any   `json:"-"`
}

type Description struct {
	FullDescription string `json:"full_description"`
}

type MonetaryMetric struct {
	Unit     string `json:"unit"`
	Type     string `json:"type"`
	Currency string `json:"currency"`
	Value    int64  `json:"value"`
}

type CountMetric struct {
	Unit  string `json:"unit"`
	Type  string `json:"type"`
	Value int64  `json:"value"`
}

type CategoryRankings struct {
	AppID   int64               `json:"app_id"`
	Country string              `json:"country"`
	IPhone  CategoryDeviceRanks `json:"iphone"`
	IPad    CategoryDeviceRanks `json:"ipad"`
}

type CategoryDeviceRanks struct {
	TopFree     *RankGroup `json:"top_free"`
	TopGrossing *RankGroup `json:"top_grossing"`
	TopPaid     *RankGroup `json:"top_paid"`
}

type RankGroup struct {
	AllCategories     *int64              `json:"all_categories"`
	PrimaryCategories []map[string]*int64 `json:"primary_categories"`
	SubCategories     []map[string]*int64 `json:"sub_categories"`
}

type CategoryRankingsResponse struct {
	Data       CategoryRankingBuckets `json:"data"`
	Date       string                 `json:"date"`
	TotalCount int                    `json:"total_count"`
	Offset     int                    `json:"offset"`
	Limit      int                    `json:"limit"`
}

type CategoryRankingBuckets struct {
	Free     []CategoryRankingEntry `json:"free"`
	Grossing []CategoryRankingEntry `json:"grossing"`
	Paid     []CategoryRankingEntry `json:"paid"`
}

type CategoryRankingEntry struct {
	OS                          string          `json:"os"`
	PublisherID                 int64           `json:"publisher_id"`
	PublisherName               string          `json:"publisher_name"`
	InAppPurchases              bool            `json:"in_app_purchases"`
	WorldwideLastMonthDownloads HumanizedMetric `json:"humanized_worldwide_last_month_downloads"`
	WorldwideLastMonthRevenue   HumanizedMetric `json:"humanized_worldwide_last_month_revenue"`
	IconURL                     string          `json:"icon_url"`
	Rank                        int             `json:"rank"`
	PreviousRank                int             `json:"previous_rank"`
	AppID                       int64           `json:"app_id"`
	ShowToBot                   bool            `json:"show_to_bot"`
	Name                        string          `json:"name"`
	Price                       Price           `json:"price"`
	Rating                      float64         `json:"rating"`
	RatingCount                 int64           `json:"rating_count"`
	AppOverviewURL              string          `json:"app_overview_url"`
}

type ResponseMeta struct {
	Cached            bool              `json:"cached,omitempty"`
	Retried           int               `json:"retried,omitempty"`
	RetryAfterSeconds int               `json:"retry_after_seconds,omitempty"`
	RateLimitHeaders  map[string]string `json:"rate_limit_headers,omitempty"`
	RequestURL        string            `json:"request_url,omitempty"`
}

type AppsBatchResult struct {
	OK     []map[string]any   `json:"ok"`
	Failed []AppsBatchFailure `json:"failed"`
	Meta   ResponseMeta       `json:"meta"`
}

type AppsBatchFailure struct {
	AppID int64  `json:"app_id"`
	Error string `json:"error"`
}

type CompetitorRecord struct {
	AppID             int64            `json:"app_id"`
	Name              string           `json:"name"`
	PublisherName     string           `json:"publisher_name"`
	Country           string           `json:"country"`
	Categories        []int            `json:"categories"`
	Buckets           []string         `json:"buckets"`
	ObservedRanks     []map[string]any `json:"observed_ranks"`
	Enriched          map[string]any   `json:"enriched,omitempty"`
	MetadataFetchedAt time.Time        `json:"metadata_fetched_at,omitempty"`
}

type MetadataAudit struct {
	AppID    int64    `json:"app_id"`
	Name     string   `json:"name"`
	Issues   []string `json:"issues"`
	Keywords []string `json:"keywords,omitempty"`
}

type KeywordGapResult struct {
	TargetAppID      int64    `json:"target_app_id"`
	CompetitorAppIDs []int64  `json:"competitor_app_ids"`
	MissingKeywords  []string `json:"missing_keywords"`
}

type Price struct {
	Currency      string `json:"currency"`
	Value         int64  `json:"value"`
	StringValue   string `json:"string_value"`
	SubunitToUnit int64  `json:"subunit_to_unit"`
}

func (a *AppDetails) UnmarshalJSON(data []byte) error {
	type alias AppDetails
	aux := (*alias)(a)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	return json.Unmarshal(data, &a.Raw)
}
