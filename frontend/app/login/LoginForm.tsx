"use client";

import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { setAuth } from "@/lib/auth";
import { ApiError, authApi, toAuthUser } from "@/lib/api";

type Mode = "login" | "register";

function safeRedirect(path: string | null): string {
  if (!path || !path.startsWith("/") || path.startsWith("//")) {
    return "/dashboard";
  }
  return path;
}

export default function LoginForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const redirect = safeRedirect(searchParams.get("redirect"));

  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  function validate(): string | null {
    if (!email.trim()) return "Email is required";
    if (!password) return "Password is required";
    if (mode === "register" && !firstName.trim()) return "First name is required";
    if (mode === "register" && !lastName.trim()) return "Last name is required";
    if (password.length < 8) return "Password must be at least 8 characters";
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
      setAuth(response.token, toAuthUser(response.user));
      router.push(redirect);
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("Connection error — is the backend running?");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-full flex-col items-center justify-center bg-grid px-6 py-12">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl bg-[var(--accent)] text-lg font-bold text-[var(--accent-fg)]">
            IE
          </div>
          <h1 className="text-2xl font-bold tracking-tight">InfluencerEdge AI</h1>
          <p className="mt-2 text-sm text-[var(--muted)]">
            Sign in to the influencer–agency matching platform
          </p>
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
                {m === "login" ? "Sign In" : "Register"}
              </button>
            ))}
          </div>

          <form onSubmit={handleSubmit} className="space-y-4">
            {mode === "register" && (
              <div className="grid gap-4 sm:grid-cols-2">
                <div>
                  <label htmlFor="first_name" className="mb-1.5 block text-sm font-medium">
                    First Name
                  </label>
                  <input
                    id="first_name"
                    type="text"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                    placeholder="Jane"
                  />
                </div>
                <div>
                  <label htmlFor="last_name" className="mb-1.5 block text-sm font-medium">
                    Last Name
                  </label>
                  <input
                    id="last_name"
                    type="text"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                    placeholder="Doe"
                  />
                </div>
              </div>
            )}

            <div>
              <label htmlFor="email" className="mb-1.5 block text-sm font-medium">
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                placeholder="you@agency.com"
                autoComplete="email"
              />
            </div>

            <div>
              <label htmlFor="password" className="mb-1.5 block text-sm font-medium">
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full rounded-lg border border-[var(--border)] bg-[var(--surface-elevated)] px-4 py-2.5 text-sm outline-none transition-colors focus:border-[var(--accent)]/60 focus:ring-1 focus:ring-[var(--accent)]/30"
                placeholder="••••••••"
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
                ? "Processing..."
                : mode === "login"
                  ? "Sign In"
                  : "Create Account"}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
