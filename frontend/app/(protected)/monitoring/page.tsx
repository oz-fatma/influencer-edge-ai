"use client";

import { useCallback, useEffect, useState } from "react";
import {
  handleUnauthorizedRedirect,
  isUnauthorized,
  monitoringApi,
  type MonitoringStats,
} from "@/lib/api";
import { formatLatency, formatTimestamp } from "@/lib/format";

const POLL_INTERVAL_MS = 10_000;

export default function MonitoringPage() {
  const [stats, setStats] = useState<MonitoringStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const loadStats = useCallback(async (isInitial = false) => {
    try {
      const data = await monitoringApi.getStats();
      setStats(data);
      setError(null);
      setLastUpdated(new Date());
    } catch (err) {
      if (isUnauthorized(err)) {
        handleUnauthorizedRedirect("/monitoring");
        return;
      }
      if (isInitial) {
        setError("Monitoring verileri yüklenemedi.");
      }
    } finally {
      if (isInitial) {
        setLoading(false);
      }
    }
  }, []);

  useEffect(() => {
    loadStats(true);
    const interval = setInterval(() => loadStats(false), POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [loadStats]);

  if (loading) {
    return (
      <div className="flex items-center justify-center py-24 text-[var(--muted)]">
        Yükleniyor...
      </div>
    );
  }

  if (error || !stats) {
    return (
      <div className="rounded-xl border border-red-500/20 bg-red-500/10 px-6 py-8 text-center text-red-400">
        {error ?? "Veri alınamadı"}
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">LLM Monitoring</h1>
          <p className="mt-1 text-[var(--muted)]">
            Web-LLM çağrı metrikleri ve performans özeti
          </p>
        </div>
        {lastUpdated && (
          <p className="text-xs text-[var(--muted)]">
            Son güncelleme:{" "}
            {lastUpdated.toLocaleTimeString("tr-TR", {
              hour: "2-digit",
              minute: "2-digit",
              second: "2-digit",
            })}
            <span className="ml-2 inline-flex items-center gap-1">
              <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-[var(--success)]" />
              10s
            </span>
          </p>
        )}
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard
          label="Toplam Çağrı"
          value={String(stats.total_calls)}
        />
        <StatCard
          label="Ortalama Yanıt Süresi"
          value={formatLatency(stats.avg_latency_ms)}
        />
        <StatCard
          label="Hata Oranı"
          value={`${stats.error_rate.toFixed(1)}%`}
          highlight={stats.error_rate > 0}
        />
      </div>

      <section className="overflow-hidden rounded-xl border border-[var(--border)] bg-[var(--surface)]">
        <div className="border-b border-[var(--border)] px-5 py-4">
          <h2 className="text-lg font-semibold">Son Çağrılar</h2>
          <p className="text-sm text-[var(--muted)]">
            En son {stats.recent_calls.length} LLM çağrısı
          </p>
        </div>

        {stats.recent_calls.length === 0 ? (
          <div className="px-6 py-12 text-center text-[var(--muted)]">
            Henüz LLM çağrısı kaydedilmedi
          </div>
        ) : (
          <>
            <div className="hidden md:block">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr className="border-b border-[var(--border)] text-[var(--muted)]">
                    <th className="px-5 py-3 font-medium">Influencer</th>
                    <th className="px-5 py-3 font-medium">Model</th>
                    <th className="px-5 py-3 font-medium">Süre</th>
                    <th className="px-5 py-3 font-medium">Durum</th>
                    <th className="px-5 py-3 font-medium">Zaman</th>
                  </tr>
                </thead>
                <tbody>
                  {stats.recent_calls.map((call) => (
                    <tr
                      key={call.id}
                      className="border-b border-[var(--border)] last:border-0 transition-colors hover:bg-[var(--surface-elevated)]"
                    >
                      <td className="px-5 py-4 font-medium">{call.influencer_name}</td>
                      <td className="px-5 py-4 text-[var(--muted)]">{call.model}</td>
                      <td className="px-5 py-4 font-mono text-xs">
                        {formatLatency(call.latency_ms)}
                      </td>
                      <td className="px-5 py-4">
                        <StatusBadge status={call.status} />
                      </td>
                      <td className="px-5 py-4 text-[var(--muted)]">
                        {formatTimestamp(call.timestamp)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="grid gap-3 p-4 md:hidden">
              {stats.recent_calls.map((call) => (
                <article
                  key={call.id}
                  className="rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] p-4"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div>
                      <p className="font-medium">{call.influencer_name}</p>
                      <p className="text-xs text-[var(--muted)]">{call.model}</p>
                    </div>
                    <StatusBadge status={call.status} />
                  </div>
                  <div className="mt-3 flex items-center justify-between text-xs text-[var(--muted)]">
                    <span className="font-mono">{formatLatency(call.latency_ms)}</span>
                    <span>{formatTimestamp(call.timestamp)}</span>
                  </div>
                </article>
              ))}
            </div>
          </>
        )}
      </section>
    </div>
  );
}

function StatCard({
  label,
  value,
  highlight,
}: {
  label: string;
  value: string;
  highlight?: boolean;
}) {
  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5">
      <p className="text-sm text-[var(--muted)]">{label}</p>
      <p
        className={`mt-1 text-3xl font-bold tracking-tight ${
          highlight ? "text-[var(--warning)]" : ""
        }`}
      >
        {value}
      </p>
    </div>
  );
}

function StatusBadge({ status }: { status: string }) {
  const isSuccess = status === "success";
  return (
    <span
      className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${
        isSuccess
          ? "bg-[var(--success)]/15 text-[var(--success)]"
          : "bg-red-500/15 text-red-400"
      }`}
    >
      {status}
    </span>
  );
}
