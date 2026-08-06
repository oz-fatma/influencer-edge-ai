"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Link } from "@/i18n/navigation";
import { getUser, getUserDisplayName, type AuthUser } from "@/lib/auth";
import {
  handleUnauthorizedRedirect,
  isUnauthorized,
  scoresApi,
  type InfluencerScore,
} from "@/lib/api";

export default function DashboardPage() {
  const t = useTranslations("dashboard");
  const [user, setUser] = useState<AuthUser | null>(null);
  const [scores, setScores] = useState<InfluencerScore[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setUser(getUser());

    async function load() {
      try {
        const data = await scoresApi.list();
        setScores(data.scores ?? []);
      } catch (err) {
        if (isUnauthorized(err)) {
          handleUnauthorizedRedirect("/dashboard");
        }
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const avgScore =
    scores.length > 0
      ? scores.reduce((sum, s) => sum + s.overall_score, 0) / scores.length
      : 0;

  const highFit = scores.filter((s) => s.overall_score >= 85).length;

  const displayName = getUserDisplayName(user);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          {displayName ? t("welcomeWithName", { name: displayName }) : t("welcome")}{" "}
          👋
        </h1>
        <p className="mt-1 text-[var(--muted)]">{t("subtitle")}</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard
          label={t("stats.influencersInPool")}
          value={loading ? t("stats.emptyValue") : String(scores.length)}
        />
        <StatCard
          label={t("stats.averageScore")}
          value={loading ? t("stats.emptyValue") : avgScore.toFixed(1)}
          suffix={loading ? undefined : t("stats.scoreSuffix")}
        />
        <StatCard
          label={t("stats.highFit")}
          value={loading ? t("stats.emptyValue") : String(highFit)}
          suffix={loading ? undefined : t("stats.candidatesSuffix")}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="mb-4 text-lg font-semibold">{t("quickAccess.title")}</h2>
          <div className="space-y-3">
            <QuickLink
              href="/influencers"
              title={t("quickAccess.influencerPoolTitle")}
              desc={t("quickAccess.influencerPoolDesc")}
            />
            <QuickLink
              href="/matching"
              title={t("quickAccess.aiMatchingTitle")}
              desc={t("quickAccess.aiMatchingDesc")}
            />
          </div>
        </section>

        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="mb-4 text-lg font-semibold">{t("recentScores.title")}</h2>
          {loading ? (
            <p className="text-sm text-[var(--muted)]">{t("recentScores.loading")}</p>
          ) : scores.length === 0 ? (
            <p className="text-sm text-[var(--muted)]">{t("recentScores.empty")}</p>
          ) : (
            <ul className="space-y-3">
              {scores.slice(0, 3).map((s) => (
                <li
                  key={s.id}
                  className="flex items-center justify-between rounded-lg bg-[var(--surface-elevated)] px-4 py-3"
                >
                  <div>
                    <p className="text-sm font-medium">{s.influencer_name}</p>
                    <p className="text-xs capitalize text-[var(--muted)]">{s.platform}</p>
                  </div>
                  <span className="text-sm font-semibold text-[var(--accent)]">
                    {s.overall_score}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}

function StatCard({
  label,
  value,
  suffix,
}: {
  label: string;
  value: string;
  suffix?: string;
}) {
  return (
    <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5">
      <p className="text-sm text-[var(--muted)]">{label}</p>
      <p className="mt-1 text-3xl font-bold tracking-tight">
        {value}
        {suffix && (
          <span className="text-base font-normal text-[var(--muted)]">{suffix}</span>
        )}
      </p>
    </div>
  );
}

function QuickLink({
  href,
  title,
  desc,
}: {
  href: string;
  title: string;
  desc: string;
}) {
  return (
    <Link
      href={href}
      className="flex items-center justify-between rounded-lg border border-[var(--border)] px-4 py-3 transition-colors hover:border-[var(--accent)]/40 hover:bg-[var(--surface-elevated)]"
    >
      <div>
        <p className="text-sm font-medium">{title}</p>
        <p className="text-xs text-[var(--muted)]">{desc}</p>
      </div>
      <span className="text-[var(--accent)]">→</span>
    </Link>
  );
}
