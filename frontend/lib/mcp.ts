import { apiFetch, SERVER_LLM_ANALYZE_TIMEOUT_MS } from "./api";

export interface MCPPayload {
  request_type: string;
  context: Record<string, unknown>;
  query: string;
}

export interface RichResult {
  data: Record<string, unknown>;
  metadata: Record<string, unknown>;
  source: string;
}

export async function sendMCPRequest(
  payload: MCPPayload,
  token?: string | null,
  timeoutMs: number = SERVER_LLM_ANALYZE_TIMEOUT_MS,
): Promise<RichResult> {
  return apiFetch<RichResult>("/api/v1/mcp/request", {
    method: "POST",
    body: JSON.stringify(payload),
    token,
    timeoutMs,
  });
}
