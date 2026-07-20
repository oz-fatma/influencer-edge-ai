import { clearAuth } from "./auth";

const API_BASE =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

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
      (typeof window !== "undefined" ? localStorage.getItem("access_token") : null);
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
    const message =
      typeof data === "object" && data !== null && "error" in data
        ? String((data as { error: string }).error)
        : `Request failed (${res.status})`;
    throw new ApiError(message, res.status, data);
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
export type RegisterPayload = LoginPayload & { name: string };

export type AuthResponse = {
  user: { id: number; email: string; name: string };
  tokens: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
    token_type: string;
  };
};

export const authApi = {
  login: (payload: LoginPayload) =>
    apiFetch<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(payload),
      auth: false,
    }),

  register: (payload: RegisterPayload) =>
    apiFetch<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(payload),
      auth: false,
    }),

  logout: (refreshToken: string) =>
    apiFetch<{ message: string }>("/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refresh_token: refreshToken }),
      auth: false,
    }),
};

export type InfluencerScore = {
  id: number;
  user_id: number;
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
  id: number;
  user_id: number;
  influencer_name: string;
  platform: string;
  analysis_type: string;
  summary: string;
  insights: string;
  raw_llm_output?: string;
  score_id?: number;
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

export const scoresApi = {
  list: () =>
    apiFetch<{ scores: InfluencerScore[]; count: number }>("/api/scores"),
  getById: (id: number) =>
    apiFetch<{ score: InfluencerScore }>(`/api/scores/${id}`),
  create: (payload: CreateScorePayload) =>
    apiFetch<{ score: InfluencerScore }>("/api/scores", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};

export const analysesApi = {
  list: () =>
    apiFetch<{ analyses: InfluencerAnalysis[]; count: number }>("/api/analyses"),

  create: (payload: {
    influencer_name: string;
    platform: string;
    analysis_type: string;
    summary: string;
    insights?: string;
    raw_llm_output?: string;
    score_id?: number;
  }) =>
    apiFetch<{ analysis: InfluencerAnalysis }>("/api/analyses", {
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
  getStats: () => apiFetch<MonitoringStats>("/api/monitoring/stats"),

  recordMetric: (payload: {
    influencer_name: string;
    latency_ms: number;
    status: "success" | "error";
    model: string;
  }) =>
    apiFetch<{ message: string }>("/api/llm-metrics", {
      method: "POST",
      body: JSON.stringify(payload),
    }),
};
