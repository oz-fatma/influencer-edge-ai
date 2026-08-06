"use client";

import { useEffect, useMemo, useState } from "react";
import { usePathname } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import { useAnalysisContext } from "@/context/AnalysisContext";
import {
  analysesApi,
  ApiError,
  brandProfilesApi,
  handleUnauthorizedRedirect,
  isUnauthorized,
  monitoringApi,
  scoresApi,
  SERVER_LLM_MODEL_ID,
  SERVER_LLM_ANALYZE_TIMEOUT_MS,
  type BrandProfile,
  type InfluencerAnalysis,
  type InfluencerAnalysisResult,
  type InfluencerScore,
} from "@/lib/api";
import { sendMCPRequest, type RichResult } from "@/lib/mcp";
import {
  buildMCPContext,
  profileSummary,
  DEFAULT_ANALYZE_QUERY,
  MAX_ANALYZE_QUERY_LEN,
} from "@/lib/influencer-profile";
import { scoreColor } from "@/lib/score-utils";
import {
  analyzeInfluencer,
  getWebLLMErrorMessage,
  isWebLLMLoading,
  isWebLLMReady,
  normalizeInsights,
  WEBLLM_MODEL_ID,
} from "@/lib/webllm";

const MCP_ANALYZE_QUERY = DEFAULT_ANALYZE_QUERY;

function mapMCPRichResultToAnalysis(rich: RichResult): {
  result: InfluencerAnalysisResult;
  rawOutput: string;
  model: string;
} {
  const { data, metadata } = rich;
  const summary = typeof data.summary === "string" ? data.summary.trim() : "";
  if (!summary) {
    throw new ApiError("MCP response missing summary", 502);
  }

  return {
    result: {
      overall_score: Number(data.overall_score) || 0,
      engagement_score: Number(data.engagement_score) || 0,
      audience_score: Number(data.audience_score) || 0,
      brand_fit_score: Number(data.brand_fit_score) || 0,
      summary,
      insights: normalizeInsights(data.insights),
    },
    rawOutput: typeof data.raw_output === "string" ? data.raw_output : "",
    model:
      typeof metadata.model === "string" && metadata.model.trim() !== ""
        ? metadata.model
        : SERVER_LLM_MODEL_ID,
  };
}

function matchAnalysis(
  score: InfluencerScore,
  analyses: InfluencerAnalysis[],
): InfluencerAnalysis | undefined {
  const history = getAnalysisHistory(score, analyses);
  return history[0];
}

function getAnalysisHistory(
  score: InfluencerScore,
  analyses: InfluencerAnalysis[],
): InfluencerAnalysis[] {
  const name = score.influencer_name.toLowerCase();
  return analyses
    .filter((a) => a.influencer_name.toLowerCase() === name)
    .sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
}

function truncateSummary(text: string, maxLen = 100): string {
  const trimmed = text.trim();
  if (trimmed.length <= maxLen) return trimmed;
  return `${trimmed.slice(0, maxLen)}...`;
}

