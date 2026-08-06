"use client";

import { useLocale } from "next-intl";
import { usePathname, useRouter } from "@/i18n/navigation";
import type { Locale } from "@/i18n/routing";

const locales: Locale[] = ["en", "tr"];

export default function LanguageToggle() {
  const locale = useLocale();
  const router = useRouter();
  const pathname = usePathname();

  function switchLocale(nextLocale: Locale) {
    if (nextLocale === locale) return;
    router.replace(pathname, { locale: nextLocale });
    router.refresh();
  }

  return (
    <div
      className="flex items-center gap-1.5 text-sm"
      role="group"
      aria-label="Language"
    >
      {locales.map((code, index) => (
        <span key={code} className="flex items-center gap-1.5">
          {index > 0 && (
            <span className="text-[var(--muted)] opacity-40" aria-hidden="true">
              |
            </span>
          )}
          <button
            type="button"
            onClick={() => switchLocale(code)}
            aria-current={locale === code ? "true" : undefined}
            className={`uppercase transition-colors ${
              locale === code
                ? "font-semibold text-[var(--accent)]"
                : "font-medium text-[var(--muted)] hover:text-[var(--foreground)]"
            }`}
          >
            {code}
          </button>
        </span>
      ))}
    </div>
  );
}
