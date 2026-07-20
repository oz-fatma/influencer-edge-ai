import type { InitProgressCallback, MLCEngine } from "@mlc-ai/web-llm";

/** Smallest fast Gemma model in WebLLM prebuilt list. */
export const WEBLLM_MODEL_ID = "gemma-2b-it-q4f16_1-MLC";

const ANALYSIS_JSON_SCHEMA = JSON.stringify({
  type: "object",
  properties: {
    overall_score: { type: "number" },
    engagement_score: { type: "number" },
    audience_score: { type: "number" },
    brand_fit_score: { type: "number" },
    summary: { type: "string" },
    insights: {
      type: "array",
      description:
        "Plain text insight sentences only. Each item MUST be a string, NOT an object.",
      items: {
        type: "string",
        description: "One insight sentence in plain text, e.g. 'Average 4.1% engagement over the last 30 days'",
      },
      minItems: 1,
    },
  },
  required: [
    "overall_score",
    "engagement_score",
    "audience_score",
    "brand_fit_score",
    "summary",
    "insights",
  ],
  additionalProperties: false,
});

const FEW_SHOT_EXAMPLE = JSON.stringify(
  {
    overall_score: 82.5,
    engagement_score: 88,
    audience_score: 79,
    brand_fit_score: 80,
    summary:
      "Strong engagement rate and audience alignment. Suitable profile for cosmetics campaigns.",
    insights: [
      "Average 4.1% engagement over the last 30 days",
      "65% of followers match the target demographic",
      "Sponsored content reaches 90% of organic performance",
    ],
  },
  null,
  2,
);

export type InfluencerInput = {
  name: string;
  platform: string;
  notes?: string;
};

export type InfluencerAnalysisResult = {
  overall_score: number;
  engagement_score: number;
  audience_score: number;
  brand_fit_score: number;
  summary: string;
  insights: string[];
};

export type AnalyzeInfluencerOutput = {
  result: InfluencerAnalysisResult;
  rawOutput: string;
};

let engineInstance: MLCEngine | null = null;
let engineInitPromise: Promise<MLCEngine> | null = null;

export async function isWebGPUAvailable(): Promise<boolean> {
  if (typeof navigator === "undefined") {
    return false;
  }
  const gpu = (navigator as Navigator & {
    gpu?: { requestAdapter: () => Promise<unknown> };
  }).gpu;
  if (!gpu) {
    return false;
  }
  try {
    const adapter = await gpu.requestAdapter();
    return adapter !== null;
  } catch {
    return false;
  }
}

export async function initWebLLMEngine(
  onProgress?: InitProgressCallback,
): Promise<MLCEngine> {
  if (typeof window === "undefined") {
    throw new Error("WebLLM only runs in the browser.");
  }

  const webgpuOk = await isWebGPUAvailable();
  if (!webgpuOk) {
    throw new Error(
      "WebGPU is not supported. Use a recent version of Chrome or Edge and ensure hardware acceleration is enabled.",
    );
  }

  if (engineInstance) {
    return engineInstance;
  }

  if (!engineInitPromise) {
    engineInitPromise = (async () => {
      const { CreateMLCEngine } = await import("@mlc-ai/web-llm");
      const engine = await CreateMLCEngine(WEBLLM_MODEL_ID, {
        initProgressCallback: (report) => {
          onProgress?.(report);
        },
      });
      engineInstance = engine;
      return engine;
    })();
  }

  return engineInitPromise;
}

function buildPrompt(input: InfluencerInput): string {
  const notes = input.notes?.trim() || "No notes provided";
  return `You are an influencer marketing analyst.

CRITICAL OUTPUT RULES:
- ONLY return valid JSON
- NO markdown
- NO explanation
- NO code fences
- NO text before or after the JSON object
- "insights" MUST be a string[] (plain text array), NOT an array of objects
  WRONG: [{"insights": "text"}] or [{"text": "..."}]
  CORRECT: ["Average 4.1% engagement over the last 30 days", "Audience matches the target demographic"]

Example input:
Influencer: Jane Smith
Platform: instagram
Notes: Beauty & lifestyle niche

Example output:
${FEW_SHOT_EXAMPLE}

Now analyze this influencer and return JSON in the exact same format:

Influencer: ${input.name}
Platform: ${input.platform}
Notes: ${notes}`;
}

