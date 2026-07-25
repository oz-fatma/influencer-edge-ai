const TOKEN_KEY = "token";
const USER_KEY = "user";
const IS_ADMIN_KEY = "is_admin";
const SESSION_KEY = "session";
const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24;

export type AuthUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  is_admin?: boolean;
};

function readCookie(name: string): string | null {
  if (typeof document === "undefined") return null;
  const prefix = `${name}=`;
  for (const part of document.cookie.split(";")) {
    const trimmed = part.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

function setCookie(name: string, value: string, maxAgeSeconds: number) {
  document.cookie = `${name}=${encodeURIComponent(value)}; path=/; max-age=${maxAgeSeconds}; SameSite=Lax`;
}

function clearCookie(name: string) {
  document.cookie = `${name}=; path=/; max-age=0; expires=Thu, 01 Jan 1970 00:00:00 GMT; SameSite=Lax`;
}

/** Lightweight marker for Next.js middleware (JWT stays in localStorage). */
export function syncAuthCookie(): boolean {
  if (typeof window === "undefined") return false;
  const token = localStorage.getItem(TOKEN_KEY);
  if (!token) {
    clearCookie(SESSION_KEY);
    clearCookie(TOKEN_KEY);
    return false;
  }
  setCookie(SESSION_KEY, "1", COOKIE_MAX_AGE_SECONDS);
  return readCookie(SESSION_KEY) === "1";
}

export function getUserDisplayName(user: AuthUser | null): string {
  if (!user) return "";
  return `${user.first_name} ${user.last_name}`.trim();
}

export function setAuth(token: string, user: AuthUser, isAdmin = false) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify({ ...user, is_admin: isAdmin }));
  localStorage.setItem(IS_ADMIN_KEY, isAdmin ? "1" : "0");
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  clearCookie("access_token");
  clearCookie("refresh_token");

  syncAuthCookie();
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem(IS_ADMIN_KEY);
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  clearCookie(SESSION_KEY);
  clearCookie(TOKEN_KEY);
  clearCookie("access_token");
  clearCookie("refresh_token");
}

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}

export function getUser(): AuthUser | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as AuthUser;
  } catch {
    return null;
  }
}

export function getIsAdmin(): boolean {
  if (typeof window === "undefined") return false;
  if (localStorage.getItem(IS_ADMIN_KEY) === "1") return true;
  return getUser()?.is_admin === true;
}

export function setIsAdmin(isAdmin: boolean) {
  if (typeof window === "undefined") return;
  localStorage.setItem(IS_ADMIN_KEY, isAdmin ? "1" : "0");
  const user = getUser();
  if (user) {
    localStorage.setItem(USER_KEY, JSON.stringify({ ...user, is_admin: isAdmin }));
  }
}
