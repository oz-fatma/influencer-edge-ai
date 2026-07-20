"use client";

import { useCallback, useEffect, useState } from "react";
import {
  handleUnauthorizedRedirect,
  isUnauthorized,
  scoresApi,
  type InfluencerScore,
} from "@/lib/api";
import { platformColors, scoreColor } from "@/lib/score-utils";
import AddInfluencerModal from "./AddInfluencerModal";

export default function InfluencersPage() {
  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  const loadScores = useCallback(async () => {
    const data = await scoresApi.list();
    setScores(data.scores ?? []);
  }, []);

  useEffect(() => {
    async function load() {
      try {
        await loadScores();
      } catch (err) {
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/influencers");
          return;
        }
        setError("Skorlar yüklenemedi. Lütfen tekrar deneyin.");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, [loadScores]);

  useEffect(() => {
    if (!successMessage) return;
    const timer = window.setTimeout(() => setSuccessMessage(null), 4000);
    return () => window.clearTimeout(timer);
  }, [successMessage]);

  async function handleAddSuccess() {
    try {
      await loadScores();
      setSuccessMessage("Influencer başarıyla eklendi.");
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/influencers");
        return;
      }
      setError("Liste yenilenemedi. Sayfayı yenileyin.");
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 text-[var(--muted)]">
        Yükleniyor...
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
            <h1 className="text-2xl font-bold tracking-tight">Influencer Havuzu</h1>
            <p className="mt-1 text-[var(--muted)]">
              Web-LLM skor sonuçları
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
            Influencer Ekle
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
            <p className="text-lg font-medium">Henüz skor eklenmedi</p>
            <p className="mt-2 text-sm text-[var(--muted)]">
              Web-LLM analiz sonuçları kaydedildiğinde veya manuel ekleme yaptığınızda burada görünecek.
            </p>
            <button
              type="button"
              onClick={() => setModalOpen(true)}
              className="mt-6 inline-flex items-center gap-2 rounded-lg border border-[var(--accent)]/40 bg-[var(--accent-muted)] px-4 py-2 text-sm font-medium text-[var(--accent)] transition-colors hover:border-[var(--accent)]/60"
            >
              + İlk influencer&apos;ı ekle
            </button>
          </div>
        ) : (
          <>
            <div className="hidden overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)] md:block">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)] text-[var(--muted)]">
                    <th className="px-5 py-3 font-medium">Influencer</th>
                    <th className="px-5 py-3 font-medium">Platform</th>
                    <th className="px-5 py-3 font-medium">Genel</th>
                    <th className="px-5 py-3 font-medium">Etkileşim</th>
                    <th className="px-5 py-3 font-medium">Kitle</th>
                    <th className="px-5 py-3 font-medium">Marka Uyumu</th>
                    <th className="px-5 py-3 font-medium">Not</th>
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
                        {s.notes || "—"}
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
                    <MiniStat label="Etkileşim" value={s.engagement_score} />
                    <MiniStat label="Kitle" value={s.audience_score} />
                    <MiniStat label="Marka" value={s.brand_fit_score} />
                  </div>
                  {s.notes && (
                    <p className="mt-3 text-xs text-[var(--muted)]">{s.notes}</p>
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
