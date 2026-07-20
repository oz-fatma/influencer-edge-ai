"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { getUser, type AuthUser } from "@/lib/auth";
import {
  handleUnauthorizedRedirect,
  isUnauthorized,
  scoresApi,
  type InfluencerScore,
} from "@/lib/api";

export default function DashboardPage() {
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

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">
          Hoş geldin{user?.name ? `, ${user.name}` : ""} 👋
        </h1>
        <p className="mt-1 text-[var(--muted)]">
          InfluencerEdge AI panelinize genel bakış
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        <StatCard
          label="Havuzdaki Influencer"
          value={loading ? "—" : String(scores.length)}
        />
        <StatCard
          label="Ortalama Skor"
          value={loading ? "—" : avgScore.toFixed(1)}
          suffix={loading ? undefined : "/100"}
        />
        <StatCard
          label="Yüksek Uyum"
          value={loading ? "—" : String(highFit)}
          suffix={loading ? undefined : " aday"}
        />
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="mb-4 text-lg font-semibold">Hızlı Erişim</h2>
          <div className="space-y-3">
            <QuickLink
              href="/influencers"
              title="Influencer Havuzu"
              desc="Skorlanmış influencer listesini incele"
            />
            <QuickLink
              href="/matching"
              title="AI Eşleştirme"
              desc="Web-LLM analiz sonuçlarını görüntüle"
            />
          </div>
        </section>

        <section className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <h2 className="mb-4 text-lg font-semibold">Son Skorlar</h2>
          {loading ? (
            <p className="text-sm text-[var(--muted)]">Yükleniyor...</p>
          ) : scores.length === 0 ? (
            <p className="text-sm text-[var(--muted)]">Henüz skor eklenmedi</p>
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
