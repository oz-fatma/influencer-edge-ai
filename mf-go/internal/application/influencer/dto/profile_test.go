package dto_test

import (
	"strings"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
)

func TestBuildAnalyzePrompt_structured(t *testing.T) {
	rate := 4.2
	got := dto.BuildAnalyzePrompt(dto.InfluencerProfile{
		Niche:          "Beauty and skincare",
		AudienceGeo:    "Türkiye",
		AudienceDemo:   "Kadın 25-34",
		FollowerRange:  "100K-500K",
		EngagementRate: &rate,
		ContentFormats: []string{"Reels", "Story"},
	}, "yerel kozmetik markaları")

	for _, want := range []string{
		"Niche: Beauty and skincare",
		"Audience: Türkiye",
		"Demographics: Kadın 25-34",
		"Followers: 100K-500K",
		"Engagement: 4.2%",
		"Content: Reels, Story",
		"Past collaborations: yerel kozmetik markaları",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q: %s", want, got)
		}
	}
}

func TestBuildAnalyzePrompt_legacyNotesOnly(t *testing.T) {
	got := dto.BuildAnalyzePrompt(dto.InfluencerProfile{}, "Beauty & lifestyle niche")
	if got != "Beauty & lifestyle niche" {
		t.Fatalf("got %q", got)
	}
}

func TestValidateCreateProfile_requiresContentFormats(t *testing.T) {
	err := dto.ValidateCreateProfile(dto.CreateScoreRequest{
		Niche:          "Beauty and skincare",
		AudienceGeo:    "Türkiye",
		AudienceDemo:   "Kadın 25-34",
		FollowerRange:  "100K-500K",
		EngagementRate: 4.2,
		ContentFormats: []string{},
	})
	if err == nil {
		t.Fatal("expected validation error for empty content_formats")
	}
}

func TestValidateCreateProfile_ok(t *testing.T) {
	err := dto.ValidateCreateProfile(dto.CreateScoreRequest{
		Niche:          "Tech",
		AudienceGeo:    "Global",
		AudienceDemo:   "Karma 18-34",
		FollowerRange:  "50K-100K",
		EngagementRate: 3.5,
		ContentFormats: []string{"Video"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
