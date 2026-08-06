package dto_test

import (
	"strings"
	"testing"

	"github.com/masterfabric-go/masterfabric/internal/application/influencer/dto"
)

func TestValidateCreateBrandProfile_valid(t *testing.T) {
	err := dto.ValidateCreateBrandProfile(dto.CreateBrandProfileRequest{
		Name:           "Runfit",
		Industry:       "Fitness ve aktif yaşam",
		TargetAudience: "18-30 yaş, kadın+erkek, büyük şehir",
		BudgetRange:    "5.000-15.000 TL",
		BrandValues:    "Enerjik, motive edici, samimi",
		CampaignGoal:   "Marka bilinirliği (awareness)",
	})
	if err != nil {
		t.Fatalf("valid profile: %v", err)
	}
}

func TestValidateCreateBrandProfile_requiresFields(t *testing.T) {
	err := dto.ValidateCreateBrandProfile(dto.CreateBrandProfileRequest{
		Name: "Runfit",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestValidateCreateBrandProfile_budgetRangeOptional(t *testing.T) {
	err := dto.ValidateCreateBrandProfile(dto.CreateBrandProfileRequest{
		Name:           "Runfit",
		Industry:       "Fitness",
		TargetAudience: "18-30",
		BrandValues:    "Enerjik",
		CampaignGoal:   "Awareness",
	})
	if err != nil {
		t.Fatalf("budget_range optional: %v", err)
	}
}

func TestValidateCreateBrandProfile_rejectsLongName(t *testing.T) {
	err := dto.ValidateCreateBrandProfile(dto.CreateBrandProfileRequest{
		Name:           strings.Repeat("a", 256),
		Industry:       "Fitness",
		TargetAudience: "18-30",
		BrandValues:    "Enerjik",
		CampaignGoal:   "Awareness",
	})
	if err == nil {
		t.Fatal("expected validation error for long name")
	}
}

func TestBuildBrandContext_formatsPromptSection(t *testing.T) {
	got := dto.BuildBrandContext(
		"Fitness and active lifestyle",
		"18-30, major cities",
		"5,000-15,000 TL",
		"Energetic, motivating",
		"Brand awareness",
	)
	for _, want := range []string{
		"Brand Context:",
		"- Industry: Fitness and active lifestyle",
		"- Target Audience: 18-30, major cities",
		"- Budget: 5,000-15,000 TL",
		"- Brand Values: Energetic, motivating",
		"- Campaign Goal: Brand awareness",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %q", want, got)
		}
	}
}

func TestAppendBrandContext_addsAfterQuestion(t *testing.T) {
	notes := "Influencer profile: Niche: Fitness\n\nQuestion: Is this a good fit?"
	brand := dto.BuildBrandContext("Fitness", "18-30", "", "Energetic", "Awareness")
	got := dto.AppendBrandContext(notes, brand)
	if !strings.Contains(got, "Question: Is this a good fit?") {
		t.Fatalf("missing question: %q", got)
	}
	if !strings.Contains(got, "Brand Context:") {
		t.Fatalf("missing brand context: %q", got)
	}
	idxQ := strings.Index(got, "Question:")
	idxB := strings.Index(got, "Brand Context:")
	if idxQ < 0 || idxB < 0 || idxB <= idxQ {
		t.Fatalf("brand context should come after question: %q", got)
	}
}
