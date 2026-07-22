const TOKEN_KEY = "token";
const USER_KEY = "user";

export type AuthUser = {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
};

function setCookie(name: string, value: string, maxAgeSeconds: number) {
  document.cookie = `${name}=${value}; path=/; max-age=${maxAgeSeconds}; SameSite=Lax`;
}

function clearCookie(name: string) {
  document.cookie = `${name}=; path=/; max-age=0`;
}

export function getUserDisplayName(user: AuthUser | null): string {
  if (!user) return "";
  return `${user.first_name} ${user.last_name}`.trim();
}

export function setAuth(token: string, user: AuthUser) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
  // Clear legacy keys from the old backend contract
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
  clearCookie("access_token");
  clearCookie("refresh_token");

  setCookie(TOKEN_KEY, token, 60 * 60 * 24);
}

export function clearAuth() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
  localStorage.removeItem("access_token");
  localStorage.removeItem("refresh_token");
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
