package dto

import (
	"fmt"
	"strings"

	domainErr "github.com/masterfabric-go/masterfabric/internal/shared/errors"
)

var (
	allowedNiches = map[string]struct{}{
		"Beauty and skincare": {}, "Fashion": {}, "Fitness and health": {},
		"Food and recipes": {}, "Travel": {}, "Tech": {}, "Home decor": {},
		"Parenting": {}, "Comedy": {}, "Gaming": {},
	}
	allowedAudienceGeo = map[string]struct{}{
		"Türkiye": {}, "ABD": {}, "Avrupa": {}, "Global": {},
	}
	allowedAudienceDemo = map[string]struct{}{
		"Kadın 18-24": {}, "Kadın 25-34": {}, "Erkek 18-24": {},
		"Erkek 25-34": {}, "Karma 18-34": {}, "Karma 35+": {},
	}
	allowedFollowerRanges = map[string]struct{}{
		"10K-50K": {}, "50K-100K": {}, "100K-500K": {}, "500K-1M": {}, "1M+": {},
	}
	allowedContentFormats = map[string]struct{}{
		"Reels": {}, "Story": {}, "Post": {}, "Video": {},
	}
)

// InfluencerProfile holds structured influencer metadata for analyze prompts.
type InfluencerProfile struct {
	Niche           string
	AudienceGeo     string
	AudienceDemo    string
	FollowerRange   string
	EngagementRate  *float64
	ContentFormats  []string
	Notes           string
}

func (p InfluencerProfile) HasStructuredFields() bool {
	if strings.TrimSpace(p.Niche) != "" ||
		strings.TrimSpace(p.AudienceGeo) != "" ||
		strings.TrimSpace(p.AudienceDemo) != "" ||
		strings.TrimSpace(p.FollowerRange) != "" ||
		p.EngagementRate != nil ||
		len(p.ContentFormats) > 0 {
		return true
	}
	return false
}

// BuildAnalyzePrompt returns LLM notes text from structured profile or legacy notes.
func BuildAnalyzePrompt(profile InfluencerProfile, legacyNotes string) string {
	legacyNotes = strings.TrimSpace(legacyNotes)
	if profile.Notes != "" && legacyNotes == "" {
		legacyNotes = strings.TrimSpace(profile.Notes)
	}

	if !profile.HasStructuredFields() {
		if legacyNotes == "" {
			return "No notes provided"
		}
		return legacyNotes
	}

	var parts []string
	if v := strings.TrimSpace(profile.Niche); v != "" {
		parts = append(parts, "Niche: "+v)
	}
	if v := strings.TrimSpace(profile.AudienceGeo); v != "" {
		parts = append(parts, "Audience: "+v)
	}
	if v := strings.TrimSpace(profile.AudienceDemo); v != "" {
		parts = append(parts, "Demographics: "+v)
	}
	if v := strings.TrimSpace(profile.FollowerRange); v != "" {
		parts = append(parts, "Followers: "+v)
	}
	if profile.EngagementRate != nil {
		parts = append(parts, fmt.Sprintf("Engagement: %.1f%%", *profile.EngagementRate))
	}
	if len(profile.ContentFormats) > 0 {
		parts = append(parts, "Content: "+strings.Join(profile.ContentFormats, ", "))
	}
	if legacyNotes != "" {
		parts = append(parts, "Past collaborations: "+legacyNotes)
	}
	if len(parts) == 0 {
		return "No notes provided"
	}
	return strings.Join(parts, ", ")
}

const (
	// MaxQueryLen is the maximum length of a user analyze question (MCP query).
	MaxQueryLen = 500
	// DefaultAnalyzeQuery is used when the client sends an empty query.
	DefaultAnalyzeQuery = "Assess brand-fit and engagement potential"
)

// ValidateAnalyzeQuery checks MCP analyze question length.
func ValidateAnalyzeQuery(query string) error {
	if len(strings.TrimSpace(query)) > MaxQueryLen {
		return domainErr.New(domainErr.ErrValidation, "query must be at most 500 characters", nil)
	}
	return nil
}

// BuildAnalyzePromptWithQuery merges structured profile context and the user question.
func BuildAnalyzePromptWithQuery(profile InfluencerProfile, legacyNotes, query string) string {
	profileText := strings.TrimSpace(BuildAnalyzePrompt(profile, legacyNotes))
	query = strings.TrimSpace(query)
	if query == "" {
		query = DefaultAnalyzeQuery
	}
	if profileText == "" || profileText == "No notes provided" {
		return "Question: " + query
	}
	return "Influencer profile: " + profileText + "\n\nQuestion: " + query
}

