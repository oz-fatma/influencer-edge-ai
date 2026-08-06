"use client";

import { FormEvent, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import LanguageToggle from "@/components/LanguageToggle";
import { clearAuth, getToken, setAuth, syncAuthCookie } from "@/lib/auth";
import { ApiError, authApi, toAuthUser } from "@/lib/api";

type Mode = "login" | "register";

function safeRedirect(path: string | null): string {
  if (!path || !path.startsWith("/") || path.startsWith("//")) {
    return "/dashboard";
  }
  return path;
}

export default function LoginForm() {
  const t = useTranslations("auth");
  const searchParams = useSearchParams();
  const redirect = safeRedirect(searchParams.get("redirect"));

  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const restoreAttemptedRef = useRef(false);

  useEffect(() => {
    if (restoreAttemptedRef.current) return;
    restoreAttemptedRef.current = true;

    const token = getToken();
    if (!token) return;

    if (!syncAuthCookie()) {
      clearAuth();
      setError(t("errors.sessionCookieRestore"));
      return;
    }

    window.location.replace(redirect);
  }, [redirect]);

  function validate(): string | null {
    if (!email.trim()) return t("validation.emailRequired");
    if (!password) return t("validation.passwordRequired");
    if (mode === "register" && !firstName.trim()) {
      return t("validation.firstNameRequired");
    }
    if (mode === "register" && !lastName.trim()) {
      return t("validation.lastNameRequired");
    }
    if (password.length < 8) return t("validation.passwordMinLength");
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

    setLoading(true);

    try {
      const payload = {
        email: email.trim().toLowerCase(),
        password,
      };

      if (mode === "register") {
        await authApi.register({
          ...payload,
          first_name: firstName.trim(),
          last_name: lastName.trim(),
        });
      }

      const response = await authApi.login(payload);
      setAuth(response.token, toAuthUser(response.user), response.is_admin === true);
      if (!syncAuthCookie()) {
        setError(t("errors.sessionCookieAfterLogin"));
        clearAuth();
        return;
      }
      window.location.replace(redirect);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError(t("errors.connectionError"));
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="relative flex min-h-full flex-col items-center justify-center bg-grid px-6 py-12">
      <div className="fixed right-6 top-6 z-50">
        <LanguageToggle />
      </div>
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-[var(--accent)] text-lg font-bold text-[var(--accent-fg)]">
            IE
          </div>
          <h1 className="text-2xl font-bold tracking-tight">{t("title")}</h1>
          <p className="mt-2 text-sm text-[var(--muted)]">{t("subtitle")}</p>
        </div>

        <div className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-8 shadow-2xl shadow-black/20">
          <div className="mb-6 flex rounded-lg bg-[var(--surface-elevated)] p-1">
            {(["login", "register"] as Mode[]).map((m) => (
              <button
                key={m}
                type="button"
                onClick={() => {
                  setMode(m);
                  setError(null);
                }}
                className={`flex-1 rounded-md py-2 text-sm font-medium transition-colors ${
                  mode === m
                    ? "bg-[var(--accent-muted)] text-[var(--accent)]"
                    : "text-[var(--muted)] hover:text-[var(--foreground)]"
                }`}
              >
                {m === "login" ? t("tabs.signIn") : t("tabs.register")}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {mode === "register" && (
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label htmlFor="first_name" className="mb-1.5 block text-sm font-medium">
                    {t("fields.firstName")}
                  </label>
                  <input
                    id="first_name"
                    type="text"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                    placeholder={t("placeholders.firstName")}
                  />
                </div>
                <div>
                  <label htmlFor="last_name" className="mb-1.5 block text-sm font-medium">
                    {t("fields.lastName")}
                  </label>
                  <input
                    id="last_name"
                    type="text"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                    placeholder={t("placeholders.lastName")}
                  />
                </div>
              </div>
            )}

            <div>
              <label htmlFor="email" className="mb-1.5 block text-sm font-medium">
                {t("fields.email")}
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                placeholder={t("placeholders.email")}
                autoComplete="email"
              />
            </div>

            <div>
              <label htmlFor="password" className="mb-1.5 block text-sm font-medium">
                {t("fields.password")}
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                placeholder={t("placeholders.password")}
                autoComplete={mode === "login" ? "current-password" : "new-password"}
              />
            </div>

            {error && (
              <p className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-400">
                {error}
              </p>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full rounded-lg bg-[var(--accent)] py-2.5 text-sm font-semibold text-[var(--accent-fg)] transition-opacity hover:opacity-90 disabled:opacity-50"
            >
              {loading
                ? t("submit.processing")
                : mode === "login"
                  ? t("submit.signIn")
                  : t("submit.createAccount")}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
