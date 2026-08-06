"use client";

import { FormEvent, useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  ApiError,
  scoresApi,
  type CreateScorePayload,
} from "@/lib/api";
import {
  AUDIENCE_DEMO_KEYS,
  AUDIENCE_DEMO_VALUES,
  AUDIENCE_GEO_KEYS,
  AUDIENCE_GEO_VALUES,
  CONTENT_FORMAT_KEYS,
  CONTENT_FORMAT_VALUES,
  FOLLOWER_RANGE_KEYS,
  FOLLOWER_RANGE_VALUES,
  NICHE_KEYS,
  NICHE_VALUES,
  PLATFORM_KEYS,
} from "@/lib/influencer-profile";

const inputClass =
  "w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30";

type Props = {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
};

export default function AddInfluencerModal({ open, onClose, onSuccess }: Props) {
  const t = useTranslations("influencersModal");
  const tOptions = useTranslations("influencersModal.options");

  const [influencerName, setInfluencerName] = useState("");
  const [platform, setPlatform] = useState<string>("instagram");
  const [niche, setNiche] = useState<string>(NICHE_VALUES[NICHE_KEYS[0]]);
  const [audienceGeo, setAudienceGeo] = useState<string>(
    AUDIENCE_GEO_VALUES[AUDIENCE_GEO_KEYS[0]],
  );
  const [audienceDemo, setAudienceDemo] = useState<string>(
    AUDIENCE_DEMO_VALUES[AUDIENCE_DEMO_KEYS[1]],
  );
  const [followerRange, setFollowerRange] = useState<string>(
    FOLLOWER_RANGE_VALUES[FOLLOWER_RANGE_KEYS[2]],
  );
  const [engagementRate, setEngagementRate] = useState("");
  const [contentFormats, setContentFormats] = useState<string[]>([
    CONTENT_FORMAT_VALUES.reels,
  ]);
  const [notes, setNotes] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open) return;
    function onKeyDown(e: KeyboardEvent) {
      if (e.key === "Escape" && !loading) onClose();
    }
    document.addEventListener("keydown", onKeyDown);
    return () => document.removeEventListener("keydown", onKeyDown);
  }, [open, loading, onClose]);

  function resetForm() {
    setInfluencerName("");
    setPlatform("instagram");
    setNiche(NICHE_VALUES[NICHE_KEYS[0]]);
    setAudienceGeo(AUDIENCE_GEO_VALUES[AUDIENCE_GEO_KEYS[0]]);
    setAudienceDemo(AUDIENCE_DEMO_VALUES[AUDIENCE_DEMO_KEYS[1]]);
    setFollowerRange(FOLLOWER_RANGE_VALUES[FOLLOWER_RANGE_KEYS[2]]);
    setEngagementRate("");
    setContentFormats([CONTENT_FORMAT_VALUES.reels]);
    setNotes("");
    setError(null);
  }

  function handleClose() {
    if (loading) return;
    resetForm();
    onClose();
  }

  function toggleFormat(formatValue: string) {
    setContentFormats((prev) =>
      prev.includes(formatValue)
        ? prev.filter((item) => item !== formatValue)
        : [...prev, formatValue],
    );
  }

  function validate(): string | null {
    if (!influencerName.trim()) return t("validation.nameRequired");
    if (!niche) return t("validation.nicheRequired");
    if (!audienceGeo) return t("validation.audienceRegionRequired");
    if (!audienceDemo) return t("validation.audienceDemoRequired");
    if (!followerRange) return t("validation.followerRangeRequired");
    const rate = Number(engagementRate);
    if (engagementRate.trim() === "" || Number.isNaN(rate)) {
      return t("validation.engagementRateRequired");
    }
    if (rate < 0 || rate > 100) {
      return t("validation.engagementRateRange");
    }
    if (contentFormats.length === 0) {
      return t("validation.contentFormatRequired");
    }
    return null;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError(null);

    const validationError = validate();
    if (validationError) {
      setError(validationError);
      return;
    }

    const payload: CreateScorePayload = {
      influencer_name: influencerName.trim(),
      platform,
      niche,
      audience_geo: audienceGeo,
      audience_demo: audienceDemo,
      follower_range: followerRange,
      engagement_rate: Number(engagementRate),
      content_formats: contentFormats,
    };
    if (notes.trim()) payload.notes = notes.trim();

    setLoading(true);
    try {
      await scoresApi.create(payload);
      resetForm();
      onSuccess();
      onClose();
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError(t("errors.addFailed"));
      }
    } finally {
      setLoading(false);
    }
  }

  if (!open) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="add-influencer-title"
    >
      <button
        type="button"
        className="absolute inset-0 bg-black/60 backdrop-blur-sm"
        onClick={handleClose}
        aria-label={t("close")}
      />

      <div className="relative z-10 max-h-[90vh] w-full max-w-xl overflow-y-auto rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-6 shadow-2xl shadow-black/40">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <h2 id="add-influencer-title" className="text-lg font-bold tracking-tight">
              {t("title")}
            </h2>
            <p className="mt-1 text-sm text-[var(--muted)]">{t("subtitle")}</p>
          </div>
          <button
            type="button"
            onClick={handleClose}
            disabled={loading}
            className="rounded-lg p-1.5 text-[var(--muted)] transition-colors hover:bg-[var(--surface-elevated)] hover:text-[var(--foreground)] disabled:opacity-50"
            aria-label={t("close")}
          >
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M18 6 6 18" />
              <path d="m6 6 12 12" />
            </svg>
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label htmlFor="influencer_name" className="mb-1.5 block text-sm font-medium">
              {t("fields.influencerName")}{" "}
              <span className="text-[var(--accent)]">*</span>
            </label>
            <input
              id="influencer_name"
              type="text"
              value={influencerName}
              onChange={(e) => setInfluencerName(e.target.value)}
              className={inputClass}
              placeholder={t("placeholders.influencerName")}
              autoFocus
            />
          </div>

          <div>
            <label htmlFor="platform" className="mb-1.5 block text-sm font-medium">
              {t("fields.platform")}{" "}
              <span className="text-[var(--accent)]">*</span>
            </label>
            <select
              id="platform"
              value={platform}
              onChange={(e) => setPlatform(e.target.value)}
              className={inputClass}
            >
              {PLATFORM_KEYS.map((key) => (
                <option key={key} value={key}>
                  {tOptions(`platforms.${key}`)}
                </option>
              ))}
            </select>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <label htmlFor="niche" className="mb-1.5 block text-sm font-medium">
                {t("fields.niche")}{" "}
                <span className="text-[var(--accent)]">*</span>
              </label>
              <select
                id="niche"
                value={niche}
                onChange={(e) => setNiche(e.target.value)}
                className={inputClass}
              >
                {NICHE_KEYS.map((key) => (
                  <option key={key} value={NICHE_VALUES[key]}>
                    {tOptions(`niches.${key}`)}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="audience_geo" className="mb-1.5 block text-sm font-medium">
                {t("fields.audienceRegion")}{" "}
                <span className="text-[var(--accent)]">*</span>
              </label>
              <select
                id="audience_geo"
                value={audienceGeo}
                onChange={(e) => setAudienceGeo(e.target.value)}
                className={inputClass}
              >
                {AUDIENCE_GEO_KEYS.map((key) => (
                  <option key={key} value={AUDIENCE_GEO_VALUES[key]}>
                    {tOptions(`audienceGeo.${key}`)}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="audience_demo" className="mb-1.5 block text-sm font-medium">
                {t("fields.audienceDemographics")}{" "}
                <span className="text-[var(--accent)]">*</span>
              </label>
              <select
                id="audience_demo"
                value={audienceDemo}
                onChange={(e) => setAudienceDemo(e.target.value)}
                className={inputClass}
              >
                {AUDIENCE_DEMO_KEYS.map((key) => (
                  <option key={key} value={AUDIENCE_DEMO_VALUES[key]}>
                    {tOptions(`audienceDemo.${key}`)}
                  </option>
                ))}
              </select>
            </div>

            <div>
              <label htmlFor="follower_range" className="mb-1.5 block text-sm font-medium">
                {t("fields.followers")}{" "}
                <span className="text-[var(--accent)]">*</span>
              </label>
              <select
                id="follower_range"
                value={followerRange}
                onChange={(e) => setFollowerRange(e.target.value)}
                className={inputClass}
              >
                {FOLLOWER_RANGE_KEYS.map((key) => (
                  <option key={key} value={FOLLOWER_RANGE_VALUES[key]}>
                    {tOptions(`followerRanges.${key}`)}
                  </option>
                ))}
              </select>
            </div>
          </div>

          <div>
            <label htmlFor="engagement_rate" className="mb-1.5 block text-sm font-medium">
              {t("fields.engagementRate")}{" "}
              <span className="text-[var(--accent)]">*</span>
            </label>
            <input
              id="engagement_rate"
              type="number"
              min="0"
              max="100"
              step="0.1"
              value={engagementRate}
              onChange={(e) => setEngagementRate(e.target.value)}
              className={inputClass}
              placeholder={t("placeholders.engagementRate")}
            />
          </div>

          <div>
            <p className="mb-2 text-sm font-medium">
              {t("fields.contentFormat")}{" "}
              <span className="text-[var(--accent)]">*</span>
            </p>
            <div className="flex flex-wrap gap-2">
              {CONTENT_FORMAT_KEYS.map((key) => {
                const formatValue = CONTENT_FORMAT_VALUES[key];
                const active = contentFormats.includes(formatValue);
                return (
                  <button
                    key={key}
                    type="button"
                    onClick={() => toggleFormat(formatValue)}
                    className={`rounded-full border px-3 py-1.5 text-sm transition-colors ${
                      active
                        ? "border-[var(--accent)] bg-[var(--accent-muted)] text-[var(--accent)]"
                        : "border-[var(--border)] text-[var(--muted)] hover:border-[var(--accent)]/40"
                    }`}
                  >
                    {tOptions(`contentFormats.${key}`)}
                  </button>
                );
              })}
            </div>
          </div>

          <div>
            <label htmlFor="notes" className="mb-1.5 block text-sm font-medium">
              {t("fields.pastCollaborationsOptional")}
            </label>
            <textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              className={`${inputClass} resize-none`}
              placeholder={t("placeholders.pastCollaborations")}
            />
          </div>

          {error && (
            <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-400">
              {error}
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={handleClose}
              disabled={loading}
              className="flex-1 rounded-lg border border-[var(--border)] py-2.5 text-sm font-medium text-[var(--muted)] transition-colors hover:border-[var(--accent)]/40 hover:text-[var(--foreground)] disabled:opacity-50"
            >
              {t("submit.cancel")}
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 rounded-lg bg-[var(--accent)] py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {loading ? t("submit.saving") : t("submit.add")}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
