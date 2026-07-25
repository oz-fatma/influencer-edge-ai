import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const PROTECTED_PREFIXES = [
  "/dashboard",
  "/influencers",
  "/matching",
  "/monitoring",
  "/admin",
];

function isProtected(pathname: string) {
  return PROTECTED_PREFIXES.some(
    (prefix) => pathname === prefix || pathname.startsWith(`${prefix}/`),
  );
}

function hasAuthCookie(request: NextRequest): boolean {
  if (request.cookies.get("session")?.value === "1") {
    return true;
  }
  const legacyToken = request.cookies.get("token")?.value?.trim();
  return Boolean(legacyToken);
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (isProtected(pathname) && !hasAuthCookie(request)) {
    const loginUrl = new URL("/login", request.url);
    loginUrl.searchParams.set("redirect", pathname);
    return NextResponse.redirect(loginUrl);
  }

  if (pathname === "/login" && hasAuthCookie(request)) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard/:path*",
    "/influencers/:path*",
    "/matching/:path*",
    "/monitoring/:path*",
    "/admin/:path*",
    "/login",
  ],
};
