"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import {
  analysesApi,
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
import { normalizeInsights } from "@/lib/webllm";

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

function getAnalyzeErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  return "An unknown error occurred.";
}

export default function MatchingPage() {
  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [analyses, setAnalyses] = useState<InfluencerAnalysis[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [analyzing, setAnalyzing] = useState(false);
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const [liveResult, setLiveResult] = useState<InfluencerAnalysisResult | null>(null);

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

  const selected = useMemo(
    () => scores.find((s) => s.id === selectedId) ?? null,
    [scores, selectedId],
  );

  const savedAnalysis = useMemo(
    () => (selected ? matchAnalysis(selected, analyses) : undefined),
    [selected, analyses],
  );

  async function recordMetric(
    influencerName: string,
    latencyMs: number,
    status: "success" | "error",
  ) {
    try {
      await monitoringApi.recordMetric({
        influencer_name: influencerName,
        latency_ms: latencyMs,
        status,
        model: SERVER_LLM_MODEL_ID,
      });
    } catch {
      // Do not interrupt the analysis flow if metric recording fails
    }
  }

  async function handleAnalyze() {
    if (!selected || analyzing) return;

    setAnalyzing(true);
    setAnalysisError(null);
    setLiveResult(null);

    const startTime = performance.now();

    try {
      const { result, raw_output: rawOutput } = await llmApi.analyze({
        influencer_name: selected.influencer_name,
        platform: selected.platform,
        notes: selected.notes,
      });

      const latencyMs = Math.round(performance.now() - startTime);
      await recordMetric(selected.influencer_name, latencyMs, "success");

      const updated = await scoresApi.update(selected.id, {
        overall_score: result.overall_score,
        engagement_score: result.engagement_score,
        audience_score: result.audience_score,
        brand_fit_score: result.brand_fit_score,
      });

      const created = await analysesApi.create({
        influencer_name: selected.influencer_name,
        platform: selected.platform,
        analysis_type: "mlc-llm",
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
    } catch (err) {
      const latencyMs = Math.round(performance.now() - startTime);
      await recordMetric(selected.influencer_name, latencyMs, "error");
      setAnalysisError(getAnalyzeErrorMessage(err));
    } finally {
      setAnalyzing(false);
    }
  }

  function handleSelect(id: string) {
    setSelectedId(id);
    setLiveResult(null);
    setAnalysisError(null);
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
          Server-side influencer analysis via MLC-LLM ({SERVER_LLM_MODEL_ID})
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
                    disabled={analyzing}
                    className="rounded-lg bg-[var(--accent)] px-5 py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
                  >
                    {analyzing ? "Analyzing on server..." : "Analyze"}
                  </button>
                </div>

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
                        <span className="ml-2 normal-case text-[var(--accent)]">· new</span>
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
