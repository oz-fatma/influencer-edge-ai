"use client";

import { useTranslations } from "next-intl";

export function LoginLoadingFallback() {
  const t = useTranslations("auth");

  return (
    <div className="flex min-h-full items-center justify-center text-[var(--muted)]">
      {t("loading")}
    </div>
  );
}