// ProfileFromMap reads structured profile fields from an MCP-style context map.
func ProfileFromMap(ctx map[string]any) InfluencerProfile {
	p := InfluencerProfile{
		Niche:          mapString(ctx, "niche"),
		AudienceGeo:    mapString(ctx, "audience_geo"),
		AudienceDemo:   mapString(ctx, "audience_demo"),
		FollowerRange:  mapString(ctx, "follower_range"),
		Notes:          mapString(ctx, "notes"),
		ContentFormats: mapStringSlice(ctx, "content_formats"),
	}
	if rate, ok := mapFloat(ctx, "engagement_rate"); ok {
		p.EngagementRate = &rate
	}
	return p
}

func mapString(ctx map[string]any, key string) string {
	if ctx == nil {
		return ""
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return strings.TrimSpace(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func mapStringSlice(ctx map[string]any, key string) []string {
	if ctx == nil {
		return nil
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return nil
	}
	switch typed := v.(type) {
	case []string:
		return trimStringSlice(typed)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			if s := strings.TrimSpace(fmt.Sprint(item)); s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func mapFloat(ctx map[string]any, key string) (float64, bool) {
	if ctx == nil {
		return 0, false
	}
	v, ok := ctx[key]
	if !ok || v == nil {
		return 0, false
	}
	switch typed := v.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	default:
		var f float64
		if _, err := fmt.Sscan(fmt.Sprint(typed), &f); err != nil {
			return 0, false
		}
		return f, true
	}
}

func trimStringSlice(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func validateAllowed(value string, allowed map[string]struct{}, field string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return domainErr.New(domainErr.ErrValidation, field+" is required", nil)
	}
	if _, ok := allowed[value]; !ok {
		return domainErr.New(domainErr.ErrValidation, field+" has an invalid value", nil)
	}
	return nil
}

func ValidateCreateProfile(req CreateScoreRequest) error {
	if err := validateAllowed(req.Niche, allowedNiches, "niche"); err != nil {
		return err
	}
	if err := validateAllowed(req.AudienceGeo, allowedAudienceGeo, "audience_geo"); err != nil {
		return err
	}
	if err := validateAllowed(req.AudienceDemo, allowedAudienceDemo, "audience_demo"); err != nil {
		return err
	}
	if err := validateAllowed(req.FollowerRange, allowedFollowerRanges, "follower_range"); err != nil {
		return err
	}
	if req.EngagementRate < 0 || req.EngagementRate > 100 {
		return domainErr.New(domainErr.ErrValidation, "engagement_rate must be between 0 and 100", nil)
	}
	if len(req.ContentFormats) == 0 {
		return domainErr.New(domainErr.ErrValidation, "content_formats must include at least one item", nil)
	}
	seen := make(map[string]struct{}, len(req.ContentFormats))
	for _, format := range req.ContentFormats {
		format = strings.TrimSpace(format)
		if format == "" {
			return domainErr.New(domainErr.ErrValidation, "content_formats must not contain empty values", nil)
		}
		if _, ok := allowedContentFormats[format]; !ok {
			return domainErr.New(domainErr.ErrValidation, "content_formats has an invalid value", nil)
		}
		if _, dup := seen[format]; dup {
			return domainErr.New(domainErr.ErrValidation, "content_formats must not contain duplicates", nil)
		}
		seen[format] = struct{}{}
	}
	return nil
}

func ProfileFromCreateRequest(req CreateScoreRequest) InfluencerProfile {
	rate := req.EngagementRate
	return InfluencerProfile{
		Niche:          strings.TrimSpace(req.Niche),
		AudienceGeo:    strings.TrimSpace(req.AudienceGeo),
		AudienceDemo:   strings.TrimSpace(req.AudienceDemo),
		FollowerRange:  strings.TrimSpace(req.FollowerRange),
		EngagementRate: &rate,
		ContentFormats: append([]string(nil), req.ContentFormats...),
		Notes:          strings.TrimSpace(req.Notes),
	}
}

func ProfileFromScoreFields(
	niche, audienceGeo, audienceDemo, followerRange string,
	engagementRate *float64,
	contentFormats []string,
	notes string,
) InfluencerProfile {
	return InfluencerProfile{
		Niche:          niche,
		AudienceGeo:    audienceGeo,
		AudienceDemo:   audienceDemo,
		FollowerRange:  followerRange,
		EngagementRate: engagementRate,
		ContentFormats: append([]string(nil), contentFormats...),
		Notes:          notes,
	}
}
