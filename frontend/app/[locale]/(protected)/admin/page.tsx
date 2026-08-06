"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useLocale, useTranslations } from "next-intl";
import {
  adminApi,
  authApi,
  handleUnauthorizedRedirect,
  isUnauthorized,
  type LLMConfig,
  type LLMLogEntry,
} from "@/lib/api";
import { setIsAdmin } from "@/lib/auth";
import { formatLatency } from "@/lib/format";

export default function AdminPage() {
  const router = useRouter();
  const locale = useLocale();
  const t = useTranslations("admin");
  const intlLocale = locale === "tr" ? "tr-TR" : "en-US";

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [models, setModels] = useState<string[]>([]);
  const [logs, setLogs] = useState<LLMLogEntry[]>([]);

  const [systemPrompt, setSystemPrompt] = useState("");
  const [temperature, setTemperature] = useState(0.1);
  const [maxTokens, setMaxTokens] = useState(100);
  const [model, setModel] = useState("gemma-influencer-ft");

  const routerRef = useRef(router);
  routerRef.current = router;

  function applyConfig(config: LLMConfig) {
    setSystemPrompt(config.system_prompt);
    setTemperature(config.temperature);
    setMaxTokens(config.max_tokens);
    setModel(config.model);
  }

  useEffect(() => {
    let cancelled = false;

    async function loadData() {
      try {
        const me = await authApi.me();
        if (cancelled) return;

        setIsAdmin(me.is_admin);
        if (!me.is_admin) {
          routerRef.current.replace("/dashboard");
          return;
        }

        const [config, modelsRes, logsRes] = await Promise.all([
          adminApi.getLLMConfig(),
          adminApi.getAllowedModels(),
          adminApi.getLLMLogs(50),
        ]);
        if (cancelled) return;

        applyConfig(config);
        setModels(modelsRes.models);
        setLogs(logsRes.logs);
        setError(null);
      } catch (err) {
        if (cancelled) return;
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/admin");
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
  }, [t]);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSuccess(null);
    setError(null);

    try {
      const updated = await adminApi.updateLLMConfig({
        system_prompt: systemPrompt.trim(),
        temperature,
        max_tokens: maxTokens,
        model,
      });
      applyConfig(updated);
      setSuccess(t("success.configSaved"));
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/admin");
        return;
      }
      setError(err instanceof Error ? err.message : t("errors.saveFailed"));
    } finally {
      setSaving(false);
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
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">{t("title")}</h1>
        <p className="mt-1 text-[var(--muted)]">{t("subtitle")}</p>
      </div>

      {error && (
        <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-4 py-3 text-sm text-red-400">
          {error}
        </div>
      )}
      {success && (
        <div className="rounded-xl border border-emerald-500/20 bg-emerald-500/10 px-4 py-3 text-sm text-emerald-400">
          {success}
        </div>
      )}

      <form
        onSubmit={handleSubmit}
        className="space-y-5 rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-6"
      >
        <div>
          <label htmlFor="system_prompt" className="mb-1.5 block text-sm font-medium">
            {t("form.systemPrompt")}
          </label>
          <textarea
            id="system_prompt"
            rows={5}
            value={systemPrompt}
            onChange={(e) => setSystemPrompt(e.target.value)}
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-3 text-sm outline-none focus:border-[var(--accent)]/60"
            required
          />
        </div>

        <div className="grid gap-4 sm:grid-cols-3">
          <div>
            <label htmlFor="temperature" className="mb-1.5 block text-sm font-medium">
              {t("form.temperature")}
            </label>
            <input
              id="temperature"
              type="number"
              min={0}
              max={1}
              step={0.05}
              value={temperature}
              onChange={(e) => setTemperature(Number(e.target.value))}
              className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none focus:border-[var(--accent)]/60"
              required
            />
          </div>
          <div>
            <label htmlFor="max_tokens" className="mb-1.5 block text-sm font-medium">
              {t("form.maxTokens")}
            </label>
            <input
              id="max_tokens"
              type="number"
              min={1}
              max={2000}
              value={maxTokens}
              onChange={(e) => setMaxTokens(Number(e.target.value))}
              className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none focus:border-[var(--accent)]/60"
              required
            />
          </div>
          <div>
            <label htmlFor="model" className="mb-1.5 block text-sm font-medium">
              {t("form.model")}
            </label>
            <select
              id="model"
              value={model}
              onChange={(e) => setModel(e.target.value)}
              className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none focus:border-[var(--accent)]/60"
            >
              {(models.length > 0 ? models : ["gemma-influencer-ft", "gemma2:2b"]).map((m) => (
                <option key={m} value={m}>
                  {m}
                </option>
              ))}
            </select>
          </div>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="rounded-lg bg-[var(--accent)] px-5 py-2.5 text-sm font-semibold text-[var(--accent-fg)] hover:opacity-90 disabled:opacity-50"
        >
          {saving ? t("form.saving") : t("form.save")}
        </button>
      </form>

      <section className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-6">
        <h2 className="text-lg font-semibold">{t("logs.title")}</h2>
        <p className="mt-1 text-sm text-[var(--muted)]">{t("logs.subtitle")}</p>

        <div className="mt-4 overflow-x-auto">
          <table className="w-full min-w-[640px] text-left text-sm">
            <thead>
              <tr className="border-b border-[var(--border)] text-[var(--muted)]">
                <th className="px-3 py-2 font-medium">{t("table.time")}</th>
                <th className="px-3 py-2 font-medium">{t("table.model")}</th>
                <th className="px-3 py-2 font-medium">{t("table.duration")}</th>
                <th className="px-3 py-2 font-medium">{t("table.status")}</th>
              </tr>
            </thead>
            <tbody>
              {logs.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-3 py-6 text-center text-[var(--muted)]">
                    {t("logs.empty")}
                  </td>
                </tr>
              ) : (
                logs.map((log, index) => (
                  <tr key={`${log.created_at}-${index}`} className="border-b border-[var(--border)]/60">
                    <td className="px-3 py-2">
                      {new Date(log.created_at).toLocaleString(intlLocale)}
                    </td>
                    <td className="px-3 py-2 font-mono text-xs">{log.model_name}</td>
                    <td className="px-3 py-2">{formatLatency(log.duration_ms)}</td>
                    <td className="px-3 py-2">
                      <span
                        className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                          log.success
                            ? "bg-emerald-500/15 text-emerald-400"
                            : "bg-red-500/15 text-red-400"
                        }`}
                      >
                        {log.success ? t("status.success") : t("status.error")}
                      </span>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </section>
    </div>
  );
}
