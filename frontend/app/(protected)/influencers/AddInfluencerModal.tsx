"use client";

import { FormEvent, useEffect, useState } from "react";
import {
  ApiError,
  scoresApi,
  type CreateScorePayload,
} from "@/lib/api";

const PLATFORMS = [
  { value: "instagram", label: "Instagram" },
  { value: "tiktok", label: "TikTok" },
  { value: "youtube", label: "YouTube" },
  { value: "twitter", label: "Twitter" },
  { value: "linkedin", label: "LinkedIn" },
  { value: "other", label: "Other" },
] as const;

const inputClass =
  "w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30";

function parseScore(value: string): number {
  if (value.trim() === "") return 0;
  const n = Number(value);
  return Number.isFinite(n) ? n : NaN;
}

type Props = {
  open: boolean;
  onClose: () => void;
  onSuccess: () => void;
};

export default function AddInfluencerModal({ open, onClose, onSuccess }: Props) {
  const [influencerName, setInfluencerName] = useState("");
  const [platform, setPlatform] = useState<string>("instagram");
  const [notes, setNotes] = useState("");
  const [overallScore, setOverallScore] = useState("");
  const [engagementScore, setEngagementScore] = useState("");
  const [audienceScore, setAudienceScore] = useState("");
  const [brandFitScore, setBrandFitScore] = useState("");
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
    setNotes("");
    setOverallScore("");
    setEngagementScore("");
    setAudienceScore("");
    setBrandFitScore("");
    setError(null);
  }

  function handleClose() {
    if (loading) return;
    resetForm();
    onClose();
  }

  function validate(): string | null {
    if (!influencerName.trim()) return "Influencer name is required";
    const scores = [
      { label: "Overall score", value: parseScore(overallScore) },
      { label: "Engagement score", value: parseScore(engagementScore) },
      { label: "Audience score", value: parseScore(audienceScore) },
      { label: "Brand fit score", value: parseScore(brandFitScore) },
    ];
    for (const { label, value } of scores) {
      if (Number.isNaN(value)) return `${label} must be a valid number`;
      if (value < 0 || value > 100) return `${label} must be between 0 and 100`;
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
      overall_score: parseScore(overallScore),
      engagement_score: parseScore(engagementScore),
      audience_score: parseScore(audienceScore),
      brand_fit_score: parseScore(brandFitScore),
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
        setError("Failed to add influencer. Please try again.");
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
        aria-label="Close"
      />

      <div className="relative z-10 w-full max-w-lg rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-6 shadow-2xl shadow-black/40">
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <h2 id="add-influencer-title" className="text-lg font-bold tracking-tight">
              Add New Influencer
            </h2>
            <p className="mt-1 text-sm text-[var(--muted)]">
              Manually add an influencer record to the pool
            </p>
          </div>
          <button
            type="button"
            onClick={handleClose}
            disabled={loading}
            className="rounded-lg p-1.5 text-[var(--muted)] transition-colors hover:bg-[var(--surface-elevated)] hover:text-[var(--foreground)] disabled:opacity-50"
            aria-label="Close"
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
              Influencer Name <span className="text-[var(--accent)]">*</span>
            </label>
            <input
              id="influencer_name"
              type="text"
              value={influencerName}
              onChange={(e) => setInfluencerName(e.target.value)}
              className={inputClass}
              placeholder="e.g. Jane Smith"
              autoFocus
            />
          </div>

          <div>
            <label htmlFor="platform" className="mb-1.5 block text-sm font-medium">
              Platform <span className="text-[var(--accent)]">*</span>
            </label>
            <select
              id="platform"
              value={platform}
              onChange={(e) => setPlatform(e.target.value)}
              className={inputClass}
            >
              {PLATFORMS.map((p) => (
                <option key={p.value} value={p.value}>
                  {p.label}
                </option>
              ))}
            </select>
          </div>

          <div>
            <label htmlFor="notes" className="mb-1.5 block text-sm font-medium">
              Notes
            </label>
            <textarea
              id="notes"
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              className={`${inputClass} resize-none`}
              placeholder="Beauty and lifestyle niche"
            />
          </div>

          <div>
            <p className="mb-2 text-sm font-medium">Scores (0–100, optional)</p>
            <div className="grid grid-cols-2 gap-3">
              <ScoreField
                id="overall_score"
                label="Overall"
                value={overallScore}
                onChange={setOverallScore}
              />
              <ScoreField
                id="engagement_score"
                label="Engagement"
                value={engagementScore}
                onChange={setEngagementScore}
              />
              <ScoreField
                id="audience_score"
                label="Audience"
                value={audienceScore}
                onChange={setAudienceScore}
              />
              <ScoreField
                id="brand_fit_score"
                label="Brand Fit"
                value={brandFitScore}
                onChange={setBrandFitScore}
              />
            </div>
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
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="flex-1 rounded-lg bg-[var(--accent)] py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {loading ? "Saving..." : "Add"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ScoreField({
  id,
  label,
  value,
  onChange,
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
}) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-xs text-[var(--muted)]">
        {label}
      </label>
      <input
        id={id}
        type="number"
        min={0}
        max={100}
        step={1}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={inputClass}
        placeholder="0"
      />
    </div>
  );
}
