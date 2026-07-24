"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  analysesApi,
  ApiError,
  handleUnauthorizedRedirect,
  isUnauthorized,
  llmApi,
  monitoringApi,
  scoresApi,
  SERVER_LLM_MODEL_ID,
  type InfluencerAnalysis,
  type InfluencerAnalysisResult,
  type InfluencerScore,
} from "@/lib/api";
import { scoreColor } from "@/lib/score-utils";
import {
  analyzeInfluencer,
  getWebLLMErrorMessage,
  isWebLLMLoading,
  isWebLLMReady,
  normalizeInsights,
  WEBLLM_MODEL_ID,
} from "@/lib/webllm";

type AnalysisPhase = "idle" | "server" | "browser";

function matchAnalysis(
  score: InfluencerScore,
  analyses: InfluencerAnalysis[],
): InfluencerAnalysis | undefined {
  const name = score.influencer_name.toLowerCase();
  return analyses.find((a) => a.influencer_name.toLowerCase() === name);
}

function parseStoredInsights(insights: string): string[] {
  return normalizeInsights(insights);
}

function shouldFallbackToWebLLM(error: unknown): boolean {
  if (error instanceof ApiError) {
    if (error.status === 401 || error.status === 403 || error.status === 400) {
      return false;
    }
    if ([502, 503, 504, 524].includes(error.status)) {
      return true;
    }
    const msg = error.message.toLowerCase();
    return (
      msg.includes("524") ||
      msg.includes("llm") ||
      msg.includes("timeout") ||
      msg.includes("gateway") ||
      msg.includes("not configured")
    );
  }
  if (error instanceof TypeError) {
    return true;
  }
  if (error instanceof Error) {
    const msg = error.message.toLowerCase();
    return msg.includes("fetch") || msg.includes("network") || msg.includes("timeout");
  }
  return false;
}