function formatAnalysisDate(iso: string, locale: string): string {
  const intlLocale = locale === "tr" ? "tr-TR" : "en-US";
  return new Intl.DateTimeFormat(intlLocale, {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(iso));
}

function parseStoredInsights(insights: string): string[] {
  return normalizeInsights(insights);
}

function shouldFallbackToWebLLM(error: unknown): boolean {
  if (error instanceof ApiError) {
    if (error.status === 401 || error.status === 403 || error.status === 400) {
      return false;
    }
    if ([408, 502, 503, 504, 524].includes(error.status)) {
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
  const t = useTranslations("matching");
  const locale = useLocale();
  const pathname = usePathname();
  const {
    analyzing,
    setAnalyzing,
    analysisPhase,
    setAnalysisPhase,
    selectedInfluencerId,
    setSelectedInfluencerId,
    liveResult,
    setLiveResult,
    analysisSource,
    setAnalysisSource,
    analysisError,
    setAnalysisError,
    analysisNotice,
    setAnalysisNotice,
  } = useAnalysisContext();

  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [analyses, setAnalyses] = useState<InfluencerAnalysis[]>([]);
  const [brandProfiles, setBrandProfiles] = useState<BrandProfile[]>([]);
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [engineBusy, setEngineBusy] = useState(false);
  const [modelProgress, setModelProgress] = useState<number | null>(null);
  const [modelProgressText, setModelProgressText] = useState("");
  const [userQuestion, setUserQuestion] = useState("");
  const [askedQuestion, setAskedQuestion] = useState<string | null>(null);
  const [selectedBrandProfileId, setSelectedBrandProfileId] = useState<string | null>(null);
  const [analyzedForBrandName, setAnalyzedForBrandName] = useState<string | null>(null);
  const [expandedHistoryId, setExpandedHistoryId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    async function loadData() {
      try {
        const [scoresData, analysesData, brandProfilesData] = await Promise.all([
          scoresApi.list(),
          analysesApi.list(),
          brandProfilesApi.list(),
        ]);
        if (cancelled) return;

        const list = scoresData.scores ?? [];
        setScores(list);
        setAnalyses(analysesData.analyses ?? []);
        setBrandProfiles(brandProfilesData.brand_profiles ?? []);
        setSelectedId((prev) => {
          const contextId =
            (analyzing || liveResult) && selectedInfluencerId
              ? selectedInfluencerId
              : null;
          if (contextId && list.some((s) => s.id === contextId)) {
            return contextId;
          }
          if (prev !== null && list.some((s) => s.id === prev)) return prev;
          return list.length > 0 ? list[0].id : null;
        });
        setError(null);
      } catch (err) {
        if (cancelled) return;
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/matching");
          return;
        }
        setError(t("errors.loadFailed"));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    loadData();
    return () => {
      cancelled = true;
    };
  }, [pathname, analyzing, selectedInfluencerId, liveResult]);

  const selected = useMemo(
    () => scores.find((s) => s.id === selectedId) ?? null,
    [scores, selectedId],
  );

  const savedAnalysis = useMemo(
    () => (selected ? matchAnalysis(selected, analyses) : undefined),
    [selected, analyses],
  );

  const pastAnalysisHistory = useMemo(() => {
    if (!selected) return [];
    return getAnalysisHistory(selected, analyses).slice(1);
  }, [selected, analyses]);

  async function persistAnalysis(
    result: InfluencerAnalysisResult,
    rawOutput: string,
    analysisType: "ollama" | "web-llm",
    model: string,
    startTime: number,
    brandProfileId?: string | null,
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
      ...(brandProfileId ? { brand_profile_id: brandProfileId } : {}),
    });

    setScores((prev) =>
      prev.map((s) => (s.id === selected.id ? updated.score : s)),
    );
    setAnalyses((prev) => [created.analysis, ...prev]);
    setLiveResult(result);
    setAnalysisSource(analysisType);
  }

  async function runWebLLMAnalyze(
    startTime: number,
    query: string,
    brandProfileId?: string | null,
  ) {
    if (!selected) return;

    setAnalysisPhase("browser");
    if (!isWebLLMReady()) {
      setModelProgress(0);
      setModelProgressText(t("analyze.preparingBrowser"));
    }

    const { result, rawOutput } = await analyzeInfluencer(
      {
        name: selected.influencer_name,
        platform: selected.platform,
        niche: selected.niche,
        audience_geo: selected.audience_geo,
        audience_demo: selected.audience_demo,
        follower_range: selected.follower_range,
        engagement_rate: selected.engagement_rate,
        content_formats: selected.content_formats,
        notes: selected.notes,
        question: query,
      },
      (report) => {
        setEngineBusy(isWebLLMLoading());
        if (!isWebLLMReady() || report.progress < 1) {
          setModelProgress(Math.round(report.progress * 100));
          setModelProgressText(report.text);
        }
      },
    );

    setEngineBusy(false);
    setModelProgress(null);
    await persistAnalysis(
      result,
      rawOutput,
      "web-llm",
      WEBLLM_MODEL_ID,
      startTime,
      brandProfileId,
    );
  }

  async function handleAnalyze() {
    if (!selected || analyzing) return;

    const trimmedQuestion = userQuestion.trim();
    if (trimmedQuestion.length > MAX_ANALYZE_QUERY_LEN) {
      setAnalysisError(t("errors.questionTooLong", { max: MAX_ANALYZE_QUERY_LEN }));
      return;
    }

    const query = trimmedQuestion || MCP_ANALYZE_QUERY;
    const brandProfileId = selectedBrandProfileId;
    const brandName = brandProfileId
      ? brandProfiles.find((p) => p.id === brandProfileId)?.name ?? null
      : null;

    setAnalyzing(true);
    setAnalysisPhase("server");
    setSelectedInfluencerId(selected.id);
    setAnalysisError(null);
    setAnalysisNotice(null);
    setLiveResult(null);
    setAnalysisSource(null);
    setAskedQuestion(trimmedQuestion || null);
    setAnalyzedForBrandName(brandName);
    setModelProgress(null);

    const startTime = performance.now();

    try {
      const rich = await sendMCPRequest(
        {
          request_type: "analyze_influencer",
          context: buildMCPContext({
            influencer_name: selected.influencer_name,
            platform: selected.platform,
            niche: selected.niche,
            audience_geo: selected.audience_geo,
            audience_demo: selected.audience_demo,
            follower_range: selected.follower_range,
            engagement_rate: selected.engagement_rate,
            content_formats: selected.content_formats,
            notes: selected.notes,
          }),
          query,
          ...(brandProfileId ? { brand_profile_id: brandProfileId } : {}),
        },
        undefined,
        SERVER_LLM_ANALYZE_TIMEOUT_MS,
      );

      const { result, rawOutput, model } = mapMCPRichResultToAnalysis(rich);
      await persistAnalysis(result, rawOutput, "ollama", model, startTime, brandProfileId);
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

      setAnalysisNotice(t("notice.webllmFallback"));

      try {
        await runWebLLMAnalyze(startTime, query, brandProfileId);
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
    if (analyzing) return;

    setSelectedId(id);
    setLiveResult(null);
    setAnalysisError(null);
    setAnalysisNotice(null);
    setAnalysisSource(null);
    setSelectedInfluencerId(null);
    setAskedQuestion(null);
    setUserQuestion("");
    setSelectedBrandProfileId(null);
    setAnalyzedForBrandName(null);
    setExpandedHistoryId(null);
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

  const analyzeButtonLabel = useMemo(() => {
    if (!analyzing) return t("analyze.button");
    if (analysisPhase === "server") return t("analyze.server");
    if (engineBusy && !isWebLLMReady()) return t("analyze.browserLoading");
    return t("analyze.browser");
  }, [analyzing, analysisPhase, engineBusy, t]);

  if (error) {
    return (
      <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-6 py-8 text-center text-red-400">
        {error}
      </div>
    );
  }

  return (
    <div className="relative space-y-6">
      {loading && (
        <div className="absolute inset-0 z-10 flex items-center justify-center rounded-xl bg-[var(--background)]/70 backdrop-blur-sm">
          <p className="text-[var(--muted)]">{t("loading")}</p>
        </div>
      )}
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t("title")}</h1>
        <p className="mt-1 text-[var(--muted)]">
          {t("subtitle", {
            serverModel: SERVER_LLM_MODEL_ID,
            browserModel: WEBLLM_MODEL_ID,
          })}
        </p>
      </div>

      {scores.length === 0 ? (
        <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] px-6 py-16 text-center">
          <p className="text-lg font-medium">{t("emptyScores.title")}</p>
          <p className="mt-2 text-sm text-[var(--muted)]">{t("emptyScores.description")}</p>
        </div>
      ) : (
        <div className="grid gap-6 lg:grid-cols-3">
          <aside className="space-y-2 lg:col-span-1">
            <p className="mb-3 text-xs font-medium uppercase tracking-wider text-[var(--muted)]">
              {t("selectInfluencer")}
            </p>
            {scores.map((s) => (
              <button
                key={s.id}
                onClick={() => handleSelect(s.id)}
                disabled={analyzing}
                className={`w-full rounded-xl border px-4 py-3 text-left transition-all disabled:cursor-not-allowed disabled:opacity-60 ${
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
                    <p className="mt-2 text-xs text-[var(--muted)]">{profileSummary(selected)}</p>
                    {selected.notes && (
                      <p className="mt-1 text-xs text-[var(--muted)]">
                        {t("collaborations", { notes: selected.notes })}
                      </p>
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

                <div className="mt-5">
                  <label
                    htmlFor="brand-profile"
                    className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
                  >
                    {t("brandProfile.label")}{" "}
                    <span className="font-normal text-[var(--muted)]">
                      {t("brandProfile.optional")}
                    </span>
                  </label>
                  <select
                    id="brand-profile"
                    value={selectedBrandProfileId ?? ""}
                    onChange={(e) =>
                      setSelectedBrandProfileId(e.target.value || null)
                    }
                    disabled={analyzing || engineBusy}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 disabled:opacity-50"
                  >
                    <option value="">{t("brandProfile.none")}</option>
                    {brandProfiles.map((profile) => (
                      <option key={profile.id} value={profile.id}>
                        {profile.name}
                      </option>
                    ))}
                  </select>
                  {brandProfiles.length === 0 && (
                    <p className="mt-1 text-xs text-[var(--muted)]">
                      {t("brandProfile.emptyHint")}
                    </p>
                  )}
                </div>

                <div className="mt-5">
                  <label
                    htmlFor="analyze-question"
                    className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
                  >
                    {t("question.label")}{" "}
                    <span className="font-normal text-[var(--muted)]">
                      {t("question.optional")}
                    </span>
                  </label>
                  <textarea
                    id="analyze-question"
                    rows={2}
                    value={userQuestion}
                    onChange={(e) => setUserQuestion(e.target.value)}
                    maxLength={MAX_ANALYZE_QUERY_LEN}
                    disabled={analyzing || engineBusy}
                    placeholder={t("question.placeholder")}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-3 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 disabled:opacity-50"
                  />
                  <p className="mt-1 text-xs text-[var(--muted)]">
                    {t("question.charCount", {
                      current: userQuestion.length,
                      max: MAX_ANALYZE_QUERY_LEN,
                    })}
                  </p>
                </div>

                {modelProgress !== null && (
                  <div className="mt-5">
                    <div className="mb-1.5 flex items-center justify-between text-xs text-[var(--muted)]">
                      <span>
                        {modelProgressText || t("analyze.browserLoading")}
                      </span>
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
                    label={liveResult ? t("scores.overallAi") : t("scores.overall")}
                    description={t("scores.overallDesc")}
                    value={displayScores.overall}
                    highlight
                  />
                  <ScorePill
                    label={t("scores.engagement")}
                    description={t("scores.engagementDesc")}
                    value={displayScores.engagement}
                  />
                  <ScorePill
                    label={t("scores.audience")}
                    description={t("scores.audienceDesc")}
                    value={displayScores.audience}
                  />
                  <ScorePill
                    label={t("scores.brandFit")}
                    description={t("scores.brandFitDesc")}
                    value={displayScores.brandFit}
                  />
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
                  {liveResult && analyzedForBrandName && (
                    <div className="inline-flex rounded-full border border-[var(--accent)]/30 bg-[var(--accent-muted)] px-4 py-1.5 text-sm font-medium text-[var(--accent)]">
                      {t("analyzedFor", { brandName: analyzedForBrandName })}
                    </div>
                  )}

                  {askedQuestion && (
                    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface-elevated)] px-5 py-4">
                      <p className="text-xs font-semibold uppercase tracking-wider text-[var(--muted)]">
                        {t("questionHeading")}
                      </p>
                      <p className="mt-1 text-sm leading-relaxed">{askedQuestion}</p>
                    </div>
                  )}

                  <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                    <h3 className="mb-2 text-sm font-semibold uppercase tracking-wider text-[var(--muted)]">
                      {t("summary.title")}
                      {liveResult && (
                        <span className="ml-2 normal-case text-[var(--accent)]">
                          {analysisSource === "web-llm"
                            ? t("summary.newBrowser")
                            : t("summary.newServer")}
                        </span>
                      )}
                    </h3>
                    <p className="leading-relaxed">{displaySummary}</p>
                  </div>

                  {displayInsights.length > 0 && (
                    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                      <h3 className="mb-3 text-sm font-semibold uppercase tracking-wider text-[var(--muted)]">
                        {t("insights.title")}
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

                  {pastAnalysisHistory.length > 0 && (
                    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
                      <h3 className="mb-1 text-sm font-semibold uppercase tracking-wider text-[var(--muted)]">
                        {t("history.title")}
                      </h3>
                      <p className="mb-4 text-xs text-[var(--muted)]">
                        {pastAnalysisHistory.length === 1
                          ? t("history.earlierOne", { count: pastAnalysisHistory.length })
                          : t("history.earlierOther", { count: pastAnalysisHistory.length })}
                      </p>
                      <div className="space-y-2">
                        {pastAnalysisHistory.map((entry) => {
                          const expanded = expandedHistoryId === entry.id;
                          const insights = parseStoredInsights(entry.insights);
                          return (
                            <div
                              key={entry.id}
                              className="overflow-hidden rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)]"
                            >
                              <button
                                type="button"
                                onClick={() =>
                                  setExpandedHistoryId((prev) =>
                                    prev === entry.id ? null : entry.id,
                                  )
                                }
                                className="flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-[var(--surface)]"
                              >
                                <span
                                  className={`mt-0.5 shrink-0 text-xs text-[var(--muted)] transition-transform ${
                                    expanded ? "rotate-90" : ""
                                  }`}
                                  aria-hidden
                                >
                                  ▶
                                </span>
                                <div className="min-w-0 flex-1">
                                  <p className="text-xs font-medium text-[var(--muted)]">
                                    {formatAnalysisDate(entry.created_at, locale)}
                                    <span className="ml-2 capitalize text-[var(--foreground)]/70">
                                      · {entry.analysis_type.replace(/-/g, " ")}
                                    </span>
                                  </p>
                                  <p className="mt-1 text-sm text-[var(--foreground)]/90">
                                    {truncateSummary(entry.summary)}
                                  </p>
                                </div>
                              </button>
                              {expanded && (
                                <div className="border-t border-[var(--border)] px-4 py-4">
                                  <p className="text-sm leading-relaxed">{entry.summary}</p>
                                  {insights.length > 0 && (
                                    <ul className="mt-4 space-y-2">
                                      {insights.map((insight, i) => (
                                        <li key={i} className="flex gap-2 text-sm">
                                          <span className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-[var(--accent)]" />
                                          {insight}
                                        </li>
                                      ))}
                                    </ul>
                                  )}
                                </div>
                              )}
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  )}
                </>
              ) : (
                !analysisError &&
                !analyzing && (
                  <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] p-8 text-center">
                    <p className="font-medium text-[var(--muted)]">{t("noAnalysis.title")}</p>
                    <p className="mt-2 text-sm text-[var(--muted)]">
                      {t("noAnalysis.description")}
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
  description,
  value,
  highlight,
}: {
  label: string;
  description?: string;
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
      {description && (
        <p className="mt-0.5 text-xs text-[var(--muted)] opacity-80">{description}</p>
      )}
      <p
        className={`mt-1 text-lg font-bold ${highlight ? "text-[var(--accent)]" : scoreColor(value)}`}
      >
        {value}
      </p>
    </div>
  );
}
