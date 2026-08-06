"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  handleUnauthorizedRedirect,
  isUnauthorized,
  scoresApi,
  type InfluencerScore,
} from "@/lib/api";
import { profileSummary } from "@/lib/influencer-profile";
import { platformColors, scoreColor } from "@/lib/score-utils";
import AddInfluencerModal from "./AddInfluencerModal";

export default function InfluencersPage() {
  const t = useTranslations("influencers");
  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  async function loadScores() {
    const data = await scoresApi.list();
    setScores(data.scores ?? []);
  }

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await scoresApi.list();
        if (cancelled) return;
        setScores(data.scores ?? []);
      } catch (err) {
        if (cancelled) return;
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/influencers");
          return;
        }
        setError(t("errors.loadFailed"));
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    load();
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (!successMessage) return;
    const timer = window.setTimeout(() => setSuccessMessage(null), 4000);
    return () => window.clearTimeout(timer);
  }, [successMessage]);

  async function handleAddSuccess() {
    try {
      await loadScores();
      setSuccessMessage(t("success.added"));
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/influencers");
        return;
      }
      setError(t("errors.refreshFailed"));
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 text-[var(--muted)]">
        {t("loading")}
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
    <>
      <div className="space-y-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="text-2xl font-bold tracking-tight">{t("title")}</h1>
            <p className="mt-1 text-[var(--muted)]">
              {t("subtitle")}
              {scores.length > 0 && (
                <span className="ml-2 text-[var(--accent)]">({scores.length})</span>
              )}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg bg-[var(--accent)] px-4 py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90"
          >
            <span aria-hidden="true">+</span>
            {t("addInfluencer")}
          </button>
        </div>

        {successMessage && (
          <div
            role="status"
            className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-400"
          >
            {successMessage}
          </div>
        )}

        {scores.length === 0 ? (
          <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] px-6 py-16 text-center">
            <p className="text-lg font-medium">{t("empty.title")}</p>
            <p className="mt-2 text-sm text-[var(--muted)]">{t("empty.description")}</p>
            <button
              type="button"
              onClick={() => setModalOpen(true)}
              className="mt-6 inline-flex items-center gap-2 rounded-lg border border-[var(--accent)]/40 bg-[var(--accent-muted)] px-4 py-2 text-sm font-medium text-[var(--accent)] transition-colors hover:border-[var(--accent)]/60"
            >
              {t("addFirstInfluencer")}
            </button>
          </div>
        ) : (
          <>
            <div className="hidden overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] md:block">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)] text-[var(--muted)]">
                    <th className="px-5 py-3 font-medium">{t("table.influencer")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.platform")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.overall")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.engagement")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.audience")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.brandFit")}</th>
                    <th className="px-5 py-3 font-medium">{t("table.profile")}</th>
                  </tr>
                </thead>
                <tbody>
                  {scores.map((s) => (
                    <tr
                      key={s.id}
                      className="border-b border-[var(--border)] last:border-0 transition-colors hover:bg-[var(--surface-elevated)]"
                    >
                      <td className="px-5 py-4 font-medium">{s.influencer_name}</td>
                      <td className="px-5 py-4">
                        <span
                          className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${platformColors[s.platform] ?? "bg-[var(--surface-elevated)] text-[var(--muted)]"}`}
                        >
                          {s.platform}
                        </span>
                      </td>
                      <td className={`px-5 py-4 font-semibold ${scoreColor(s.overall_score)}`}>
                        {s.overall_score}
                      </td>
                      <td className="px-5 py-4 text-[var(--muted)]">{s.engagement_score}</td>
                      <td className="px-5 py-4 text-[var(--muted)]">{s.audience_score}</td>
                      <td className="px-5 py-4 text-[var(--muted)]">{s.brand_fit_score}</td>
                      <td className="max-w-xs truncate px-5 py-4 text-[var(--muted)]">
                        {profileSummary(s)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="grid gap-4 md:hidden">
              {scores.map((s) => (
                <article
                  key={s.id}
                  className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5"
                >
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="font-semibold">{s.influencer_name}</h3>
                      <span
                        className={`mt-1 inline-block rounded-full px-2.5 py-0.5 text-xs font-medium capitalize ${platformColors[s.platform] ?? ""}`}
                      >
                        {s.platform}
                      </span>
                    </div>
                    <span className={`text-2xl font-bold ${scoreColor(s.overall_score)}`}>
                      {s.overall_score}
                    </span>
                  </div>
                  <div className="mt-4 grid grid-cols-3 gap-2 text-center">
                    <MiniStat label={t("mobile.engagement")} value={s.engagement_score} />
                    <MiniStat label={t("mobile.audience")} value={s.audience_score} />
                    <MiniStat label={t("mobile.brand")} value={s.brand_fit_score} />
                  </div>
                  <p className="mt-3 text-xs text-[var(--muted)]">{profileSummary(s)}</p>
                  {s.notes && (
                    <p className="mt-1 text-xs text-[var(--muted)]">
                      {t("mobile.collaborations", { notes: s.notes })}
                    </p>
                  )}
                </article>
              ))}
            </div>
          </>
        )}
      </div>

      <AddInfluencerModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        onSuccess={handleAddSuccess}
      />
    </>
  );
}

function MiniStat({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-lg bg-[var(--surface-elevated)] px-2 py-2">
      <p className="text-[10px] uppercase tracking-wider text-[var(--muted)]">{label}</p>
      <p className="text-sm font-semibold">{value}</p>
    </div>
  );
}
