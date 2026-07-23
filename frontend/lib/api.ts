import { clearAuth, type AuthUser } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8081";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public body?: unknown,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type ApiFetchOptions = RequestInit & {
  token?: string | null;
  auth?: boolean;
};

function extractErrorMessage(data: unknown, status: number): string {
  if (typeof data === "object" && data !== null) {
    const body = data as { message?: string; error?: string };
    if (body.message) return body.message;
    if (body.error) return body.error;
  }
  return `Request failed (${status})`;
}

export async function apiFetch<T>(
  path: string,
  options: ApiFetchOptions = {},
): Promise<T> {
  const { token, auth = true, headers: initHeaders, ...rest } = options;

  const headers = new Headers(initHeaders);
  if (!headers.has("Content-Type") && rest.body) {
    headers.set("Content-Type", "application/json");
  }

  if (auth) {
    const authToken =
      token ??
      (typeof window !== "undefined" ? localStorage.getItem("token") : null);
    if (authToken) {
      headers.set("Authorization", `Bearer ${authToken}`);
    }
  }

  const res = await fetch(`${API_BASE}${path}`, {
    ...rest,
    headers,
  });

  const contentType = res.headers.get("content-type");
  const data = contentType?.includes("application/json")
    ? await res.json()
    : await res.text();

  if (!res.ok) {
    throw new ApiError(extractErrorMessage(data, res.status), res.status, data);
  }

  return data as T;
}

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export function handleUnauthorizedRedirect(currentPath?: string) {
  clearAuth();
  const params = currentPath
    ? `?redirect=${encodeURIComponent(currentPath)}`
    : "";
  window.location.href = `/login${params}`;
}

export type LoginPayload = { email: string; password: string };
export type RegisterPayload = LoginPayload & {
  first_name: string;
  last_name: string;
};

export type UserResponse = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  status?: string;
  created_at?: string;
};

export type LoginResponse = {
  token: string;
  user: UserResponse;
};

export function toAuthUser(user: UserResponse): AuthUser {
  return {
    id: user.id,
    email: user.email,
    first_name: user.first_name,
    last_name: user.last_name,
  };
}

export const authApi = {
  login: (payload: LoginPayload) =>
    apiFetch<LoginResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
      auth: false,
    }),

  register: (payload: RegisterPayload) =>
    apiFetch<UserResponse>("/api/v1/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
      auth: false,
    }),
};

export type InfluencerScore = {
  id: string;
  user_id: string;
  influencer_name: string;
  platform: string;
  overall_score: number;
  engagement_score: number;
  audience_score: number;
  brand_fit_score: number;
  raw_payload?: string;
  notes?: string;
  created_at: string;
  updated_at: string;
};

export type InfluencerAnalysis = {
  id: string;
  user_id: string;
  influencer_name: string;
  platform: string;
  analysis_type: string;
  summary: string;
  insights: string;
  raw_llm_output?: string;
  score_id?: string;
  created_at: string;
  updated_at: string;
};

export type CreateScorePayload = {
  influencer_name: string;
  platform: string;
  overall_score?: number;
  engagement_score?: number;
  audience_score?: number;
  brand_fit_score?: number;
  notes?: string;
};

export type UpdateScorePayload = {
  overall_score?: number;
  engagement_score?: number;
  audience_score?: number;
  brand_fit_score?: number;
  notes?: string;
};

export const scoresApi = {
  list: () =>
    apiFetch<{ scores: InfluencerScore[]; count: number }>("/api/v1/scores"),
  getById: (id: string) =>
    apiFetch<{ score: InfluencerScore }>(`/api/v1/scores/${id}`),
  create: (payload: CreateScorePayload) =>
    apiFetch<{ score: InfluencerScore }>("/api/v1/scores", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
  update: (id: string, payload: UpdateScorePayload) =>
    apiFetch<{ score: InfluencerScore }>(`/api/v1/scores/${id}`, {
      method: "PUT",
      body: JSON.stringify(payload),
    }),
};

export const analysesApi = {
  list: () =>
    apiFetch<{ analyses: InfluencerAnalysis[]; count: number }>(
      "/api/v1/analyses",
    ),

  create: (payload: {
    influencer_name: string;
    platform: string;
    analysis_type: string;
    summary: string;
    insights?: string;
    raw_llm_output?: string;
    score_id?: string;
  }) =>
    apiFetch<{ analysis: InfluencerAnalysis }>("/api/v1/analyses", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};

export type LLMCallMetric = {
  id: string;
  influencer_name: string;
  latency_ms: number;
  status: "success" | "error";
  model: string;
  timestamp: number;
};

export type MonitoringStats = {
  total_calls: number;
  avg_latency_ms: number;
  error_rate: number;
  recent_calls: LLMCallMetric[];
};

export const monitoringApi = {
  getStats: () => apiFetch<MonitoringStats>("/api/v1/monitoring/stats"),

  recordMetric: (payload: {
    influencer_name: string;
    latency_ms: number;
    status: "success" | "error";
    model: string;
  }) =>
    apiFetch<{ message: string }>("/api/v1/llm-metrics", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};

export const SERVER_LLM_MODEL_ID = "gemma-2b-it-q4f16_1-MLC";

export type InfluencerAnalysisResult = {
  overall_score: number;
  engagement_score: number;
  audience_score: number;
  brand_fit_score: number;
  summary: string;
  insights: string[];
};

export const llmApi = {
  analyze: (payload: {
    influencer_name: string;
    platform: string;
    notes?: string;
  }) =>
    apiFetch<{ result: InfluencerAnalysisResult; raw_output: string }>(
      "/api/v1/llm/analyze",
      {
        method: "POST",
        body: JSON.stringify(payload),
      },
    ),
};
