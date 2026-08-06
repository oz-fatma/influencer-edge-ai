"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Link, usePathname, useRouter } from "@/i18n/navigation";
import LanguageToggle from "@/components/LanguageToggle";
import { authApi, isUnauthorized } from "@/lib/api";
import { clearAuth, getIsAdmin, setIsAdmin } from "@/lib/auth";

const baseNavItems = [
  { href: "/dashboard", labelKey: "dashboard" as const },
  { href: "/influencers", labelKey: "influencerPool" as const },
  { href: "/brand-profiles", labelKey: "brandProfiles" as const },
  { href: "/matching", labelKey: "matchingPanel" as const },
];

const adminNavItems = [
  { href: "/monitoring", labelKey: "monitoring" as const },
  { href: "/admin", labelKey: "admin" as const },
];

export default function Navbar() {
  const t = useTranslations("nav");
  const pathname = usePathname();
  const router = useRouter();
  const [loggingOut, setLoggingOut] = useState(false);
  const [isAdmin, setIsAdminState] = useState(false);

  useEffect(() => {
    let cancelled = false;

    setIsAdminState(getIsAdmin());
    authApi
      .me()
      .then((me) => {
        if (cancelled) return;
        setIsAdmin(me.is_admin);
        setIsAdminState((prev) => (prev === me.is_admin ? prev : me.is_admin));
      })
      .catch((err) => {
        if (cancelled || isUnauthorized(err)) return;
      });

    return () => {
      cancelled = true;
    };
  }, []);

  async function handleLogout() {
    setLoggingOut(true);
    try {
      clearAuth();
      router.push("/login");
    } finally {
      setLoggingOut(false);
    }
  }

  const items = isAdmin
    ? [...baseNavItems, ...adminNavItems]
    : baseNavItems;

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--surface)]/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <div className="flex items-center gap-10">
          <Link href="/dashboard" className="flex items-center gap-2.5">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--accent)] text-sm font-bold text-[var(--accent-fg)]">
              IE
            </span>
            <span className="text-sm font-semibold tracking-tight text-[var(--foreground)]">
              {t("appName")}
            </span>
          </Link>

          <nav className="hidden items-center gap-1 sm:flex">
            {items.map((item) => {
              const active = pathname === item.href;
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={`rounded-lg px-3.5 py-2 text-sm font-medium transition-colors ${
                    active
                      ? "bg-[var(--accent-muted)] text-[var(--accent)]"
                      : "text-[var(--muted)] hover:bg-[var(--surface-elevated)] hover:text-[var(--foreground)]"
                  }`}
                >
                  {t(item.labelKey)}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <LanguageToggle />
          <button
            onClick={handleLogout}
            disabled={loggingOut}
            className="rounded-lg border border-[var(--border)] px-4 py-2 text-sm font-medium text-[var(--muted)] transition-colors hover:border-[var(--accent)]/40 hover:text-[var(--foreground)] disabled:opacity-50"
          >
            {loggingOut ? t("loggingOut") : t("logout")}
          </button>
        </div>
      </div>
    </header>
  );
}
