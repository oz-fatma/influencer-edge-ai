export const NICHES = [
  "Beauty and skincare",
  "Fashion",
  "Fitness and health",
  "Food and recipes",
  "Travel",
  "Tech",
  "Home decor",
  "Parenting",
  "Comedy",
  "Gaming",
] as const;

export const AUDIENCE_GEO = ["Türkiye", "ABD", "Avrupa", "Global"] as const;

export const AUDIENCE_DEMO = [
  "Kadın 18-24",
  "Kadın 25-34",
  "Erkek 18-24",
  "Erkek 25-34",
  "Karma 18-34",
  "Karma 35+",
] as const;

export const FOLLOWER_RANGES = [
  "10K-50K",
  "50K-100K",
  "100K-500K",
  "500K-1M",
  "1M+",
] as const;

export const CONTENT_FORMATS = ["Reels", "Story", "Post", "Video"] as const;

export const DEFAULT_ANALYZE_QUERY = "Assess brand-fit and engagement potential";
export const MAX_ANALYZE_QUERY_LEN = 500;

export type InfluencerProfileFields = {
  niche?: string;
  audience_geo?: string;
  audience_demo?: string;
  follower_range?: string;
  engagement_rate?: number;
  content_formats?: string[];
  notes?: string;
};

export function hasStructuredProfile(
  score: InfluencerProfileFields,
): boolean {
  return Boolean(
    score.niche?.trim() ||
      score.audience_geo?.trim() ||
      score.audience_demo?.trim() ||
      score.follower_range?.trim() ||
      score.engagement_rate != null ||
      (score.content_formats?.length ?? 0) > 0,
  );
}

export function buildAnalyzeNotes(
  score: InfluencerProfileFields,
): string {
  const legacyNotes = score.notes?.trim() ?? "";
  if (!hasStructuredProfile(score)) {
    return legacyNotes || "No notes provided";
  }

  const parts: string[] = [];
  if (score.niche?.trim()) parts.push(`Niche: ${score.niche.trim()}`);
  if (score.audience_geo?.trim()) {
    parts.push(`Audience: ${score.audience_geo.trim()}`);
  }
  if (score.audience_demo?.trim()) {
    parts.push(`Demographics: ${score.audience_demo.trim()}`);
  }
  if (score.follower_range?.trim()) {
    parts.push(`Followers: ${score.follower_range.trim()}`);
  }
  if (score.engagement_rate != null) {
    parts.push(`Engagement: ${score.engagement_rate.toFixed(1)}%`);
  }
  if (score.content_formats?.length) {
    parts.push(`Content: ${score.content_formats.join(", ")}`);
  }
  if (legacyNotes) {
    parts.push(`Past collaborations: ${legacyNotes}`);
  }
  return parts.join(", ");
}

/** Merges profile context and user question for LLM notes (matches backend format). */
export function buildAnalyzePromptWithQuery(
  score: InfluencerProfileFields,
  question?: string,
): string {
  const profileText = buildAnalyzeNotes(score).trim();
  const q = question?.trim() || DEFAULT_ANALYZE_QUERY;
  if (!profileText || profileText === "No notes provided") {
    return `Question: ${q}`;
  }
  return `Influencer profile: ${profileText}\n\nQuestion: ${q}`;
}

export function profileSummary(score: InfluencerProfileFields): string {
  if (!hasStructuredProfile(score)) {
    return score.notes?.trim() || "—";
  }
  const parts = [
    score.niche,
    score.audience_geo,
    score.follower_range,
  ].filter(Boolean);
  return parts.join(" · ") || "—";
}

export function buildMCPContext(score: InfluencerProfileFields & {
  influencer_name: string;
  platform: string;
}): Record<string, unknown> {
  const context: Record<string, unknown> = {
    influencer_name: score.influencer_name,
    platform: score.platform,
  };
  if (score.niche) context.niche = score.niche;
  if (score.audience_geo) context.audience_geo = score.audience_geo;
  if (score.audience_demo) context.audience_demo = score.audience_demo;
  if (score.follower_range) context.follower_range = score.follower_range;
  if (score.engagement_rate != null) {
    context.engagement_rate = score.engagement_rate;
  }
  if (score.content_formats?.length) {
    context.content_formats = score.content_formats;
  }
  if (score.notes?.trim()) context.notes = score.notes.trim();
  return context;
}
