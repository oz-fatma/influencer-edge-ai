export const NICHE_KEYS = [
  "beautyAndSkincare",
  "fashion",
  "fitnessAndHealth",
  "foodAndRecipes",
  "travel",
  "tech",
  "homeDecor",
  "parenting",
  "comedy",
  "gaming",
] as const;

export const NICHE_VALUES: Record<(typeof NICHE_KEYS)[number], string> = {
  beautyAndSkincare: "Beauty and skincare",
  fashion: "Fashion",
  fitnessAndHealth: "Fitness and health",
  foodAndRecipes: "Food and recipes",
  travel: "Travel",
  tech: "Tech",
  homeDecor: "Home decor",
  parenting: "Parenting",
  comedy: "Comedy",
  gaming: "Gaming",
};

export const AUDIENCE_GEO_KEYS = ["turkey", "usa", "europe", "global"] as const;

export const AUDIENCE_GEO_VALUES: Record<(typeof AUDIENCE_GEO_KEYS)[number], string> = {
  turkey: "Türkiye",
  usa: "ABD",
  europe: "Avrupa",
  global: "Global",
};

export const AUDIENCE_DEMO_KEYS = [
  "women1824",
  "women2534",
  "men1824",
  "men2534",
  "mixed1834",
  "mixed35plus",
] as const;

export const AUDIENCE_DEMO_VALUES: Record<(typeof AUDIENCE_DEMO_KEYS)[number], string> = {
  women1824: "Kadın 18-24",
  women2534: "Kadın 25-34",
  men1824: "Erkek 18-24",
  men2534: "Erkek 25-34",
  mixed1834: "Karma 18-34",
  mixed35plus: "Karma 35+",
};

export const FOLLOWER_RANGE_KEYS = [
  "10k50k",
  "50k100k",
  "100k500k",
  "500k1m",
  "1mPlus",
] as const;

export const FOLLOWER_RANGE_VALUES: Record<(typeof FOLLOWER_RANGE_KEYS)[number], string> = {
  "10k50k": "10K-50K",
  "50k100k": "50K-100K",
  "100k500k": "100K-500K",
  "500k1m": "500K-1M",
  "1mPlus": "1M+",
};

export const CONTENT_FORMAT_KEYS = ["reels", "story", "post", "video"] as const;

export const CONTENT_FORMAT_VALUES: Record<(typeof CONTENT_FORMAT_KEYS)[number], string> = {
  reels: "Reels",
  story: "Story",
  post: "Post",
  video: "Video",
};

export const PLATFORM_KEYS = [
  "instagram",
  "tiktok",
  "youtube",
  "twitter",
  "linkedin",
  "other",
] as const;

/** @deprecated Use NICHE_KEYS + NICHE_VALUES; kept for API canonical strings */
export const NICHES = NICHE_KEYS.map((key) => NICHE_VALUES[key]);

/** @deprecated Use AUDIENCE_GEO_KEYS + AUDIENCE_GEO_VALUES */
export const AUDIENCE_GEO = AUDIENCE_GEO_KEYS.map((key) => AUDIENCE_GEO_VALUES[key]);

/** @deprecated Use AUDIENCE_DEMO_KEYS + AUDIENCE_DEMO_VALUES */
export const AUDIENCE_DEMO = AUDIENCE_DEMO_KEYS.map((key) => AUDIENCE_DEMO_VALUES[key]);

/** @deprecated Use FOLLOWER_RANGE_KEYS + FOLLOWER_RANGE_VALUES */
export const FOLLOWER_RANGES = FOLLOWER_RANGE_KEYS.map((key) => FOLLOWER_RANGE_VALUES[key]);

/** @deprecated Use CONTENT_FORMAT_KEYS + CONTENT_FORMAT_VALUES */
export const CONTENT_FORMATS = CONTENT_FORMAT_KEYS.map((key) => CONTENT_FORMAT_VALUES[key]);

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