/** Remove markdown code fence markers from model output. */
function stripCodeFences(text: string): string {
  return text
    .replace(/^\s*```(?:json)?\s*/i, "")
    .replace(/\s*```\s*$/i, "")
    .replace(/```(?:json)?/gi, "")
    .trim();
}

/** Extract the first JSON object substring between outermost { and }. */
function extractJsonObject(text: string): string | null {
  const match = text.match(/\{[\s\S]*\}/);
  return match?.[0] ?? null;
}

function collectJsonCandidates(raw: string): string[] {
  const stripped = stripCodeFences(raw);
  const fromBraces = extractJsonObject(stripped) ?? extractJsonObject(raw);

  const candidates = [stripped, fromBraces, raw.trim()].filter(
    (value): value is string => Boolean(value),
  );

  return [...new Set(candidates)];
}

function clampScore(value: unknown): number {
  const n = Number(value);
  if (Number.isNaN(n)) return 0;
  return Math.min(100, Math.max(0, Math.round(n * 10) / 10));
}

function stripQuotes(text: string): string {
  return text.trim().replace(/^['"]+|['"]+$/g, "").trim();
}

function extractTextFromInsightObject(obj: Record<string, unknown>): string | null {
  const keys = ["insights", "text", "value", "insight", "content", "message"];
  for (const key of keys) {
    const val = obj[key];
    if (typeof val === "string" && val.trim()) {
      return stripQuotes(val);
    }
  }
  const values = Object.values(obj);
  if (values.length === 1 && typeof values[0] === "string") {
    return stripQuotes(values[0]);
  }
  return null;
}

/** Normalize a single insight item to plain display text. */
export function normalizeInsightItem(item: unknown): string | null {
  if (item == null) return null;

  if (typeof item === "string") {
    const trimmed = item.trim();
    if (!trimmed) return null;

    // Object-literal string: {insights: '...'} or {"insights":"..."}
    if (trimmed.startsWith("{")) {
      try {
        const parsed = JSON.parse(trimmed) as Record<string, unknown>;
        const extracted = extractTextFromInsightObject(parsed);
        if (extracted) return extracted;
      } catch {
        const loose = trimmed.match(
          /insights['"]?\s*:\s*['"]([^'"]+)['"]/i,
        );
        if (loose?.[1]) return stripQuotes(loose[1]);
      }
    }

    const cleaned = stripQuotes(trimmed.replace(/^[-•*]\s*/, ""));
    return cleaned || null;
  }

  if (typeof item === "object") {
    return extractTextFromInsightObject(item as Record<string, unknown>);
  }

  const asString = stripQuotes(String(item));
  return asString || null;
}

/** Normalize insights from model/backend into plain string[]. */
export function normalizeInsights(raw: unknown): string[] {
  console.log("[webllm] Raw insights from model:", raw);

  if (raw == null) return [];

  if (typeof raw === "string") {
    const trimmed = raw.trim();
    if (!trimmed) return [];

    if (trimmed.startsWith("[") || trimmed.startsWith("{")) {
      try {
        const parsed = JSON.parse(trimmed);
        return normalizeInsights(parsed);
      } catch {
        // fall through to line split
      }
    }

    return trimmed
      .split(/\n+/)
      .map(normalizeInsightItem)
      .filter((item): item is string => Boolean(item));
  }

  if (Array.isArray(raw)) {
    return raw
      .map(normalizeInsightItem)
      .filter((item): item is string => Boolean(item));
  }

  if (typeof raw === "object") {
    const text = extractTextFromInsightObject(raw as Record<string, unknown>);
    return text ? [text] : [];
  }

  return [];
}

function normalizeParsed(parsed: Record<string, unknown>): InfluencerAnalysisResult {
  const insights = normalizeInsights(parsed.insights);
  console.log("[webllm] Normalized insights:", insights);

  const summary = String(parsed.summary ?? "").trim();
  if (!summary) {
    throw new Error("Model returned an empty summary.");
  }

  return {
    overall_score: clampScore(parsed.overall_score),
    engagement_score: clampScore(parsed.engagement_score),
    audience_score: clampScore(parsed.audience_score),
    brand_fit_score: clampScore(parsed.brand_fit_score),
    summary,
    insights,
  };
}

function parseAnalysisJson(raw: string): InfluencerAnalysisResult {
  const candidates = collectJsonCandidates(raw);
  const errors: string[] = [];

  for (const jsonStr of candidates) {
    try {
      const parsed = JSON.parse(jsonStr) as Record<string, unknown>;
      return normalizeParsed(parsed);
    } catch (err) {
      errors.push(err instanceof Error ? err.message : String(err));
    }
  }

  console.log("[webllm] Raw model output (JSON parse failed):", raw);
  console.log("[webllm] Parse attempts:", errors);
  throw new Error("Model did not return valid JSON. Please try again.");
}

export async function analyzeInfluencer(
  input: InfluencerInput,
  onProgress?: InitProgressCallback,
): Promise<AnalyzeInfluencerOutput> {
  const engine = await initWebLLMEngine(onProgress);

  const response = await engine.chat.completions.create({
    messages: [
      {
        role: "system",
        content:
          "You are an expert influencer marketing analyst. ONLY return valid JSON. No markdown, no explanation, no code fences.",
      },
      { role: "user", content: buildPrompt(input) },
    ],
    temperature: 0.1,
    max_tokens: 700,
    response_format: {
      type: "json_object",
      schema: ANALYSIS_JSON_SCHEMA,
    },
  });

  const rawOutput = response.choices[0]?.message?.content ?? "";
  if (!rawOutput.trim()) {
    throw new Error("Model returned an empty response.");
  }

  const result = parseAnalysisJson(rawOutput);
  return { result, rawOutput };
}

export function getWebLLMErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "An unknown error occurred.";
}
