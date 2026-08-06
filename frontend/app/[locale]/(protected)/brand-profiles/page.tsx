"use client";

import { FormEvent, useEffect, useState, type ReactNode } from "react";
import { useTranslations } from "next-intl";
import {
  ApiError,
  brandProfilesApi,
  handleUnauthorizedRedirect,
  isUnauthorized,
  type BrandProfile,
  type CreateBrandProfilePayload,
} from "@/lib/api";

type FormState = CreateBrandProfilePayload;

const emptyForm: FormState = {
  name: "",
  industry: "",
  target_audience: "",
  budget_range: "",
  brand_values: "",
  campaign_goal: "",
};

export default function BrandProfilesPage() {
  const t = useTranslations("brandProfiles");
  const [profiles, setProfiles] = useState<BrandProfile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<FormState>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

  async function loadProfiles() {
    const data = await brandProfilesApi.list();
    setProfiles(data.brand_profiles ?? []);
  }

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await brandProfilesApi.list();
        if (cancelled) return;
        setProfiles(data.brand_profiles ?? []);
      } catch (err) {
        if (cancelled) return;
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/brand-profiles");
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

  function openCreateForm() {
    setEditingId(null);
    setForm(emptyForm);
    setShowForm(true);
    setError(null);
  }

  function openEditForm(profile: BrandProfile) {
    setEditingId(profile.id);
    setForm({
      name: profile.name,
      industry: profile.industry,
      target_audience: profile.target_audience,
      budget_range: profile.budget_range ?? "",
      brand_values: profile.brand_values,
      campaign_goal: profile.campaign_goal,
    });
    setShowForm(true);
    setError(null);
  }

  function closeForm() {
    setShowForm(false);
    setEditingId(null);
    setForm(emptyForm);
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);

    const payload: CreateBrandProfilePayload = {
      name: form.name.trim(),
      industry: form.industry.trim(),
      target_audience: form.target_audience.trim(),
      brand_values: form.brand_values.trim(),
      campaign_goal: form.campaign_goal.trim(),
    };
    const budget = form.budget_range?.trim();
    if (budget) {
      payload.budget_range = budget;
    }

    try {
      if (editingId) {
        await brandProfilesApi.update(editingId, payload);
        setSuccessMessage(t("success.updated"));
      } else {
        await brandProfilesApi.create(payload);
        setSuccessMessage(t("success.created"));
      }
      await loadProfiles();
      closeForm();
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/brand-profiles");
        return;
      }
      setError(err instanceof ApiError ? err.message : t("errors.saveFailed"));
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(profile: BrandProfile) {
    if (!window.confirm(t("confirmDelete", { name: profile.name }))) {
      return;
    }

    setDeletingId(profile.id);
    setError(null);
    try {
      await brandProfilesApi.delete(profile.id);
      setProfiles((prev) => prev.filter((p) => p.id !== profile.id));
      if (editingId === profile.id) {
        closeForm();
      }
      setSuccessMessage(t("success.deleted"));
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/brand-profiles");
        return;
      }
      setError(err instanceof ApiError ? err.message : t("errors.deleteFailed"));
    } finally {
      setDeletingId(null);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 text-[var(--muted)]">
        {t("loading")}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("title")}</h1>
          <p className="mt-1 text-[var(--muted)]">
            {t("subtitle")}
            {profiles.length > 0 && (
              <span className="ml-2 text-[var(--accent)]">({profiles.length})</span>
            )}
          </p>
        </div>
        {!showForm && (
          <button
            type="button"
            onClick={openCreateForm}
            className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg bg-[var(--accent)] px-4 py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90"
          >
            <span aria-hidden="true">+</span>
            {t("newProfile")}
          </button>
        )}
      </div>

      {successMessage && (
        <div
          role="status"
          className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-400"
        >
          {successMessage}
        </div>
      )}

      {error && (
        <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          {error}
        </div>
      )}

      {showForm && (
        <form
          onSubmit={handleSubmit}
          className="space-y-5 rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-6"
        >
          <div className="flex items-center justify-between gap-4">
            <h2 className="text-lg font-semibold">
              {editingId ? t("form.editTitle") : t("form.createTitle")}
            </h2>
            <button
              type="button"
              onClick={closeForm}
              className="text-sm text-[var(--muted)] transition-colors hover:text-[var(--foreground)]"
            >
              {t("form.cancel")}
            </button>
          </div>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field label={t("fields.name")} id="name" required>
              <input
                id="name"
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                className={inputClass}
                placeholder={t("placeholders.name")}
                required
              />
            </Field>
            <Field label={t("fields.industry")} id="industry" required>
              <input
                id="industry"
                value={form.industry}
                onChange={(e) => setForm((f) => ({ ...f, industry: e.target.value }))}
                className={inputClass}
                placeholder={t("placeholders.industry")}
                required
              />
            </Field>
          </div>

          <Field label={t("fields.targetAudience")} id="target_audience" required>
            <textarea
              id="target_audience"
              rows={3}
              value={form.target_audience}
              onChange={(e) =>
                setForm((f) => ({ ...f, target_audience: e.target.value }))
              }
              className={inputClass}
              placeholder={t("placeholders.targetAudience")}
              required
            />
          </Field>

          <div className="grid gap-4 sm:grid-cols-2">
            <Field label={t("fields.budgetRangeOptional")} id="budget_range">
              <input
                id="budget_range"
                value={form.budget_range}
                onChange={(e) =>
                  setForm((f) => ({ ...f, budget_range: e.target.value }))
                }
                className={inputClass}
                placeholder={t("placeholders.budgetRange")}
              />
            </Field>
            <Field label={t("fields.campaignGoal")} id="campaign_goal" required>
              <input
                id="campaign_goal"
                value={form.campaign_goal}
                onChange={(e) =>
                  setForm((f) => ({ ...f, campaign_goal: e.target.value }))
                }
                className={inputClass}
                placeholder={t("placeholders.campaignGoal")}
                required
              />
            </Field>
          </div>

          <Field label={t("fields.brandValues")} id="brand_values" required>
            <textarea
              id="brand_values"
              rows={3}
              value={form.brand_values}
              onChange={(e) =>
                setForm((f) => ({ ...f, brand_values: e.target.value }))
              }
              className={inputClass}
              placeholder={t("placeholders.brandValues")}
              required
            />
          </Field>

          <div className="flex justify-end gap-3 pt-2">
            <button
              type="button"
              onClick={closeForm}
              className="rounded-lg border border-[var(--border)] px-4 py-2.5 text-sm font-medium text-[var(--muted)] transition-colors hover:text-[var(--foreground)]"
            >
              {t("form.cancel")}
            </button>
            <button
              type="submit"
              disabled={saving}
              className="rounded-lg bg-[var(--accent)] px-4 py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {saving
                ? t("form.saving")
                : editingId
                  ? t("form.saveChanges")
                  : t("form.createProfile")}
            </button>
          </div>
        </form>
      )}

      {profiles.length === 0 ? (
        <div className="rounded-xl border border-dashed border-[var(--border)] bg-[var(--surface)] px-6 py-16 text-center">
          <p className="text-lg font-medium">{t("empty.title")}</p>
          <p className="mt-2 text-sm text-[var(--muted)]">{t("empty.description")}</p>
          {!showForm && (
            <button
              type="button"
              onClick={openCreateForm}
              className="mt-6 inline-flex items-center gap-2 rounded-lg border border-[var(--accent)]/40 bg-[var(--accent-muted)] px-4 py-2 text-sm font-medium text-[var(--accent)] transition-colors hover:border-[var(--accent)]/60"
            >
              {t("createFirstProfile")}
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {profiles.map((profile) => (
            <article
              key={profile.id}
              className="flex flex-col rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5"
            >
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <h3 className="text-lg font-semibold">{profile.name}</h3>
                  <p className="mt-1 text-sm text-[var(--accent)]">{profile.industry}</p>
                </div>
                <div className="flex shrink-0 gap-2">
                  <button
                    type="button"
                    onClick={() => openEditForm(profile)}
                    className="rounded-lg border border-[var(--border)] px-3 py-1.5 text-xs font-medium text-[var(--muted)] transition-colors hover:border-[var(--accent)]/40 hover:text-[var(--foreground)]"
                  >
                    {t("card.edit")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleDelete(profile)}
                    disabled={deletingId === profile.id}
                    className="rounded-lg border border-red-500/30 px-3 py-1.5 text-xs font-medium text-red-400 transition-colors hover:border-red-500/50 disabled:opacity-50"
                  >
                    {deletingId === profile.id ? t("card.deleting") : t("card.delete")}
                  </button>
                </div>
              </div>

              <dl className="mt-4 space-y-3 text-sm">
                <ProfileDetail
                  label={t("card.targetAudience")}
                  value={profile.target_audience}
                />
                {profile.budget_range && (
                  <ProfileDetail label={t("card.budget")} value={profile.budget_range} />
                )}
                <ProfileDetail
                  label={t("card.campaignGoal")}
                  value={profile.campaign_goal}
                />
                <ProfileDetail
                  label={t("card.brandValues")}
                  value={profile.brand_values}
                />
              </dl>
            </article>
          ))}
        </div>
      )}
    </div>
  );
}

const inputClass =
  "w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none focus:border-[var(--accent)]/60";

function Field({
  label,
  id,
  required,
  children,
}: {
  label: string;
  id: string;
  required?: boolean;
  children: ReactNode;
}) {
  return (
    <div>
      <label htmlFor={id} className="mb-1.5 block text-sm font-medium">
        {label}
        {required && <span className="text-[var(--muted)]"> *</span>}
      </label>
      {children}
    </div>
  );
}

function ProfileDetail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-wider text-[var(--muted)]">
        {label}
      </dt>
      <dd className="mt-1 leading-relaxed text-[var(--foreground)]/90">{value}</dd>
    </div>
  );
}