export default function MatchingPage() {
  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [analyses, setAnalyses] = useState<InfluencerAnalysis[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [analyzing, setAnalyzing] = useState(false);
  const [analysisPhase, setAnalysisPhase] = useState<AnalysisPhase>("idle");
  const [analysisNotice, setAnalysisNotice] = useState<string | null>(null);
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const [liveResult, setLiveResult] = useState<InfluencerAnalysisResult | null>(null);
  const [analysisSource, setAnalysisSource] = useState<"ollama" | "web-llm" | null>(
    null,
  );

  const [engineBusy, setEngineBusy] = useState(false);
  const [modelProgress, setModelProgress] = useState<number | null>(null);
  const [modelProgressText, setModelProgressText] = useState("");

  const loadData = useCallback(async () => {
    const [scoresData, analysesData] = await Promise.all([
      scoresApi.list(),
      analysesApi.list(),
    ]);
    const list = scoresData.scores ?? [];
    setScores(list);
    setAnalyses(analysesData.analyses ?? []);
    setSelectedId((prev) => {
      if (prev !== null && list.some((s) => s.id === prev)) return prev;
      return list.length > 0 ? list[0].id : null;
    });
  }, []);

  useEffect(() => {
    async function load() {
      try {
        await loadData();
      } catch (err) {
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/matching");
          return;
        }
        setError("Failed to load data. Please try again.");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [loadData]);

  useEffect(() => {
    const syncEngineBusy = () => setEngineBusy(isWebLLMLoading());
    syncEngineBusy();
    const id = window.setInterval(syncEngineBusy, 300);
    return () => window.clearInterval(id);
  }, []);

  const selected = useMemo(
    () => scores.find((s) => s.id === selectedId) ?? null,
    [scores, selectedId],
  );

  const savedAnalysis = useMemo(
    () => (selected ? matchAnalysis(selected, analyses) : undefined),
    [selected, analyses],
  );

  async function persistAnalysis(
    result: InfluencerAnalysisResult,
    rawOutput: string,
    analysisType: "ollama" | "web-llm",
    model: string,
    startTime: number,
  ) {
    if (!selected) return;

    const latencyMs = Math.round(performance.now() - startTime);
    try {
      await monitoringApi.recordMetric({
        influencer_name: selected.influencer_name,
        latency_ms: latencyMs,
        status: "success",
        model,
      });
    } catch {
      // Do not interrupt the analysis flow if metric recording fails
    }

    const updated = await scoresApi.update(selected.id, {
      overall_score: result.overall_score,
      engagement_score: result.engagement_score,
      audience_score: result.audience_score,
      brand_fit_score: result.brand_fit_score,
    });

    const created = await analysesApi.create({
      influencer_name: selected.influencer_name,
      platform: selected.platform,
      analysis_type: analysisType,
      summary: result.summary,
      insights: result.insights.join("\n"),
      raw_llm_output: rawOutput,
      score_id: selected.id,
    });

    setScores((prev) =>
      prev.map((s) => (s.id === selected.id ? updated.score : s)),
    );
    setAnalyses((prev) => [created.analysis, ...prev]);
    setLiveResult(result);
    setAnalysisSource(analysisType);
  }

  async function runWebLLMAnalyze(startTime: number) {
    if (!selected) return;

    setAnalysisPhase("browser");
    if (!isWebLLMReady()) {
      setModelProgress(0);
      setModelProgressText("Preparing browser model...");
    }

    const { result, rawOutput } = await analyzeInfluencer(
      {
        name: selected.influencer_name,
        platform: selected.platform,
        notes: selected.notes,
      },
      (report) => {
        setEngineBusy(isWebLLMLoading());
        if (!isWebLLMReady() || report.progress < 1) {
          setModelProgress(Math.round(report.progress * 100));
          setModelProgressText(report.text);
        }
      },
    );

    setModelProgress(null);
    await persistAnalysis(result, rawOutput, "web-llm", WEBLLM_MODEL_ID, startTime);
  }

  async function handleAnalyze() {
    if (!selected || analyzing) return;

    setAnalyzing(true);
    setAnalysisPhase("server");
    setAnalysisError(null);
    setAnalysisNotice(null);
    setLiveResult(null);
    setAnalysisSource(null);
    setModelProgress(null);

    const startTime = performance.now();

    try {
      const { result, raw_output: rawOutput } = await llmApi.analyze({
        influencer_name: selected.influencer_name,
        platform: selected.platform,
        notes: selected.notes,
      });

      await persistAnalysis(result, rawOutput, "ollama", SERVER_LLM_MODEL_ID, startTime);
    } catch (serverErr) {
      if (isUnauthorized(serverErr)) {
        handleUnauthorizedRedirect("/matching");
        return;
      }

      if (!shouldFallbackToWebLLM(serverErr)) {
        const latencyMs = Math.round(performance.now() - startTime);
        try {
          await monitoringApi.recordMetric({
            influencer_name: selected.influencer_name,
            latency_ms: latencyMs,
            status: "error",
            model: SERVER_LLM_MODEL_ID,
          });
        } catch {
          // ignore
        }
        setAnalysisError(
          serverErr instanceof ApiError
            ? serverErr.message
            : getWebLLMErrorMessage(serverErr),
        );
        return;
      }

      setAnalysisNotice(
        "Server Ollama timed out or is unavailable. Continuing with browser WebLLM…",
      );

      try {
        await runWebLLMAnalyze(startTime);
      } catch (browserErr) {
        const latencyMs = Math.round(performance.now() - startTime);
        try {
          await monitoringApi.recordMetric({
            influencer_name: selected.influencer_name,
            latency_ms: latencyMs,
            status: "error",
            model: WEBLLM_MODEL_ID,
          });
        } catch {
          // ignore
        }
        setAnalysisError(getWebLLMErrorMessage(browserErr));
      }
    } finally {
      setAnalyzing(false);
      setAnalysisPhase("idle");
      setEngineBusy(isWebLLMLoading());
    }
  }

  function handleSelect(id: string) {
    setSelectedId(id);
    setLiveResult(null);
    setAnalysisError(null);
    setAnalysisNotice(null);
    setAnalysisSource(null);
  }

  const displayScores = liveResult
    ? {
        overall: liveResult.overall_score,
        engagement: liveResult.engagement_score,
        audience: liveResult.audience_score,
        brandFit: liveResult.brand_fit_score,
      }
    : selected
      ? {
          overall: selected.overall_score,
          engagement: selected.engagement_score,
          audience: selected.audience_score,
          brandFit: selected.brand_fit_score,
        }
      : null;

  const displaySummary = liveResult?.summary ?? savedAnalysis?.summary;
  const displayInsights = liveResult
    ? normalizeInsights(liveResult.insights)
    : savedAnalysis?.insights
      ? parseStoredInsights(savedAnalysis.insights)
      : [];

  const analyzeButtonLabel = (() => {
    if (!analyzing) return "Analyze";
    if (analysisPhase === "server") return "Analyzing on server...";
    if (engineBusy && !isWebLLMReady()) return "Loading browser model...";
    return "Analyzing in browser...";
  })();

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 text-[var(--muted)]">
        Loading...
      </div>
    );
  }

  if (error) {
    return (
      <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-6 py-8 text-center text-red-400">
        {error}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">AI Matching Panel</h1>
        <p className="mt-1 text-[var(--muted)]">
          Hybrid analysis: server Ollama ({SERVER_LLM_MODEL_ID}) with WebLLM
          fallback ({WEBLLM_MODEL_ID})
        </p>
      </div>

      {scores.length === 0 ? (
        <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] px-6 py-16 text-center">
          <p className="text-lg font-medium">No scores added yet</p>
          <p className="mt-2 text-sm text-[var(--muted)]">
            Add influencer scores first before running matching analysis.
          </p>
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-3">
          <aside className="space-y-2 lg:col-span-1">
            <p className="mb-3 text-xs font-medium uppercase tracking-wider text-[var(--muted)]">
              Select Influencer
            </p>
            {scores.map((s) => (
              <button
                key={s.id}
                onClick={() => handleSelect(s.id)}
                className={`w-full rounded-xl border px-4 py-3 text-left transition-all ${
                  selectedId === s.id
                    ? "border-[var(--accent)]/50 bg-[var(--accent-muted)]"
                    : "border-[var(--border)] bg-[var(--surface)] hover:border-[var(--accent)]/20"
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">{s.influencer_name}</span>
                  <span className="text-sm font-semibold text-[var(--accent)]">
                    {s.overall_score}
                  </span>
                </div>
                <span className="text-xs capitalize text-[var(--muted)]">{s.platform}</span>
              </button>
            ))}
          </aside>

          {selected && displayScores && (
            <section className="space-y-5 lg:col-span-2">
              <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <h2 className="text-xl font-bold">{selected.influencer_name}</h2>
                    <p className="text-sm capitalize text-[var(--muted)]">{selected.platform}</p>
                    {selected.notes && (
                      <p className="mt-2 text-xs text-[var(--muted)]">{selected.notes}</p>
                    )}
                  </div>
                  <button
                    onClick={handleAnalyze}
                    disabled={analyzing || engineBusy}
                    className="rounded-lg bg-[var(--accent)] px-5 py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
                  >
                    {analyzeButtonLabel}
                  </button>
                </div>

                {modelProgress !== null && (
                  <div className="mt-5">
                    <div className="mb-1.5 flex items-center justify-between text-xs text-[var(--muted)]">
                      <span>{modelProgressText || "Loading browser model..."}</span>
                      <span>{modelProgress}%</span>
                    </div>
                    <div className="h-2 overflow-hidden rounded-full bg-[var(--surface-elevated)]">
                      <div
                        className="h-full rounded-full bg-[var(--accent)] transition-all duration-300"
                        style={{ width: `${modelProgress}%` }}
                      />
                    </div>
                  </div>
                )}

                <div className="mt-5 grid grid-cols-2 gap-3 sm:grid-cols-4">
                  <ScorePill
                    label={liveResult ? "Overall (AI)" : "Overall"}
                    value={displayScores.overall}
                    highlight
                  />
                  <ScorePill label="Engagement" value={displayScores.engagement} />
                  <ScorePill label="Audience" value={displayScores.audience} />
                  <ScorePill label="Brand Fit" value={displayScores.brandFit} />
                </div>
              </div>

              {analysisNotice && (
                <div className="rounded-xl border border-amber-500/25 bg-amber-500/10 px-5 py-4 text-sm text-amber-200">
                  {analysisNotice}
                </div>
              )}

              {analysisError && (
                <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-5 py-4 text-sm text-red-400">
                  {analysisError}
                </div>
              )}

              {displaySummary ? (
                <>
                  <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                    <h3 className="mb-2 text-sm font-semibold uppercase tracking-wider text-[var(--muted)]">
                      AI Summary
                      {liveResult && (
                        <span className="ml-2 normal-case text-[var(--accent)]">
                          · new
                          {analysisSource === "web-llm" ? " (browser)" : " (server)"}
                        </span>
                      )}
                    </h3>
                    <p className="leading-relaxed">{displaySummary}</p>
                  </div>

                  {displayInsights.length > 0 && (
                    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                      <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-[var(--muted)]">
                        Insights
                      </h3>
                      <ul className="space-y-2">
                        {displayInsights.map((insight, i) => (
                          <li key={i} className="flex gap-2 text-sm">
                            <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--accent)]" />
                            {insight}
                          </li>
                        ))}
                      </ul>
                    </div>
                  )}
                </>
              ) : (
                !analysisError &&
                !analyzing && (
                  <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] p-8 text-center">
                    <p className="font-medium text-[var(--muted)]">
                      No analysis yet
                    </p>
                    <p className="mt-2 text-sm text-[var(--muted)]">
                      Click &quot;Analyze&quot; for the selected influencer.
                    </p>
                  </div>
                )
              )}
            </section>
          )}
        </div>
      )}
    </div>
  );
}

function ScorePill({
  label,
  value,
  highlight,
}: {
  label: string;
  value: number;
  highlight?: boolean;
}) {
  return (
    <div
      className={`rounded-lg px-3 py-3 text-center ${
        highlight ? "bg-[var(--accent-muted)]" : "bg-[var(--surface-elevated)]"
      }`}
    >
      <p className="text-[10px] uppercase tracking-wider text-[var(--muted)]">{label}</p>
      <p
        className={`text-lg font-bold ${highlight ? "text-[var(--accent)]" : scoreColor(value)}`}
      >
        {value}
      </p>
    </div>
  );
}
