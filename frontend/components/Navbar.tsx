"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { clearAuth } from "@/lib/auth";

const navItems = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/influencers", label: "Influencer Pool" },
  { href: "/matching", label: "Matching Panel" },
  { href: "/monitoring", label: "Monitoring" },
];

export default function Navbar() {
  const pathname = usePathname();
  const router = useRouter();
  const [loggingOut, setLoggingOut] = useState(false);

  async function handleLogout() {
    setLoggingOut(true);
    try {
      clearAuth();
      router.push("/login");
    } finally {
      setLoggingOut(false);
    }
  }

  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--surface)]/80 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
        <div className="flex items-center gap-10">
          <Link href="/dashboard" className="flex items-center gap-2.5">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-[var(--accent)] text-sm font-bold text-[var(--accent-fg)]">
              IE
            </span>
            <span className="text-sm font-semibold tracking-tight text-[var(--foreground)]">
              InfluencerEdge
            </span>
          </Link>

          <nav className="hidden items-center gap-1 sm:flex">
            {navItems.map((item) => {
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
                  {item.label}
                </Link>
              );
            })}
          </nav>
        </div>

        <button
          onClick={handleLogout}
          disabled={loggingOut}
          className="rounded-lg border border-[var(--border)] px-4 py-2 text-sm font-medium text-[var(--muted)] transition-colors hover:border-[var(--accent)]/40 hover:text-[var(--foreground)] disabled:opacity-50"
        >
          {loggingOut ? "Logging out..." : "Logout"}
        </button>
      </div>
    </header>
  );
}
